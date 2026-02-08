package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AgentProviderReconciler reconciles an AgentProvider object
type AgentProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AgentProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var provider agentv1alpha1.AgentProvider
	if err := r.Get(ctx, req.NamespacedName, &provider); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Validate secret reference exists
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
			// Check if the key exists in the secret
			if _, ok := secret.Data[provider.Spec.AuthTokenSecretRef.Key]; !ok {
				ready = false
				log.Info("key not found in secret", "secret", secretName, "key", provider.Spec.AuthTokenSecretRef.Key)
			}
		}
	}

	// Update status
	provider.Status.Ready = ready
	condition := metav1.Condition{
		Type:               "Ready",
		LastTransitionTime: metav1.Now(),
		Reason:             "SecretValidation",
	}
	if ready {
		condition.Status = metav1.ConditionTrue
		condition.Message = "Provider is ready"
	} else {
		condition.Status = metav1.ConditionFalse
		condition.Message = "Referenced secret or key not found"
	}
	provider.Status.Conditions = []metav1.Condition{condition}

	if err := r.Status().Update(ctx, &provider); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AgentProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1alpha1.AgentProvider{}).
		Complete(r)
}
