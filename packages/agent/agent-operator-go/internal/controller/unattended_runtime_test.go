package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	governancecontracts "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/governance"
	runtel "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/telemetry"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

type recordingTelemetrySink struct {
	signals []runtel.Signal
}

type stubEscalationRouter struct {
	deliveryReference string
	err               error
}

func (router *stubEscalationRouter) Raise(_ context.Context, _ *agentv1alpha1.AgentRun, _ *agentv1alpha1.AgentEscalationStatus) (string, error) {
	return router.deliveryReference, router.err
}

func markRunAdmissionEvidence(t *testing.T, run *agentv1alpha1.AgentRun) {
	t.Helper()
	operation := map[string]any{"kind": "development", "input": map[string]any{"workShape": "adaptive"}}
	encodedOperation := `{"input":{"workShape":"adaptive"},"kind":"development"}`
	operationDigest, err := governancecontracts.DigestJSON(operation)
	if err != nil {
		t.Fatalf("digest operation: %v", err)
	}
	subjectRevision, err := governancecontracts.AgentRunSubjectRevision(run)
	if err != nil {
		t.Fatalf("digest run spec: %v", err)
	}
	protectedTargetDigest, err := governancecontracts.DigestJSON([]string{"repository:main"})
	if err != nil {
		t.Fatalf("digest protected targets: %v", err)
	}
	if run.Annotations == nil {
		run.Annotations = map[string]string{}
	}
	run.Annotations[governanceCorrelationAnnotation] = "correlation-1"
	run.Annotations[governanceOperationAnnotation] = encodedOperation
	run.Annotations[governanceDecisionAnnotation] = "allow"
	run.Annotations[governanceSubjectRevisionAnnotation] = subjectRevision
	run.Annotations[governancePolicyVersionAnnotation] = "governance-policy/1"
	run.Annotations[governanceEvaluatorVersionAnnotation] = "governance-evaluator/1"
	run.Annotations[governanceOperationDigestAnnotation] = operationDigest
	run.Annotations[governanceDecisionEvidenceAnnotation] = "evidence://decision#sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	run.Annotations[governanceEnforcementEvidenceAnnotation] = "evidence://enforcement#sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	run.Annotations[governanceProtectedTargetsDigestAnnotation] = protectedTargetDigest
}

func runJournal(run *agentv1alpha1.AgentRun) *agentv1alpha1.AgentRunJournal {
	return &agentv1alpha1.AgentRunJournal{
		Generation:              run.Generation,
		AccountableOwner:        run.Spec.AccountableOwner,
		Model:                   run.Spec.Provenance.Model,
		Provider:                run.Spec.Provenance.Provider,
		PromptReference:         run.Spec.Provenance.PromptReference,
		Tools:                   append([]string(nil), run.Spec.Provenance.Tools...),
		GrantedPermissions:      append([]string(nil), run.Spec.Provenance.GrantedPermissions...),
		ExecutionImage:          "ghcr.io/example/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		RuntimeClassName:        "kata",
		AgentType:               agentv1alpha1.AgentTypeClaude,
		ProviderEnvironmentKeys: []string{},
		ExecutionSpecDigest:     "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
	}
}

func markRunBlocked(t *testing.T, run *agentv1alpha1.AgentRun) {
	t.Helper()
	subjectRevision, err := governancecontracts.AgentRunSubjectRevision(run)
	if err != nil {
		t.Fatalf("digest blocked run: %v", err)
	}
	run.Status.Blocked = &agentv1alpha1.AgentBlockedStatus{
		ReportedAt:        metav1.NewTime(time.Now().UTC()),
		SubjectRevision:   subjectRevision,
		Reporter:          "agent-harness:claude",
		ReasonCode:        "accountable-decision-required",
		EvidenceReference: "evidence://blocked#sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	}
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

func TestAgentTriggerReceiver_StatusUpdateFailureIsReported(t *testing.T) {
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
	baseClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentTrigger{}, &agentv1alpha1.AgentRun{}).
		WithObjects(trigger, secret).
		Build()
	failingClient := interceptor.NewClient(baseClient, interceptor.Funcs{
		SubResourceUpdate: func(context.Context, client.Client, string, client.Object, ...client.SubResourceUpdateOption) error {
			return apierrors.NewServiceUnavailable("status API unavailable")
		},
	})
	receiver := &AgentTriggerReceiver{Client: failingClient, Scheme: scheme}
	request := httptest.NewRequest(http.MethodPost, "/v1/triggers/test/repository", nil)
	request.Header.Set("Authorization", "Bearer correct-token")
	request.Header.Set("Idempotency-Key", "event-status-failure")
	response := httptest.NewRecorder()

	receiver.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError || !strings.Contains(response.Body.String(), "trigger-status-update-failed") {
		t.Fatalf("response = %d %s, want status-update failure", response.Code, response.Body.String())
	}
}

func TestAgentTriggerReceiver_AuthenticatedEscalationApproval(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "needs-approval", Namespace: "test"},
		Spec:       unattendedRunSpec(),
		Status: agentv1alpha1.AgentRunStatus{Escalation: &agentv1alpha1.AgentEscalationStatus{
			Recipient:         "human:on-call",
			SubjectRevision:   "spec-sha256:exact",
			RequestedAt:       metav1.NewTime(now),
			ExpiresAt:         metav1.NewTime(now.Add(time.Minute)),
			SafeDefault:       agentv1alpha1.EscalationSafeDefaultPause,
			Outcome:           agentv1alpha1.EscalationOutcomePending,
			DeliveryReference: "pagerduty:incident-1",
		}},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "escalation-token", Namespace: "test"},
		Data:       map[string][]byte{"signal": []byte("signal-token"), "response": []byte("response-token")},
	}
	scheme := testutil.NewScheme()
	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, secret).
		Build()
	receiver := &AgentTriggerReceiver{Client: kubeClient, Scheme: scheme, Now: func() time.Time { return now }}
	payload, err := json.Marshal(map[string]string{
		"decision": "approve", "subjectRevision": "spec-sha256:exact",
		"actor": "human:reviewer", "responseReference": "pagerduty:response-1",
	})
	if err != nil {
		t.Fatalf("encode response: %v", err)
	}

	unauthorized := httptest.NewRecorder()
	receiver.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/v1/escalations/test/needs-approval", bytes.NewReader(payload)))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	authorizedRequest := httptest.NewRequest(http.MethodPost, "/v1/escalations/test/needs-approval", bytes.NewReader(payload))
	authorizedRequest.Header.Set("Authorization", "Bearer response-token")
	authorized := httptest.NewRecorder()
	receiver.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusAccepted {
		t.Fatalf("authorized response = %d %s", authorized.Code, authorized.Body.String())
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Escalation.Outcome != agentv1alpha1.EscalationOutcomeApproved ||
		updated.Status.Escalation.Responder != "human:reviewer" || updated.Status.Escalation.RespondedAt == nil {
		t.Fatalf("escalation = %#v", updated.Status.Escalation)
	}

	wrongStatusCredential := httptest.NewRequest(http.MethodGet, "/v1/escalations/test/needs-approval", nil)
	wrongStatusCredential.Header.Set("Authorization", "Bearer response-token")
	wrongStatusResponse := httptest.NewRecorder()
	receiver.ServeHTTP(wrongStatusResponse, wrongStatusCredential)
	if wrongStatusResponse.Code != http.StatusUnauthorized {
		t.Fatalf("response credential read status = %d, want unauthorized", wrongStatusResponse.Code)
	}

	statusRequest := httptest.NewRequest(http.MethodGet, "/v1/escalations/test/needs-approval", nil)
	statusRequest.Header.Set("Authorization", "Bearer signal-token")
	statusResponse := httptest.NewRecorder()
	receiver.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK || !strings.Contains(statusResponse.Body.String(), `"outcome":"Approved"`) {
		t.Fatalf("runtime escalation status = %d %s", statusResponse.Code, statusResponse.Body.String())
	}
}

func TestAgentTriggerReceiver_AuthenticatedBlockedSignal(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "running-agent", Namespace: "test"},
		Spec:       unattendedRunSpec(),
	}
	run.Status.JobName = "running-agent"
	run.Status.Journal = runJournal(run)
	subjectRevision, err := governancecontracts.AgentRunSubjectRevision(run)
	if err != nil {
		t.Fatalf("digest run: %v", err)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "escalation-token", Namespace: "test"},
		Data:       map[string][]byte{"signal": []byte("signal-token"), "response": []byte("response-token")},
	}
	scheme := testutil.NewScheme()
	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, secret).
		Build()
	receiver := &AgentTriggerReceiver{Client: kubeClient, Scheme: scheme, Now: func() time.Time { return now }}
	payload, err := json.Marshal(map[string]string{
		"subjectRevision":   subjectRevision,
		"reporter":          "agent-harness:claude",
		"reasonCode":        "accountable-decision-required",
		"evidenceReference": "evidence://blocked#sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
	})
	if err != nil {
		t.Fatalf("encode blocked signal: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/blocks/test/running-agent", bytes.NewReader(payload))
	request.Header.Set("Authorization", "Bearer signal-token")
	response := httptest.NewRecorder()

	receiver.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("blocked response = %d %s", response.Code, response.Body.String())
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Blocked == nil || updated.Status.Blocked.SubjectRevision != subjectRevision ||
		updated.Status.Blocked.Reporter != "agent-harness:claude" {
		t.Fatalf("blocked status = %#v", updated.Status.Blocked)
	}
}

func TestAgentRunReconciler_DeliversEscalationBeforeWaiting(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "needs-human", Namespace: "test"},
		Spec:       unattendedRunSpec(),
	}
	markRunAdmissionEvidence(t, run)
	run.Status.Journal = runJournal(run)
	markRunBlocked(t, run)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "escalation-token", Namespace: "test"},
		Data:       map[string][]byte{"signal": []byte("signal-token"), "response": []byte("response-token")},
	}
	kubeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, secret).
		Build()
	reconciler := &AgentRunReconciler{
		Client: kubeClient, Now: func() time.Time { return now },
		EscalationRouter: &stubEscalationRouter{deliveryReference: "pagerduty:incident-1"},
	}

	if _, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(run)}); err != nil {
		t.Fatalf("reconcile escalation: %v", err)
	}
	var updated agentv1alpha1.AgentRun
	if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if updated.Status.Escalation == nil || updated.Status.Escalation.Outcome != agentv1alpha1.EscalationOutcomePending ||
		updated.Status.Escalation.DeliveryReference != "pagerduty:incident-1" {
		t.Fatalf("escalation = %#v", updated.Status.Escalation)
	}
	if updated.Status.JobName != "" {
		t.Fatalf("execution advanced before response: %q", updated.Status.JobName)
	}
}

func TestAgentRunReconciler_RecordsResolvedExecutionInsteadOfTrustingDeclaredProvenance(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	runtimeClassName := "kata"
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "observed-provenance", Namespace: "test"},
		Spec:       unattendedRunSpec(),
	}
	run.Status.Journal = runJournal(run)
	run.Status.Journal.ExecutionImage = ""
	run.Status.Journal.RuntimeClassName = ""
	run.Status.Journal.AgentType = ""
	run.Status.Journal.ProviderEnvironmentKeys = nil
	run.Status.Journal.ExecutionSpecDigest = ""
	job := &batchv1.Job{Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
		RuntimeClassName: &runtimeClassName,
		Containers: []corev1.Container{{
			Name: "agent", Image: "ghcr.io/example/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Command: []string{"claude"}, Args: []string{"--print", "private prompt"},
			Env: []corev1.EnvVar{{Name: "ANTHROPIC_API_KEY", Value: "secret-value"}},
		}},
	}}}}
	execution := &resolvedRunExecution{
		agentType:   agentv1alpha1.AgentTypeClaude,
		providerEnv: map[string]string{"ANTHROPIC_API_KEY": "secret-value"},
		image:       "ghcr.io/example/agent@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	reconciler := &AgentRunReconciler{Now: func() time.Time { return now }}

	observed, err := reconciler.observeResolvedExecution(run, execution, job)
	if err != nil || !observed {
		t.Fatalf("observeResolvedExecution() = %v, %v", observed, err)
	}
	if run.Status.Journal.ExecutionSpecDigest == "" ||
		!slices.Equal(run.Status.Journal.ProviderEnvironmentKeys, []string{"ANTHROPIC_API_KEY"}) {
		t.Fatalf("journal = %#v", run.Status.Journal)
	}
	if strings.Contains(run.Status.Journal.ExecutionSpecDigest, "secret-value") || strings.Contains(run.Status.Journal.ExecutionSpecDigest, "private prompt") {
		t.Fatalf("execution digest leaked input content: %q", run.Status.Journal.ExecutionSpecDigest)
	}

	job.Spec.Template.Spec.Containers[0].Image = "ghcr.io/example/agent@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	if _, err := reconciler.observeResolvedExecution(run, execution, job); err == nil {
		t.Fatal("changed resolved execution did not cause provenance drift")
	}
}

func TestWriteTriggerResponse_EncodingFailureReturnsInternalError(t *testing.T) {
	response := httptest.NewRecorder()

	writeTriggerResponse(context.Background(), response, http.StatusCreated, make(chan int))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
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
	markRunAdmissionEvidence(t, run)
	run.Status.Journal = runJournal(run)
	markRunBlocked(t, run)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "escalation-token", Namespace: "test"},
		Data:       map[string][]byte{"signal": []byte("signal-token"), "response": []byte("response-token")},
	}
	kubeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, secret).
		Build()
	reconciler := &AgentRunReconciler{
		Client: kubeClient, Now: func() time.Time { return now },
		EscalationRouter: &stubEscalationRouter{deliveryReference: "pagerduty:incident-1"},
	}

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

func TestAgentRunReconciler_ApprovedEscalationResumesExactRun(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "approved-run", Namespace: "test"},
		Spec:       unattendedRunSpec(),
	}
	markRunAdmissionEvidence(t, run)
	run.Status.Journal = runJournal(run)
	markRunBlocked(t, run)
	run.Status.Escalation = &agentv1alpha1.AgentEscalationStatus{
		Recipient:         "human:on-call",
		SubjectRevision:   run.Status.Blocked.SubjectRevision,
		RequestedAt:       metav1.NewTime(now.Add(-time.Minute)),
		ExpiresAt:         metav1.NewTime(now.Add(time.Minute)),
		SafeDefault:       agentv1alpha1.EscalationSafeDefaultPause,
		Outcome:           agentv1alpha1.EscalationOutcomeApproved,
		DeliveryReference: "pagerduty:incident-1",
		Responder:         "human:reviewer",
		ResponseReference: "pagerduty:response-1",
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "escalation-token", Namespace: "test"},
		Data:       map[string][]byte{"signal": []byte("signal-token"), "response": []byte("response-token")},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(secret).Build()
	reconciler := &AgentRunReconciler{
		Client: kubeClient, Now: func() time.Time { return now },
		EscalationRouter: &stubEscalationRouter{deliveryReference: "pagerduty:incident-1"},
	}

	_, handled, err := reconciler.reconcileOversight(context.Background(), run)

	if err != nil || handled {
		t.Fatalf("reconcileOversight() handled/error = %v/%v, want resume", handled, err)
	}
}

func TestAgentRunReconciler_RejectsEqualBlockedAndResponseCredentials(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "equal-credentials", Namespace: "test"},
		Spec:       unattendedRunSpec(),
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "escalation-token", Namespace: "test"},
		Data:       map[string][]byte{"signal": []byte("same-token"), "response": []byte("same-token")},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(secret).Build()
	reconciler := &AgentRunReconciler{
		Client:           kubeClient,
		EscalationRouter: &stubEscalationRouter{deliveryReference: "pagerduty:incident-1"},
	}

	if reconciler.observedEscalationRouteHealthy(context.Background(), run) {
		t.Fatal("equal blocked-signal and response credentials were treated as healthy")
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
				Recipient:                   "human:on-call",
				Window:                      metav1.Duration{Duration: time.Minute},
				SafeDefault:                 agentv1alpha1.EscalationSafeDefaultPause,
				BlockedSignalTokenSecretRef: agentv1alpha1.SecretKeyRef{Name: "escalation-token", Key: "signal"},
				ResponseTokenSecretRef:      agentv1alpha1.SecretKeyRef{Name: "escalation-token", Key: "response"},
			},
		},
	}
}
