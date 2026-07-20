package controller

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	governancecontracts "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/governance"
)

const (
	triggerPathPrefix    = "/v1/triggers/"
	escalationPathPrefix = "/v1/escalations/"
	blockedPathPrefix    = "/v1/blocks/"
)

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
	if strings.HasPrefix(request.URL.Path, blockedPathPrefix) {
		r.serveBlockedSignal(response, request)
		return
	}
	if strings.HasPrefix(request.URL.Path, escalationPathPrefix) {
		r.serveEscalation(response, request)
		return
	}
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

type blockedSignalRequest struct {
	SubjectRevision   string `json:"subjectRevision"`
	Reporter          string `json:"reporter"`
	ReasonCode        string `json:"reasonCode"`
	EvidenceReference string `json:"evidenceReference"`
}

func decodeStrictJSON(input io.Reader, maximumBytes int64, target any) error {
	payload, err := io.ReadAll(io.LimitReader(input, maximumBytes+1))
	if err != nil {
		return err
	}
	if int64(len(payload)) > maximumBytes {
		return fmt.Errorf("JSON input exceeds %d bytes", maximumBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func (r *AgentTriggerReceiver) serveBlockedSignal(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeTriggerResponse(request.Context(), response, http.StatusMethodNotAllowed, map[string]string{"failureCode": "blocked-signal-method-not-allowed"})
		return
	}
	namespace, name, matched := parseNamespacedPath(request.URL.Path, blockedPathPrefix)
	if !matched {
		writeTriggerResponse(request.Context(), response, http.StatusNotFound, map[string]string{"failureCode": "blocked-signal-run-not-found"})
		return
	}

	var run agentv1alpha1.AgentRun
	if err := r.Client.Get(request.Context(), types.NamespacedName{Namespace: namespace, Name: name}, &run); err != nil {
		status := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeTriggerResponse(request.Context(), response, status, map[string]string{"failureCode": "blocked-signal-run-not-found"})
		return
	}
	if run.Spec.Autonomy == nil || run.Spec.Autonomy.EscalationRoute == nil {
		writeTriggerResponse(request.Context(), response, http.StatusConflict, map[string]string{"failureCode": "blocked-signal-not-supported"})
		return
	}
	if run.Status.JobName == "" || !observedProvenanceHealthy(&run) {
		writeTriggerResponse(request.Context(), response, http.StatusConflict, map[string]string{"failureCode": "blocked-signal-runtime-not-observed"})
		return
	}
	token, err := readSecretToken(request.Context(), r.Client, run.Namespace, run.Spec.Autonomy.EscalationRoute.BlockedSignalTokenSecretRef)
	if err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusServiceUnavailable, map[string]string{"failureCode": "blocked-signal-token-unavailable"})
		return
	}
	provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), token) != 1 {
		writeTriggerResponse(request.Context(), response, http.StatusUnauthorized, map[string]string{"failureCode": "blocked-signal-unauthorized"})
		return
	}

	var submitted blockedSignalRequest
	if err := decodeStrictJSON(request.Body, 65_536, &submitted); err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusBadRequest, map[string]string{"failureCode": "blocked-signal-invalid"})
		return
	}
	subjectRevision, err := governancecontracts.AgentRunSubjectRevision(&run)
	if err != nil || submitted.SubjectRevision != subjectRevision ||
		strings.TrimSpace(submitted.Reporter) == "" || strings.TrimSpace(submitted.ReasonCode) == "" ||
		!governancecontracts.ContentAddressedEvidenceReference(submitted.EvidenceReference) {
		writeTriggerResponse(request.Context(), response, http.StatusBadRequest, map[string]string{"failureCode": "blocked-signal-invalid"})
		return
	}
	if run.Status.Blocked != nil {
		writeTriggerResponse(request.Context(), response, http.StatusConflict, map[string]string{"failureCode": "blocked-signal-already-recorded"})
		return
	}

	reportedAt := metav1.NewTime(time.Now().UTC())
	if r.Now != nil {
		reportedAt = metav1.NewTime(r.Now().UTC())
	}
	run.Status.Blocked = &agentv1alpha1.AgentBlockedStatus{
		ReportedAt:        reportedAt,
		SubjectRevision:   submitted.SubjectRevision,
		Reporter:          submitted.Reporter,
		ReasonCode:        submitted.ReasonCode,
		EvidenceReference: submitted.EvidenceReference,
	}
	if err := r.Client.Status().Update(request.Context(), &run); err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusInternalServerError, map[string]string{"failureCode": "blocked-signal-status-update-failed"})
		return
	}
	writeTriggerResponse(request.Context(), response, http.StatusAccepted, map[string]string{
		"run":    run.Namespace + "/" + run.Name,
		"status": "Blocked",
	})
}

type escalationResponseRequest struct {
	Decision          string `json:"decision"`
	SubjectRevision   string `json:"subjectRevision"`
	Actor             string `json:"actor"`
	ResponseReference string `json:"responseReference"`
}

func (r *AgentTriggerReceiver) serveEscalation(response http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		r.serveEscalationStatus(response, request)
		return
	}
	if request.Method != http.MethodPost {
		writeTriggerResponse(request.Context(), response, http.StatusMethodNotAllowed, map[string]string{"failureCode": "escalation-method-not-allowed"})
		return
	}
	namespace, name, matched := parseNamespacedPath(request.URL.Path, escalationPathPrefix)
	if !matched {
		writeTriggerResponse(request.Context(), response, http.StatusNotFound, map[string]string{"failureCode": "escalation-not-found"})
		return
	}

	var run agentv1alpha1.AgentRun
	if err := r.Client.Get(request.Context(), types.NamespacedName{Namespace: namespace, Name: name}, &run); err != nil {
		status := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeTriggerResponse(request.Context(), response, status, map[string]string{"failureCode": "escalation-not-found"})
		return
	}
	if run.Spec.Autonomy == nil || run.Spec.Autonomy.EscalationRoute == nil || run.Status.Escalation == nil {
		writeTriggerResponse(request.Context(), response, http.StatusConflict, map[string]string{"failureCode": "escalation-not-pending"})
		return
	}
	token, err := readSecretToken(request.Context(), r.Client, run.Namespace, run.Spec.Autonomy.EscalationRoute.ResponseTokenSecretRef)
	if err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusServiceUnavailable, map[string]string{"failureCode": "escalation-token-unavailable"})
		return
	}
	provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), token) != 1 {
		writeTriggerResponse(request.Context(), response, http.StatusUnauthorized, map[string]string{"failureCode": "escalation-unauthorized"})
		return
	}

	var submitted escalationResponseRequest
	if err := decodeStrictJSON(request.Body, 65_536, &submitted); err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusBadRequest, map[string]string{"failureCode": "escalation-response-invalid"})
		return
	}
	if (submitted.Decision != "approve" && submitted.Decision != "reject") ||
		submitted.SubjectRevision != run.Status.Escalation.SubjectRevision ||
		strings.TrimSpace(submitted.Actor) == "" || strings.TrimSpace(submitted.ResponseReference) == "" {
		writeTriggerResponse(request.Context(), response, http.StatusBadRequest, map[string]string{"failureCode": "escalation-response-invalid"})
		return
	}
	if run.Status.Escalation.Outcome != agentv1alpha1.EscalationOutcomePending {
		writeTriggerResponse(request.Context(), response, http.StatusConflict, map[string]string{"failureCode": "escalation-already-resolved"})
		return
	}

	outcome := agentv1alpha1.EscalationOutcomeApproved
	if submitted.Decision == "reject" {
		outcome = agentv1alpha1.EscalationOutcomeRejected
	}
	respondedAt := metav1.NewTime(time.Now().UTC())
	if r.Now != nil {
		respondedAt = metav1.NewTime(r.Now().UTC())
	}
	run.Status.Escalation.Outcome = outcome
	run.Status.Escalation.Responder = submitted.Actor
	run.Status.Escalation.ResponseReference = submitted.ResponseReference
	run.Status.Escalation.RespondedAt = &respondedAt
	if err := r.Client.Status().Update(request.Context(), &run); err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusInternalServerError, map[string]string{"failureCode": "escalation-status-update-failed"})
		return
	}
	writeTriggerResponse(request.Context(), response, http.StatusAccepted, map[string]string{
		"run":     run.Namespace + "/" + run.Name,
		"outcome": string(outcome),
	})
}

func (r *AgentTriggerReceiver) serveEscalationStatus(response http.ResponseWriter, request *http.Request) {
	namespace, name, matched := parseNamespacedPath(request.URL.Path, escalationPathPrefix)
	if !matched {
		writeTriggerResponse(request.Context(), response, http.StatusNotFound, map[string]string{"failureCode": "escalation-not-found"})
		return
	}

	var run agentv1alpha1.AgentRun
	if err := r.Client.Get(request.Context(), types.NamespacedName{Namespace: namespace, Name: name}, &run); err != nil {
		status := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeTriggerResponse(request.Context(), response, status, map[string]string{"failureCode": "escalation-not-found"})
		return
	}
	if run.Spec.Autonomy == nil || run.Spec.Autonomy.EscalationRoute == nil || run.Status.Escalation == nil {
		writeTriggerResponse(request.Context(), response, http.StatusConflict, map[string]string{"failureCode": "escalation-not-pending"})
		return
	}
	token, err := readSecretToken(request.Context(), r.Client, run.Namespace, run.Spec.Autonomy.EscalationRoute.BlockedSignalTokenSecretRef)
	if err != nil {
		writeTriggerResponse(request.Context(), response, http.StatusServiceUnavailable, map[string]string{"failureCode": "escalation-status-token-unavailable"})
		return
	}
	provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), token) != 1 {
		writeTriggerResponse(request.Context(), response, http.StatusUnauthorized, map[string]string{"failureCode": "escalation-status-unauthorized"})
		return
	}

	escalation := run.Status.Escalation
	writeTriggerResponse(request.Context(), response, http.StatusOK, map[string]any{
		"run":             run.Namespace + "/" + run.Name,
		"subjectRevision": escalation.SubjectRevision,
		"outcome":         escalation.Outcome,
		"expiresAt":       escalation.ExpiresAt,
		"safeDefault":     escalation.SafeDefault,
	})
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
