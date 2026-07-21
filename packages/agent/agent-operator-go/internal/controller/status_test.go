package controller

import (
	"context"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	isoshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestReconcileExecutionResourcesCreatesHardenedWorkload(t *testing.T) {
	ctx := context.Background()
	run := testutil.NewAgentRun("test", "run")
	scheme := testutil.NewScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run).
		Build()
	reconciler := &AgentRunReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: events.NewFakeRecorder(10),
	}
	execution := &resolvedRunExecution{
		agentType: agentv1alpha1.AgentTypeClaude,
		image:     run.Spec.Image,
		defaults: resolver.ResolvedDefaults{
			Image:   run.Spec.Image,
			Timeout: time.Hour,
		},
	}

	_, err := reconciler.reconcileExecutionResources(
		ctx,
		run,
		execution,
		"run-workspace",
		agentv1alpha1.WorkspaceTypeGit,
		nil,
	)

	if err != nil {
		t.Fatalf("reconcileExecutionResources() error = %v", err)
	}
	for _, object := range []client.Object{
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: isoshared.AgentServiceAccountName, Namespace: "test"}},
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "run-netpol", Namespace: "test"}},
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "test"}},
	} {
		if err := fakeClient.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
			t.Fatalf("get %T: %v", object, err)
		}
	}
	var updated agentv1alpha1.AgentRun
	if err := fakeClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get AgentRun: %v", err)
	}
	if updated.Status.JobName != "run" {
		t.Fatalf("JobName = %q, want run", updated.Status.JobName)
	}
}

func TestAgentWorkspaceReconcileCreatesHardenedBootstrapResources(t *testing.T) {
	ctx := context.Background()
	runtimeClassName := "kata"
	workspace := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{Name: "workspace", Namespace: "test", UID: "workspace-uid"},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository:       agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"},
			StorageSize:      "1Gi",
			RuntimeClassName: &runtimeClassName,
		},
	}
	scheme := testutil.NewScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentWorkspace{}).
		WithObjects(workspace).
		Build()
	reconciler := &AgentWorkspaceReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: events.NewFakeRecorder(10),
	}

	_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(workspace)})

	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	for _, object := range []client.Object{
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: isoshared.AgentServiceAccountName, Namespace: "test"}},
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: "workspace-init-netpol", Namespace: "test"}},
		&batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "workspace-init", Namespace: "test"}},
	} {
		if err := fakeClient.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
			t.Fatalf("get %T: %v", object, err)
		}
	}
}

func TestReconcileJobStatusMapsTerminalAndRunningStates(t *testing.T) {
	now := metav1.Now()
	tests := []struct {
		name      string
		configure func(*agentv1alpha1.AgentRun, *batchv1.Job)
		want      agentv1alpha1.AgentRunPhase
	}{
		{
			name: "completed",
			configure: func(_ *agentv1alpha1.AgentRun, job *batchv1.Job) {
				job.Status.Conditions = []batchv1.JobCondition{{Type: batchv1.JobComplete, Status: corev1.ConditionTrue}}
			},
			want: agentv1alpha1.AgentRunPhaseSucceeded,
		},
		{
			name: "failed",
			configure: func(_ *agentv1alpha1.AgentRun, job *batchv1.Job) {
				job.Status.Conditions = []batchv1.JobCondition{{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "failed"}}
			},
			want: agentv1alpha1.AgentRunPhaseFailed,
		},
		{
			name: "started",
			configure: func(_ *agentv1alpha1.AgentRun, job *batchv1.Job) {
				job.Status.Active = 1
			},
			want: agentv1alpha1.AgentRunPhaseRunning,
		},
		{
			name: "timed out",
			configure: func(run *agentv1alpha1.AgentRun, job *batchv1.Job) {
				job.Status.Active = 1
				run.Status.Phase = agentv1alpha1.AgentRunPhaseRunning
				started := metav1.NewTime(now.Add(-2 * time.Second))
				run.Status.StartTime = &started
				run.Spec.Timeout = &metav1.Duration{Duration: time.Second}
			},
			want: agentv1alpha1.AgentRunPhaseTimedOut,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			run := testutil.NewAgentRun("test", "run")
			job := &batchv1.Job{}
			test.configure(run, job)
			fakeClient := fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithStatusSubresource(&agentv1alpha1.AgentRun{}).
				WithObjects(run).
				Build()
			reconciler := &AgentRunReconciler{Client: fakeClient, Recorder: events.NewFakeRecorder(10)}

			_, err := reconciler.reconcileJobStatus(ctx, run, job)

			if err != nil {
				t.Fatalf("reconcileJobStatus() error = %v", err)
			}
			var updated agentv1alpha1.AgentRun
			if err := fakeClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
				t.Fatalf("get AgentRun: %v", err)
			}
			if updated.Status.Phase != test.want {
				t.Fatalf("phase = %q, want %q", updated.Status.Phase, test.want)
			}
		})
	}
}

func TestReconcileJobStatusRecordsPodNameAndAgentExitCode(t *testing.T) {
	ctx := context.Background()
	run := testutil.NewAgentRun("test", "run")
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "test"},
		Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
			Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "failed",
		}}},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "run-pod", Namespace: "test", Labels: map[string]string{"job-name": "run"}},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name: "agent",
			State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{
				ExitCode: 17,
			}},
		}}},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, pod).
		Build()
	reconciler := &AgentRunReconciler{Client: fakeClient, Recorder: events.NewFakeRecorder(10)}

	_, err := reconciler.reconcileJobStatus(ctx, run, job)

	if err != nil {
		t.Fatalf("reconcileJobStatus() error = %v", err)
	}
	var updated agentv1alpha1.AgentRun
	if err := fakeClient.Get(ctx, client.ObjectKeyFromObject(run), &updated); err != nil {
		t.Fatalf("get AgentRun: %v", err)
	}
	if updated.Status.PodName != "run-pod" {
		t.Fatalf("PodName = %q, want run-pod", updated.Status.PodName)
	}
	if updated.Status.ExitCode == nil || *updated.Status.ExitCode != 17 {
		t.Fatalf("ExitCode = %v, want 17", updated.Status.ExitCode)
	}
}

func TestReconcileInitJobStatusMapsTerminalStates(t *testing.T) {
	tests := []struct {
		name      string
		condition batchv1.JobCondition
		want      agentv1alpha1.AgentWorkspacePhase
	}{
		{
			name:      "completed",
			condition: batchv1.JobCondition{Type: batchv1.JobComplete, Status: corev1.ConditionTrue},
			want:      agentv1alpha1.AgentWorkspacePhaseReady,
		},
		{
			name:      "failed",
			condition: batchv1.JobCondition{Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Message: "failed"},
			want:      agentv1alpha1.AgentWorkspacePhaseFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			workspace := &agentv1alpha1.AgentWorkspace{
				ObjectMeta: metav1.ObjectMeta{Name: "workspace", Namespace: "test"},
			}
			job := &batchv1.Job{Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{test.condition},
			}}
			fakeClient := fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithStatusSubresource(&agentv1alpha1.AgentWorkspace{}).
				WithObjects(workspace).
				Build()
			reconciler := &AgentWorkspaceReconciler{Client: fakeClient, Recorder: events.NewFakeRecorder(10)}

			_, err := reconciler.reconcileInitJobStatus(ctx, workspace, job)

			if err != nil {
				t.Fatalf("reconcileInitJobStatus() error = %v", err)
			}
			var updated agentv1alpha1.AgentWorkspace
			if err := fakeClient.Get(ctx, client.ObjectKeyFromObject(workspace), &updated); err != nil {
				t.Fatalf("get AgentWorkspace: %v", err)
			}
			if updated.Status.Phase != test.want {
				t.Fatalf("phase = %q, want %q", updated.Status.Phase, test.want)
			}
		})
	}
}

func TestProviderReconcileReportsSecretReadiness(t *testing.T) {
	tests := []struct {
		name    string
		objects func(*agentv1alpha1.AgentProvider) []client.Object
		ready   bool
	}{
		{
			name: "missing secret",
			objects: func(provider *agentv1alpha1.AgentProvider) []client.Object {
				return []client.Object{provider}
			},
			ready: false,
		},
		{
			name: "resolved secret",
			objects: func(provider *agentv1alpha1.AgentProvider) []client.Object {
				return []client.Object{
					provider,
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{Name: "credentials", Namespace: "test"},
						Data:       map[string][]byte{"token": []byte("secret")},
					},
				}
			},
			ready: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := testutil.NewAgentProvider(
				"test",
				"provider",
				testutil.WithAuthTokenSecretRef("credentials", "token"),
			)
			fakeClient := fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithStatusSubresource(&agentv1alpha1.AgentProvider{}).
				WithObjects(test.objects(provider)...).
				Build()
			reconciler := &AgentProviderReconciler{Client: fakeClient, Recorder: events.NewFakeRecorder(10)}

			_, err := reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(provider)})

			if err != nil {
				t.Fatalf("Reconcile() error = %v", err)
			}
			var updated agentv1alpha1.AgentProvider
			if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(provider), &updated); err != nil {
				t.Fatalf("get AgentProvider: %v", err)
			}
			if updated.Status.Ready != test.ready {
				t.Fatalf("Ready = %v, want %v", updated.Status.Ready, test.ready)
			}
		})
	}
}

func TestProviderReconcileLeavesStableStatusUnchanged(t *testing.T) {
	provider := testutil.NewAgentProvider(
		"test",
		"provider",
		testutil.WithAuthTokenSecretRef("credentials", "token"),
	)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "credentials", Namespace: "test"},
		Data:       map[string][]byte{"token": []byte("secret")},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentProvider{}).
		WithObjects(provider, secret).
		Build()
	reconciler := &AgentProviderReconciler{Client: fakeClient, Recorder: events.NewFakeRecorder(10)}
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(provider)}

	if _, err := reconciler.Reconcile(context.Background(), request); err != nil {
		t.Fatalf("first Reconcile() error = %v", err)
	}
	var first agentv1alpha1.AgentProvider
	if err := fakeClient.Get(context.Background(), request.NamespacedName, &first); err != nil {
		t.Fatalf("get first AgentProvider: %v", err)
	}
	if _, err := reconciler.Reconcile(context.Background(), request); err != nil {
		t.Fatalf("second Reconcile() error = %v", err)
	}
	var second agentv1alpha1.AgentProvider
	if err := fakeClient.Get(context.Background(), request.NamespacedName, &second); err != nil {
		t.Fatalf("get second AgentProvider: %v", err)
	}

	if first.ResourceVersion != second.ResourceVersion {
		t.Errorf("ResourceVersion changed from %q to %q for stable status", first.ResourceVersion, second.ResourceVersion)
	}
	firstReady := meta.FindStatusCondition(first.Status.Conditions, "Ready")
	secondReady := meta.FindStatusCondition(second.Status.Conditions, "Ready")
	if firstReady == nil || secondReady == nil || !firstReady.LastTransitionTime.Equal(&secondReady.LastTransitionTime) {
		t.Errorf("Ready transition time changed for stable status: first=%#v second=%#v", firstReady, secondReady)
	}
}

func TestProvidersForSecretReturnsDependentProviders(t *testing.T) {
	matching := testutil.NewAgentProvider(
		"test",
		"matching",
		testutil.WithAuthTokenSecretRef("credentials", "token"),
	)
	unrelated := testutil.NewAgentProvider(
		"test",
		"unrelated",
		testutil.WithAuthTokenSecretRef("other", "token"),
	)
	fakeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(matching, unrelated).
		Build()
	reconciler := &AgentProviderReconciler{Client: fakeClient}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "credentials", Namespace: "test"}}

	requests := reconciler.providersForSecret(context.Background(), secret)

	if len(requests) != 1 || requests[0].NamespacedName != client.ObjectKeyFromObject(matching) {
		t.Fatalf("providersForSecret() = %v, want matching provider", requests)
	}
}
