package controller

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const triggerPathPrefix = "/v1/triggers/"

// AgentTriggerReconciler verifies that an event trigger's bearer-token source
// is readable before the receiver advertises it as ready.
type AgentTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agenttriggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agenttriggers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *AgentTriggerReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	var trigger agentv1alpha1.AgentTrigger
	if err := r.Get(ctx, request.NamespacedName, &trigger); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "TokenAvailable",
		Message:            "bearer token source is available",
		ObservedGeneration: trigger.Generation,
	}
	if _, err := readTriggerToken(ctx, r.Client, &trigger); err != nil {
		condition.Status = metav1.ConditionFalse
		condition.Reason = "TokenUnavailable"
		condition.Message = err.Error()
	}
	apimeta.SetStatusCondition(&trigger.Status.Conditions, condition)
	return ctrl.Result{}, r.Status().Update(ctx, &trigger)
}

func (r *AgentTriggerReconciler) SetupWithManager(manager ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(manager).
		For(&agentv1alpha1.AgentTrigger{}).
		Owns(&agentv1alpha1.AgentRun{}).
		Complete(r)
}

// AgentTriggerReceiver is the authenticated event ingress port. It resolves a
// declared AgentTrigger and Secret through Kubernetes rather than retaining
// tokens in process memory between requests.
type AgentTriggerReceiver struct {
	Client  client.Client
	Scheme  *runtime.Scheme
	Address string
	Now     func() time.Time
}

func (r *AgentTriggerReceiver) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeTriggerResponse(request.Context(), response, http.StatusMethodNotAllowed, map[string]string{"failureCode": "trigger-method-not-allowed"})
		return
	}
	namespace, endpoint, matched := parseTriggerPath(request.URL.Path)
	if !matched {
		writeTriggerResponse(request.Context(), response, http.StatusNotFound, map[string]string{"failureCode": "trigger-not-found"})
		return
	}

	trigger, err := r.findTrigger(request.Context(), namespace, endpoint)
	if err != nil {
		status := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeTriggerResponse(request.Context(), response, status, map[string]string{"failureCode": "trigger-not-found"})
		return
	}
	token, err := readTriggerToken(request.Context(), r.Client, trigger)
	if err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusServiceUnavailable, map[string]string{"failureCode": "trigger-token-unavailable"})
		return
	}
	provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), token) != 1 {
		writeTriggerResponse(request.Context(), response, http.StatusUnauthorized, map[string]string{"failureCode": "trigger-unauthorized"})
		return
	}

	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	idempotencyKey := request.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = fmt.Sprintf("%d", now.UnixNano())
	}
	run := buildTriggeredRun(trigger.Spec.Template, trigger.Namespace, trigger.Name, "AgentTrigger", idempotencyKey)
	if err := ctrl.SetControllerReference(trigger, run, r.Scheme); err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusInternalServerError, map[string]string{"failureCode": "trigger-owner-reference-failed"})
		return
	}
	if err := r.Client.Create(request.Context(), run); err != nil && !apierrors.IsAlreadyExists(err) {
		writeTriggerResponse(request.Context(), response, http.StatusInternalServerError, map[string]string{"failureCode": "trigger-run-create-failed"})
		return
	}

	trigger.Status.LastRunName = run.Name
	triggeredAt := metav1.NewTime(now)
	trigger.Status.LastTriggeredTime = &triggeredAt
	if err := r.Client.Status().Update(request.Context(), trigger); err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusInternalServerError, map[string]string{"failureCode": "trigger-status-update-failed"})
		return
	}
	writeTriggerResponse(request.Context(), response, http.StatusCreated, map[string]string{"run": run.Namespace + "/" + run.Name})
}

func (r *AgentTriggerReceiver) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:              r.Address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr <- server.Shutdown(shutdownContext)
	}()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if ctx.Err() != nil {
		if err := <-shutdownErr; err != nil {
			return fmt.Errorf("shut down trigger receiver: %w", err)
		}
	}
	return nil
}

func (r *AgentTriggerReceiver) NeedLeaderElection() bool {
	return false
}

func (r *AgentTriggerReceiver) findTrigger(ctx context.Context, namespace, endpoint string) (*agentv1alpha1.AgentTrigger, error) {
	var triggers agentv1alpha1.AgentTriggerList
	if err := r.Client.List(ctx, &triggers, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	for index := range triggers.Items {
		if triggers.Items[index].Spec.Endpoint == endpoint {
			return triggers.Items[index].DeepCopy(), nil
		}
	}
	return nil, apierrors.NewNotFound(agentv1alpha1.GroupVersion.WithResource("agenttriggers").GroupResource(), endpoint)
}

func parseTriggerPath(path string) (string, string, bool) {
	return parseNamespacedPath(path, triggerPathPrefix)
}

func parseNamespacedPath(path, prefix string) (string, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func readTriggerToken(ctx context.Context, kubeClient client.Client, trigger *agentv1alpha1.AgentTrigger) ([]byte, error) {
	return readSecretToken(ctx, kubeClient, trigger.Namespace, trigger.Spec.TokenSecretRef)
}

func readSecretToken(ctx context.Context, kubeClient client.Client, namespace string, reference agentv1alpha1.SecretKeyRef) ([]byte, error) {
	if reference.Name == "" || reference.Key == "" {
		return nil, fmt.Errorf("tokenSecretRef name and key are required")
	}
	var secret corev1.Secret
	if err := kubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: reference.Name}, &secret); err != nil {
		return nil, fmt.Errorf("read bearer token Secret: %w", err)
	}
	token := secret.Data[reference.Key]
	if len(token) == 0 {
		return nil, fmt.Errorf("bearer token Secret key is empty")
	}
	return token, nil
}

func buildTriggeredRun(template agentv1alpha1.AgentRunTemplate, namespace, sourceName, sourceKind, idempotencyKey string) *agentv1alpha1.AgentRun {
	digest := sha256.Sum256([]byte(sourceKind + ":" + sourceName + ":" + idempotencyKey))
	suffix := hex.EncodeToString(digest[:])[:12]
	base := strings.ToLower(sourceName)
	if len(base) > 45 {
		base = base[:45]
	}
	annotations := map[string]string{}
	for key, value := range template.Metadata.Annotations {
		annotations[key] = value
	}
	annotations[agentv1alpha1.TriggeredByKindAnnotation] = sourceKind
	annotations[agentv1alpha1.TriggeredByNameAnnotation] = sourceName

	return &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        base + "-" + suffix,
			Namespace:   namespace,
			Labels:      template.Metadata.Labels,
			Annotations: annotations,
		},
		Spec: *template.Spec.DeepCopy(),
	}
}

func writeTriggerResponse(ctx context.Context, response http.ResponseWriter, status int, payload any) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		log.FromContext(ctx).Error(err, "encode trigger response")
		http.Error(response, "failed to encode response", http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	if _, err := response.Write(append(encoded, '\n')); err != nil {
		log.FromContext(ctx).Error(err, "write trigger response")
	}
}
