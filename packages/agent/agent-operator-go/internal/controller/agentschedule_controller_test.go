package controller

import (
	"context"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestAgentScheduleReconcilerRejectsInvalidSchedule(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	schedule := newAgentSchedule(now)
	schedule.Spec.Schedule = "not a cron expression"
	reconciler, kubeClient := newScheduleReconciler(schedule, now)

	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}); err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var updated agentv1alpha1.AgentSchedule
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(schedule), &updated); err != nil {
		t.Fatalf("get AgentSchedule: %v", err)
	}
	condition := apimeta.FindStatusCondition(updated.Status.Conditions, "Ready")
	if condition == nil || condition.Status != metav1.ConditionFalse || condition.Reason != "InvalidSchedule" {
		t.Fatalf("Ready condition = %#v, want invalid schedule", condition)
	}
}

func TestAgentScheduleReconcilerHonorsSuspension(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	schedule := newAgentSchedule(now)
	schedule.Spec.Suspend = true
	reconciler, kubeClient := newScheduleReconciler(schedule, now)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("result = %#v, want no requeue", result)
	}
	assertAgentRunCount(t, kubeClient, 0)
}

func TestAgentScheduleReconcilerWaitsForNextOccurrence(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	schedule := newAgentSchedule(now)
	schedule.CreationTimestamp = metav1.NewTime(now)
	reconciler, kubeClient := newScheduleReconciler(schedule, now)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result.RequeueAfter != time.Minute {
		t.Fatalf("RequeueAfter = %s, want %s", result.RequeueAfter, time.Minute)
	}
	assertAgentRunCount(t, kubeClient, 0)
}

func TestAgentScheduleReconcilerCreatesDueRun(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	schedule := newAgentSchedule(now)
	reconciler, kubeClient := newScheduleReconciler(schedule, now)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result.RequeueAfter != time.Minute {
		t.Fatalf("RequeueAfter = %s, want %s", result.RequeueAfter, time.Minute)
	}

	var updated agentv1alpha1.AgentSchedule
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(schedule), &updated); err != nil {
		t.Fatalf("get AgentSchedule: %v", err)
	}
	if updated.Status.ActiveRunName == "" || updated.Status.LastScheduleTime == nil || !updated.Status.LastScheduleTime.Time.Equal(now.Add(-time.Minute)) {
		t.Fatalf("status = %#v, want run scheduled for prior minute", updated.Status)
	}
	condition := apimeta.FindStatusCondition(updated.Status.Conditions, "Ready")
	if condition == nil || condition.Status != metav1.ConditionTrue || condition.Reason != "RunScheduled" {
		t.Fatalf("Ready condition = %#v, want scheduled run", condition)
	}
	var run agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), types.NamespacedName{Namespace: "team", Name: updated.Status.ActiveRunName}, &run); err != nil {
		t.Fatalf("get created AgentRun: %v", err)
	}
	if run.Spec.Prompt != "scheduled work" || len(run.OwnerReferences) != 1 || run.OwnerReferences[0].Name != schedule.Name {
		t.Fatalf("created run = %#v, want schedule template and owner", run)
	}
}

func TestAgentScheduleReconcilerHonorsConcurrencyPolicy(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		policy       agentv1alpha1.AgentScheduleConcurrencyPolicy
		wantOldRun   bool
		wantRunCount int
	}{
		{name: "forbid", policy: agentv1alpha1.AgentScheduleConcurrencyForbid, wantOldRun: true, wantRunCount: 1},
		{name: "replace", policy: agentv1alpha1.AgentScheduleConcurrencyReplace, wantRunCount: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schedule := newAgentSchedule(now)
			schedule.Spec.ConcurrencyPolicy = test.policy
			schedule.Status.ActiveRunName = "active"
			active := testutil.NewAgentRun("team", "active", testutil.WithPhase(agentv1alpha1.AgentRunPhaseRunning))
			reconciler, kubeClient := newScheduleReconcilerWithObjects(schedule, now, active)

			if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}); err != nil {
				t.Fatalf("Reconcile() error = %v", err)
			}
			assertAgentRunCount(t, kubeClient, test.wantRunCount)
			var oldRun agentv1alpha1.AgentRun
			err := kubeClient.Get(context.Background(), types.NamespacedName{Namespace: "team", Name: "active"}, &oldRun)
			if test.wantOldRun && err != nil {
				t.Fatalf("get active AgentRun: %v", err)
			}
			if !test.wantOldRun && !apierrors.IsNotFound(err) {
				t.Fatalf("get replaced AgentRun error = %v, want not found", err)
			}
		})
	}
}

func TestAgentScheduleActiveRunIgnoresCompletedRun(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	terminalPhases := []agentv1alpha1.AgentRunPhase{
		agentv1alpha1.AgentRunPhaseSucceeded,
		agentv1alpha1.AgentRunPhaseFailed,
		agentv1alpha1.AgentRunPhaseTimedOut,
	}

	for _, phase := range terminalPhases {
		t.Run(string(phase), func(t *testing.T) {
			schedule := newAgentSchedule(now)
			schedule.Status.ActiveRunName = "completed"
			completed := testutil.NewAgentRun("team", "completed", testutil.WithPhase(phase))
			reconciler, _ := newScheduleReconcilerWithObjects(schedule, now, completed)

			active, err := reconciler.activeRun(context.Background(), schedule)
			if err != nil {
				t.Fatalf("activeRun() error = %v", err)
			}
			if active != nil {
				t.Fatalf("activeRun() = %#v, want nil for phase %q", active, phase)
			}
		})
	}
}

func newAgentSchedule(now time.Time) *agentv1alpha1.AgentSchedule {
	templateRun := testutil.NewAgentRun("team", "template", testutil.WithPrompt("scheduled work"))
	return &agentv1alpha1.AgentSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nightly",
			Namespace:         "team",
			UID:               types.UID("schedule-uid"),
			CreationTimestamp: metav1.NewTime(now.Add(-2 * time.Minute)),
		},
		Spec: agentv1alpha1.AgentScheduleSpec{
			Schedule:          "* * * * *",
			ConcurrencyPolicy: agentv1alpha1.AgentScheduleConcurrencyAllow,
			Template:          agentv1alpha1.AgentRunTemplate{Spec: templateRun.Spec},
		},
	}
}

func newScheduleReconciler(schedule *agentv1alpha1.AgentSchedule, now time.Time) (*AgentScheduleReconciler, client.Client) {
	return newScheduleReconcilerWithObjects(schedule, now)
}

func newScheduleReconcilerWithObjects(schedule *agentv1alpha1.AgentSchedule, now time.Time, objects ...client.Object) (*AgentScheduleReconciler, client.Client) {
	scheme := testutil.NewScheme()
	allObjects := append([]client.Object{schedule}, objects...)
	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentSchedule{}, &agentv1alpha1.AgentRun{}).
		WithObjects(allObjects...).
		Build()
	return &AgentScheduleReconciler{Client: kubeClient, Scheme: scheme, Now: func() time.Time { return now }}, kubeClient
}

func assertAgentRunCount(t *testing.T, kubeClient client.Client, want int) {
	t.Helper()
	var runs agentv1alpha1.AgentRunList
	if err := kubeClient.List(context.Background(), &runs, client.InNamespace("team")); err != nil {
		t.Fatalf("list AgentRuns: %v", err)
	}
	if len(runs.Items) != want {
		t.Fatalf("AgentRun count = %d, want %d", len(runs.Items), want)
	}
}
