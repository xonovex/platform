package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// AgentScheduleReconciler materializes time-based triggers as immutable AgentRuns.
type AgentScheduleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Now    func() time.Time
}

// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentschedules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentschedules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agent.xonovex.com,resources=agentruns,verbs=get;list;watch;create;delete

func (r *AgentScheduleReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	var schedule agentv1alpha1.AgentSchedule
	if err := r.Get(ctx, request.NamespacedName, &schedule); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	parsed, err := cron.ParseStandard(schedule.Spec.Schedule)
	if err != nil {
		return r.setInvalidSchedule(ctx, &schedule, err)
	}
	if schedule.Spec.Suspend {
		return ctrl.Result{}, nil
	}

	now := time.Now().UTC()
	if r.Now != nil {
		now = r.Now().UTC()
	}
	baseline := schedule.CreationTimestamp.Time
	if schedule.Status.LastScheduleTime != nil {
		baseline = schedule.Status.LastScheduleTime.Time
	}
	next := parsed.Next(baseline)
	if next.After(now) {
		return ctrl.Result{RequeueAfter: next.Sub(now)}, nil
	}

	if schedule.Status.ActiveRunName != "" {
		active, activeErr := r.activeRun(ctx, &schedule)
		if activeErr != nil {
			return ctrl.Result{}, activeErr
		}
		if active != nil {
			switch schedule.Spec.ConcurrencyPolicy {
			case agentv1alpha1.AgentScheduleConcurrencyForbid:
				return ctrl.Result{RequeueAfter: parsed.Next(now).Sub(now)}, nil
			case agentv1alpha1.AgentScheduleConcurrencyReplace:
				if err := r.Delete(ctx, active); err != nil && !apierrors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}
		}
	}

	run := buildTriggeredRun(schedule.Spec.Template, schedule.Namespace, schedule.Name, "AgentSchedule", fmt.Sprintf("%d", next.Unix()))
	if err := ctrl.SetControllerReference(&schedule, run, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.Create(ctx, run); err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}

	scheduledAt := metav1.NewTime(next)
	schedule.Status.LastScheduleTime = &scheduledAt
	schedule.Status.ActiveRunName = run.Name
	apimeta.SetStatusCondition(&schedule.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "RunScheduled",
		Message:            "scheduled run created",
		ObservedGeneration: schedule.Generation,
	})
	if err := r.Status().Update(ctx, &schedule); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: parsed.Next(now).Sub(now)}, nil
}

func (r *AgentScheduleReconciler) activeRun(ctx context.Context, schedule *agentv1alpha1.AgentSchedule) (*agentv1alpha1.AgentRun, error) {
	var run agentv1alpha1.AgentRun
	err := r.Get(ctx, types.NamespacedName{Namespace: schedule.Namespace, Name: schedule.Status.ActiveRunName}, &run)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if run.Status.Phase == agentv1alpha1.AgentRunPhaseSucceeded ||
		run.Status.Phase == agentv1alpha1.AgentRunPhaseFailed ||
		run.Status.Phase == agentv1alpha1.AgentRunPhaseTimedOut ||
		run.Status.Phase == agentv1alpha1.AgentRunPhasePaused {
		return nil, nil
	}
	return &run, nil
}

func (r *AgentScheduleReconciler) setInvalidSchedule(ctx context.Context, schedule *agentv1alpha1.AgentSchedule, parseError error) (ctrl.Result, error) {
	apimeta.SetStatusCondition(&schedule.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "InvalidSchedule",
		Message:            parseError.Error(),
		ObservedGeneration: schedule.Generation,
	})
	return ctrl.Result{}, r.Status().Update(ctx, schedule)
}

func (r *AgentScheduleReconciler) SetupWithManager(manager ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(manager).
		For(&agentv1alpha1.AgentSchedule{}).
		Owns(&agentv1alpha1.AgentRun{}).
		Complete(r)
}
