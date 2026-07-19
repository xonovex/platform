package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	runtel "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/telemetry"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

type recordingTelemetrySink struct {
	signals []runtel.Signal
}

func (sink *recordingTelemetrySink) Record(_ context.Context, signal runtel.Signal) {
	sink.signals = append(sink.signals, signal)
}

func TestAgentScheduleReconciler_CreatesJournaledRun(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 30, 0, time.UTC)
	schedule := &agentv1alpha1.AgentSchedule{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "daily-review",
			Namespace:         "test",
			CreationTimestamp: metav1.NewTime(now.Add(-2 * time.Minute)),
		},
		Spec: agentv1alpha1.AgentScheduleSpec{
			Schedule: "* * * * *",
			Template: agentv1alpha1.AgentRunTemplate{Spec: unattendedRunSpec()},
		},
	}
	scheme := testutil.NewScheme()
	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentSchedule{}, &agentv1alpha1.AgentRun{}).
		WithObjects(schedule).
		Build()
	reconciler := &AgentScheduleReconciler{Client: kubeClient, Scheme: scheme, Now: func() time.Time { return now }}

	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}); err != nil {
		t.Fatalf("reconcile schedule: %v", err)
	}
	var runs agentv1alpha1.AgentRunList
	if err := kubeClient.List(context.Background(), &runs, client.InNamespace("test")); err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(runs.Items) != 1 {
		t.Fatalf("created runs = %d, want 1", len(runs.Items))
	}
	if runs.Items[0].Spec.AccountableOwner != "team:platform" {
		t.Fatalf("accountable owner = %q", runs.Items[0].Spec.AccountableOwner)
	}

	runReconciler := &AgentRunReconciler{Client: kubeClient, Now: func() time.Time { return now }}
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&runs.Items[0])}
	if _, err := runReconciler.Reconcile(context.Background(), request); err != nil {
		t.Fatalf("reconcile triggered run journal: %v", err)
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), request.NamespacedName, &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Journal == nil || updated.Status.Journal.PromptReference != "prompt://scheduled-review/1" {
		t.Fatalf("journal = %#v", updated.Status.Journal)
	}
}

func TestAgentTriggerReceiver_AuthenticationControlsCreation(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	trigger := &agentv1alpha1.AgentTrigger{
		ObjectMeta: metav1.ObjectMeta{Name: "repository-event", Namespace: "test"},
		Spec: agentv1alpha1.AgentTriggerSpec{
			Endpoint:       "repository",
			TokenSecretRef: agentv1alpha1.SecretKeyRef{Name: "trigger-token", Key: "token"},
			Template:       agentv1alpha1.AgentRunTemplate{Spec: unattendedRunSpec()},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "trigger-token", Namespace: "test"},
		Data:       map[string][]byte{"token": []byte("correct-token")},
	}
	scheme := testutil.NewScheme()
	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentTrigger{}, &agentv1alpha1.AgentRun{}).
		WithObjects(trigger, secret).
		Build()
	receiver := &AgentTriggerReceiver{Client: kubeClient, Scheme: scheme, Now: func() time.Time { return now }}

	unauthorized := httptest.NewRecorder()
	receiver.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/v1/triggers/test/repository", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	unmatched := httptest.NewRecorder()
	unmatchedRequest := httptest.NewRequest(http.MethodPost, "/v1/triggers/test/missing", nil)
	unmatchedRequest.Header.Set("Authorization", "Bearer correct-token")
	receiver.ServeHTTP(unmatched, unmatchedRequest)
	if unmatched.Code != http.StatusNotFound {
		t.Fatalf("unmatched status = %d", unmatched.Code)
	}

	authorized := httptest.NewRecorder()
	authorizedRequest := httptest.NewRequest(http.MethodPost, "/v1/triggers/test/repository", nil)
	authorizedRequest.Header.Set("Authorization", "Bearer correct-token")
	authorizedRequest.Header.Set("Idempotency-Key", "event-1")
	receiver.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusCreated {
		t.Fatalf("authorized status = %d, body = %s", authorized.Code, authorized.Body.String())
	}

	var runs agentv1alpha1.AgentRunList
	if err := kubeClient.List(context.Background(), &runs, client.InNamespace("test")); err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(runs.Items) != 1 {
		t.Fatalf("created runs = %d, want 1", len(runs.Items))
	}
	if runs.Items[0].Annotations[agentv1alpha1.TriggeredByKindAnnotation] != "AgentTrigger" {
		t.Fatalf("trigger annotation = %#v", runs.Items[0].Annotations)
	}

	runReconciler := &AgentRunReconciler{Client: kubeClient, Now: func() time.Time { return now }}
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&runs.Items[0])}
	if _, err := runReconciler.Reconcile(context.Background(), request); err != nil {
		t.Fatalf("reconcile triggered run journal: %v", err)
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), request.NamespacedName, &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Journal == nil {
		t.Fatal("triggered run journal is nil")
	}
}

func TestAgentRunReconciler_ExpiredEscalationUsesSafeDefault(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "needs-review", Namespace: "test",
			Annotations: map[string]string{governanceCorrelationAnnotation: "correlation-1"},
		},
		Spec: unattendedRunSpec(),
		Status: agentv1alpha1.AgentRunStatus{
			Journal: &agentv1alpha1.AgentRunJournal{PromptReference: "prompt://scheduled-review/1"},
			Escalation: &agentv1alpha1.AgentEscalationStatus{
				Recipient:   "human:on-call",
				RequestedAt: metav1.NewTime(now.Add(-2 * time.Minute)),
				ExpiresAt:   metav1.NewTime(now.Add(-time.Minute)),
				SafeDefault: agentv1alpha1.EscalationSafeDefaultPause,
				Outcome:     agentv1alpha1.EscalationOutcomePending,
			},
		},
	}
	run.Spec.Autonomy.NeedsHuman = true
	kubeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run).
		Build()
	reconciler := &AgentRunReconciler{Client: kubeClient, Now: func() time.Time { return now }}

	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)}); err != nil {
		t.Fatalf("reconcile escalation: %v", err)
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Phase != agentv1alpha1.AgentRunPhasePaused || updated.Status.Escalation.Outcome != agentv1alpha1.EscalationOutcomePaused {
		t.Fatalf("phase/outcome = %s/%s", updated.Status.Phase, updated.Status.Escalation.Outcome)
	}
	if updated.Status.JobName != "" {
		t.Fatalf("silence advanced to job %q", updated.Status.JobName)
	}
}

func TestAgentRunReconciler_DriftDemotesAndContainsLiveRun(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{
			Name: "drifting", Namespace: "test",
			Annotations: map[string]string{
				governanceCorrelationAnnotation: "correlation-drift",
				oversightStateAnnotation:        "degraded",
			},
		},
		Spec: unattendedRunSpec(),
		Status: agentv1alpha1.AgentRunStatus{
			Phase:             agentv1alpha1.AgentRunPhaseRunning,
			JobName:           "drifting-job",
			Journal:           &agentv1alpha1.AgentRunJournal{PromptReference: "prompt://scheduled-review/1"},
			EffectiveAutonomy: agentv1alpha1.AutonomyLevelUnattended,
		},
	}
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "drifting-job", Namespace: "test"}}
	kubeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, job).
		Build()
	sink := &recordingTelemetrySink{}
	reconciler := &AgentRunReconciler{Client: kubeClient, Now: func() time.Time { return now }, Telemetry: sink}

	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)}); err != nil {
		t.Fatalf("reconcile drift: %v", err)
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Phase != agentv1alpha1.AgentRunPhasePaused || updated.Status.EffectiveAutonomy != agentv1alpha1.AutonomyLevelSupervised {
		t.Fatalf("phase/autonomy = %s/%s", updated.Status.Phase, updated.Status.EffectiveAutonomy)
	}
	if updated.Status.Containment == nil || updated.Status.Containment.Action != "kill-and-pause" {
		t.Fatalf("containment = %#v", updated.Status.Containment)
	}
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(job), &batchv1.Job{}); !apierrors.IsNotFound(err) {
		t.Fatalf("contained Job still exists or get failed: %v", err)
	}
	foundContainment := false
	for _, signal := range sink.signals {
		if signal.Kind == "incident.containment" && signal.CorrelationID == "correlation-drift" {
			foundContainment = true
		}
	}
	if !foundContainment {
		t.Fatalf("signals = %#v", sink.signals)
	}
}

func unattendedRunSpec() agentv1alpha1.AgentRunSpec {
	return agentv1alpha1.AgentRunSpec{
		AccountableOwner: "team:platform",
		WorkspaceRef:     "workspace",
		Provenance: &agentv1alpha1.AgentRunProvenance{
			Model:              "claude-sonnet",
			Provider:           "anthropic",
			PromptReference:    "prompt://scheduled-review/1",
			Tools:              []string{"Read"},
			GrantedPermissions: []string{"repository:read"},
		},
		Autonomy: &agentv1alpha1.AgentAutonomySpec{
			Level:            agentv1alpha1.AutonomyLevelUnattended,
			ProtectedTargets: []string{"repository:main"},
			EscalationRoute: &agentv1alpha1.AgentEscalationRoute{
				Recipient:   "human:on-call",
				Window:      metav1.Duration{Duration: time.Minute},
				SafeDefault: agentv1alpha1.EscalationSafeDefaultPause,
			},
		},
	}
}
