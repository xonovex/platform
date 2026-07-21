package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AgentProviderReconciler reconciles an AgentProvider object
type AgentProviderReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder events.EventRecorder
}

func (r *AgentProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues(
		"provider", req.Name,
		"namespace", req.Namespace,
	)

	var provider agentv1alpha1.AgentProvider
	if err := r.Get(ctx, req.NamespacedName, &provider); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !provider.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	ready := true
	if provider.Spec.AuthTokenSecretRef != nil {
		var secret corev1.Secret
		secretName := types.NamespacedName{
			Name:      provider.Spec.AuthTokenSecretRef.Name,
			Namespace: provider.Namespace,
		}
		if err := r.Get(ctx, secretName, &secret); err != nil {
			if errors.IsNotFound(err) {
				ready = false
				log.Info("referenced secret not found", "secret", secretName)
			} else {
				return ctrl.Result{}, err
			}
		} else {
			if _, ok := secret.Data[provider.Spec.AuthTokenSecretRef.Key]; !ok {
				ready = false
				log.Info("key not found in secret", "secret", secretName, "key", provider.Spec.AuthTokenSecretRef.Key)
			}
		}
	}

	conditionStatus := metav1.ConditionFalse
	message := "Referenced secret or key not found"
	if ready {
		conditionStatus = metav1.ConditionTrue
		message = "Provider is ready"
	}
	previous := meta.FindStatusCondition(provider.Status.Conditions, "Ready")
	reportTransition := provider.Spec.AuthTokenSecretRef != nil &&
		(previous == nil || previous.Status != conditionStatus || previous.ObservedGeneration != provider.Generation)

	desired := provider.DeepCopy()
	desired.Status.Ready = ready
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             conditionStatus,
		ObservedGeneration: provider.Generation,
		Reason:             "SecretValidation",
		Message:            message,
	}
	meta.SetStatusCondition(&desired.Status.Conditions, condition)

	if err := r.updateProviderStatus(ctx, desired); err != nil {
		return ctrl.Result{}, err
	}
	if reportTransition {
		r.recordSecretTransition(desired, ready)
	}

	return ctrl.Result{}, nil
}

func (r *AgentProviderReconciler) recordSecretTransition(provider *agentv1alpha1.AgentProvider, ready bool) {
	if r.Recorder == nil || provider.Spec.AuthTokenSecretRef == nil {
		return
	}
	ref := provider.Spec.AuthTokenSecretRef
	if ready {
		r.Recorder.Eventf(provider, nil, corev1.EventTypeNormal, "ProviderSecretResolved", "ProviderSecretResolved",
			"Secret %s key %s resolved successfully", ref.Name, ref.Key)
		return
	}
	r.Recorder.Eventf(provider, nil, corev1.EventTypeWarning, "ProviderSecretMissing", "ProviderSecretMissing",
		"Secret %s not found or key %s missing", ref.Name, ref.Key)
}

func (r *AgentProviderReconciler) updateProviderStatus(ctx context.Context, provider *agentv1alpha1.AgentProvider) error {
	desired := provider.DeepCopy().Status
	key := client.ObjectKeyFromObject(provider)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var current agentv1alpha1.AgentProvider
		if err := r.Get(ctx, key, &current); err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if equality.Semantic.DeepEqual(current.Status, desired) {
			*provider = current
			return nil
		}
		current.Status = desired
		if err := r.Status().Update(ctx, &current); err != nil {
			return err
		}
		*provider = current
		return nil
	})
}

func (r *AgentProviderReconciler) providersForSecret(ctx context.Context, object client.Object) []reconcile.Request {
	secret, ok := object.(*corev1.Secret)
	if !ok {
		return nil
	}

	var providers agentv1alpha1.AgentProviderList
	if err := r.List(ctx, &providers, client.InNamespace(secret.Namespace)); err != nil {
		log.FromContext(ctx).Error(err, "list providers for secret", "secret", client.ObjectKeyFromObject(secret))
		return nil
	}

	requests := make([]reconcile.Request, 0)
	for i := range providers.Items {
		provider := &providers.Items[i]
		if provider.Spec.AuthTokenSecretRef != nil && provider.Spec.AuthTokenSecretRef.Name == secret.Name {
			requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(provider)})
		}
	}
	return requests
}

func (r *AgentProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentProvider{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.providersForSecret)).
		Complete(r)
}
