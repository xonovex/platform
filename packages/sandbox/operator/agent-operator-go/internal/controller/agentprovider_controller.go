package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// AgentProviderDependencies are the ports of the AgentProvider reconciler.
type AgentProviderDependencies struct {
	Store  ProviderStore
	Events EventRecorder
	Clock  Clock
}

// AgentProviderReconciler reconciles an AgentProvider object
type AgentProviderReconciler struct {
	deps AgentProviderDependencies
}

// NewAgentProviderReconciler returns an AgentProvider reconciler over its
// ports.
func NewAgentProviderReconciler(deps AgentProviderDependencies) *AgentProviderReconciler {
	return &AgentProviderReconciler{deps: deps}
}

func (r *AgentProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	provider, err := r.deps.Store.GetAgentProvider(ctx, req.NamespacedName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !provider.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	var secret *corev1.Secret
	if ref := provider.Spec.AuthTokenSecretRef; ref != nil {
		secret, err = r.deps.Store.GetSecret(ctx, types.NamespacedName{Namespace: provider.Namespace, Name: ref.Name})
		if err != nil && !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	desired, transition := decideProviderStatus(provider, secret, r.deps.Clock.Now())
	if !desired.Status.Ready {
		log.FromContext(ctx).Info("referenced secret or key not found", "provider", req.NamespacedName)
	}
	if err := r.deps.Store.UpdateAgentProviderStatus(ctx, desired); err != nil {
		return ctrl.Result{}, err
	}
	if transition != nil {
		r.deps.Events.Record(desired, transition.eventType, transition.reason, transition.message)
	}
	return ctrl.Result{}, nil
}

// decideProviderStatus returns the provider with the Ready status that secret
// decides, and the event to record when readiness of a secret-backed provider
// changes. A nil secret means the referenced Secret does not exist.
func decideProviderStatus(provider *agentv1alpha1.AgentProvider, secret *corev1.Secret, now time.Time) (*agentv1alpha1.AgentProvider, *event) {
	ref := provider.Spec.AuthTokenSecretRef
	ready := ref == nil
	if ref != nil && secret != nil {
		_, ready = secret.Data[ref.Key]
	}

	conditionStatus := metav1.ConditionFalse
	message := "Referenced secret or key not found"
	if ready {
		conditionStatus = metav1.ConditionTrue
		message = "Provider is ready"
	}
	previous := meta.FindStatusCondition(provider.Status.Conditions, "Ready")
	changed := previous == nil || previous.Status != conditionStatus || previous.ObservedGeneration != provider.Generation

	desired := provider.DeepCopy()
	desired.Status.Ready = ready
	meta.SetStatusCondition(&desired.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             conditionStatus,
		ObservedGeneration: provider.Generation,
		LastTransitionTime: metav1.NewTime(now),
		Reason:             "SecretValidation",
		Message:            message,
	})

	if ref == nil || !changed {
		return desired, nil
	}
	if ready {
		resolved := normalEvent("ProviderSecretResolved", fmt.Sprintf("Secret %s key %s resolved successfully", ref.Name, ref.Key))
		return desired, &resolved
	}
	missing := warningEvent("ProviderSecretMissing", fmt.Sprintf("Secret %s not found or key %s missing", ref.Name, ref.Key))
	return desired, &missing
}

// providersReferencingSecret returns a reconcile request for each provider
// whose auth token names the Secret.
func providersReferencingSecret(providers []agentv1alpha1.AgentProvider, secretName string) []reconcile.Request {
	requests := make([]reconcile.Request, 0)
	for index := range providers {
		provider := &providers[index]
		if provider.Spec.AuthTokenSecretRef != nil && provider.Spec.AuthTokenSecretRef.Name == secretName {
			requests = append(requests, reconcile.Request{NamespacedName: keyOf(provider)})
		}
	}
	return requests
}

func (r *AgentProviderReconciler) providersForSecret(ctx context.Context, object client.Object) []reconcile.Request {
	secret, ok := object.(*corev1.Secret)
	if !ok {
		return nil
	}
	providers, err := r.deps.Store.ListAgentProviders(ctx, secret.Namespace)
	if err != nil {
		log.FromContext(ctx).Error(err, "list providers for secret", "secret", keyOf(secret))
		return nil
	}
	return providersReferencingSecret(providers, secret.Name)
}

// SetupWithManager registers the reconciler with the manager.
func (r *AgentProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentProvider{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.providersForSecret)).
		Complete(r)
}
