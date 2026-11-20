package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/isolation/shared"
)

// event is a Kubernetes event a decision asks the reconciler to record.
type event struct {
	eventType string
	reason    string
	message   string
}

func normalEvent(reason, message string) event {
	return event{eventType: corev1.EventTypeNormal, reason: reason, message: message}
}

func warningEvent(reason, message string) event {
	return event{eventType: corev1.EventTypeWarning, reason: reason, message: message}
}

// phaseCondition returns the condition that carries the message of a phase
// change, stamped with now.
func phaseCondition(phase, message string, now time.Time) metav1.Condition {
	return metav1.Condition{
		Type:               phase,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(now),
		Reason:             phase,
		Message:            message,
	}
}

// setController makes owner, of the given kind, the controller of object.
func setController(owner metav1.Object, kind string, object metav1.Object) {
	reference := metav1.NewControllerRef(owner, agentv1alpha1.GroupVersion.WithKind(kind))
	object.SetOwnerReferences(append(object.GetOwnerReferences(), *reference))
}

func keyOf(object metav1.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: object.GetNamespace(), Name: object.GetName()}
}

// verifyControlledResource accepts an existing object only when owner controls
// it and its specification matches the desired one.
func verifyControlledResource(owner metav1.Object, ownerKind string, existing metav1.Object, desiredSpec, existingSpec any) error {
	if !metav1.IsControlledBy(existing, owner) {
		return fmt.Errorf("resource is not controlled by %s UID %q", ownerKind, owner.GetUID())
	}
	if !equality.Semantic.DeepDerivative(desiredSpec, existingSpec) {
		return fmt.Errorf("resource specification differs from the admitted execution")
	}
	return nil
}

// ensureNetworkPolicy creates the desired NetworkPolicy or verifies the one that
// exists, and reports whether it created it.
func ensureNetworkPolicy(ctx context.Context, store resourceStore, owner metav1.Object, ownerKind string, desired *networkingv1.NetworkPolicy) (bool, error) {
	err := store.CreateNetworkPolicy(ctx, desired)
	if err == nil {
		return true, nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return false, fmt.Errorf("create NetworkPolicy: %w", err)
	}
	existing, err := store.GetNetworkPolicy(ctx, keyOf(desired))
	if err != nil {
		return false, fmt.Errorf("get existing NetworkPolicy: %w", err)
	}
	if err := verifyControlledResource(owner, ownerKind, existing, desired.Spec, existing.Spec); err != nil {
		return false, fmt.Errorf("existing NetworkPolicy %q: %w", desired.Name, err)
	}
	return false, nil
}

// ensureJob creates the desired Job or verifies the one that exists, and
// reports whether it created it.
func ensureJob(ctx context.Context, store resourceStore, owner metav1.Object, ownerKind string, desired *batchv1.Job) (bool, error) {
	err := store.CreateJob(ctx, desired)
	if err == nil {
		return true, nil
	}
	if !apierrors.IsAlreadyExists(err) {
		return false, fmt.Errorf("create Job: %w", err)
	}
	existing, err := store.GetJob(ctx, keyOf(desired))
	if err != nil {
		return false, fmt.Errorf("get existing Job: %w", err)
	}
	if err := verifyControlledResource(owner, ownerKind, existing, desired.Spec, existing.Spec); err != nil {
		return false, fmt.Errorf("existing Job %q: %w", desired.Name, err)
	}
	return false, nil
}

// ensureClaim creates a PersistentVolumeClaim; an existing claim of the same
// name is the one an earlier reconcile created.
func ensureClaim(ctx context.Context, store resourceStore, claim *corev1.PersistentVolumeClaim) error {
	if err := store.CreatePersistentVolumeClaim(ctx, claim); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create PersistentVolumeClaim %q: %w", claim.Name, err)
	}
	return nil
}

// ensureAgentServiceAccount creates the dedicated zero-RBAC ServiceAccount used
// by all untrusted agent and workspace-init pods. It is a shared namespace
// resource, so it has no owner reference.
func ensureAgentServiceAccount(ctx context.Context, store resourceStore, namespace string) error {
	if err := store.CreateServiceAccount(ctx, isoshared.BuildAgentServiceAccount(namespace)); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create agent ServiceAccount: %w", err)
	}
	return nil
}

// jobOutcome reports the terminal condition of a Job, if it has one.
func jobOutcome(job *batchv1.Job) (complete bool, failed *batchv1.JobCondition) {
	for index := range job.Status.Conditions {
		condition := &job.Status.Conditions[index]
		if condition.Status != corev1.ConditionTrue {
			continue
		}
		switch condition.Type {
		case batchv1.JobComplete:
			return true, nil
		case batchv1.JobFailed:
			return false, condition
		}
	}
	return false, nil
}
