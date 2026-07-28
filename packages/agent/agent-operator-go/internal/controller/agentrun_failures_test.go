package controller

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

// errInjected stands in for any API failure that is neither NotFound nor
// AlreadyExists, so the controller must surface it rather than absorb it.
var errInjected = errors.New("injected API failure")

func newInterceptedRunReconciler(
	funcs interceptor.Funcs,
	objects ...client.Object,
) *AgentRunReconciler {
	scheme := testutil.NewScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(objects...).
		WithInterceptorFuncs(funcs).
		Build()
	return &AgentRunReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Recorder: events.NewFakeRecorder(20),
	}
}

func networkedExecution(run *agentv1alpha1.AgentRun) *resolvedRunExecution {
	return &resolvedRunExecution{
		agentType: agentv1alpha1.AgentTypeClaude,
		defaults:  resolver.ResolvedDefaults{Timeout: time.Hour},
		image:     run.Spec.Image,
	}
}

func TestReconcileExecutionResources_SurfacesANetworkPolicyCreateFailure(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler := newInterceptedRunReconciler(interceptor.Funcs{
		Create: func(
			_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption,
		) error {
			if _, ok := obj.(*networkingv1.NetworkPolicy); ok {
				return errInjected
			}
			return nil
		},
	}, run)

	_, err := reconciler.reconcileExecutionResources(
		context.Background(), run, networkedExecution(run),
		"workspace-pvc", agentv1alpha1.WorkspaceTypeGit, nil,
	)

	if err == nil {
		t.Fatal("reconcileExecutionResources() must surface a NetworkPolicy create failure")
	}
	if !strings.Contains(err.Error(), "create network policy") {
		t.Errorf("error = %v, want it to name the NetworkPolicy create", err)
	}
}

// AlreadyExists sends the controller to a Get so it can verify ownership; that
// Get failing is a distinct error the controller must not swallow.
func TestReconcileExecutionResources_SurfacesANetworkPolicyGetFailure(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler := newInterceptedRunReconciler(interceptor.Funcs{
		Create: func(
			_ context.Context, _ client.WithWatch, obj client.Object, _ ...client.CreateOption,
		) error {
			if _, ok := obj.(*networkingv1.NetworkPolicy); ok {
				return apierrors.NewAlreadyExists(
					schema.GroupResource{Resource: "networkpolicies"}, obj.GetName())
			}
			return nil
		},
		Get: func(
			_ context.Context, _ client.WithWatch, _ client.ObjectKey,
			obj client.Object, _ ...client.GetOption,
		) error {
			if _, ok := obj.(*networkingv1.NetworkPolicy); ok {
				return errInjected
			}
			return nil
		},
	}, run)

	_, err := reconciler.reconcileExecutionResources(
		context.Background(), run, networkedExecution(run),
		"workspace-pvc", agentv1alpha1.WorkspaceTypeGit, nil,
	)

	if err == nil {
		t.Fatal("reconcileExecutionResources() must surface a NetworkPolicy get failure")
	}
	if !strings.Contains(err.Error(), "get existing network policy") {
		t.Errorf("error = %v, want it to name the NetworkPolicy get", err)
	}
}

func TestReconcileExecutionResources_SurfacesAServiceAccountFailure(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler := newInterceptedRunReconciler(interceptor.Funcs{
		Create: func(
			_ context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption,
		) error {
			if _, ok := obj.(*corev1.ServiceAccount); ok {
				return errInjected
			}
			return c.Create(context.Background(), obj, opts...)
		},
	}, run)
	execution := networkedExecution(run)
	execution.defaults.NetworkPolicy = &agentv1alpha1.AgentNetworkPolicy{Disabled: true}

	_, err := reconciler.reconcileExecutionResources(
		context.Background(), run, execution,
		"workspace-pvc", agentv1alpha1.WorkspaceTypeGit, nil,
	)

	if !errors.Is(err, errInjected) {
		t.Errorf("error = %v, want the injected ServiceAccount failure", err)
	}
}

func TestReconcileExecutionResources_SurfacesAJobCreateFailure(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler := newInterceptedRunReconciler(interceptor.Funcs{
		Create: func(
			_ context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption,
		) error {
			if _, ok := obj.(*batchv1.Job); ok {
				return errInjected
			}
			return c.Create(context.Background(), obj, opts...)
		},
	}, run)
	execution := networkedExecution(run)
	execution.defaults.NetworkPolicy = &agentv1alpha1.AgentNetworkPolicy{Disabled: true}

	_, err := reconciler.reconcileExecutionResources(
		context.Background(), run, execution,
		"workspace-pvc", agentv1alpha1.WorkspaceTypeGit, nil,
	)

	if err == nil {
		t.Fatal("reconcileExecutionResources() must surface a Job create failure")
	}
}

func TestObservePodStatus_SurfacesAListFailure(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler := newInterceptedRunReconciler(interceptor.Funcs{
		List: func(
			_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption,
		) error {
			if _, ok := list.(*corev1.PodList); ok {
				return errInjected
			}
			return nil
		},
	}, run)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"}}

	_, err := reconciler.observePodStatus(context.Background(), run, job)

	if err == nil {
		t.Fatal("observePodStatus() must surface a pod list failure")
	}
	if !strings.Contains(err.Error(), "list pods for Job") {
		t.Errorf("error = %v, want it to name the pod list", err)
	}
}

func newAgentPod(name string, created time.Time, exitCode *int32) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			Labels:            map[string]string{"job-name": "run"},
			CreationTimestamp: metav1.NewTime(created),
		},
	}
	if exitCode != nil {
		pod.Status.ContainerStatuses = []corev1.ContainerStatus{{
			Name: "agent",
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{ExitCode: *exitCode},
			},
		}}
	}
	return pod
}

// A retried Job leaves several pods behind, so the newest one is the run's.
func TestObservePodStatus_ObservesTheNewestPod(t *testing.T) {
	base := time.Now().Truncate(time.Second)
	run := testutil.NewAgentRun("default", "run")
	older := newAgentPod("run-older", base, nil)
	newer := newAgentPod("run-newer", base.Add(time.Minute), nil)
	reconciler, _ := newRunReconciler(run, older, newer)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"}}

	changed, err := reconciler.observePodStatus(context.Background(), run, job)

	if err != nil {
		t.Fatalf("observePodStatus() error = %v", err)
	}
	if !changed {
		t.Error("observePodStatus() = false, want it to report the newly observed pod")
	}
	if run.Status.PodName != "run-newer" {
		t.Errorf("pod = %q, want the newest %q", run.Status.PodName, "run-newer")
	}
}

// Identical creation timestamps are common at second granularity, so the name
// breaks the tie and keeps the choice deterministic.
func TestObservePodStatus_BreaksTimestampTiesByName(t *testing.T) {
	created := time.Now().Truncate(time.Second)
	run := testutil.NewAgentRun("default", "run")
	first := newAgentPod("run-aaa", created, nil)
	second := newAgentPod("run-zzz", created, nil)
	reconciler, _ := newRunReconciler(run, first, second)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"}}

	if _, err := reconciler.observePodStatus(context.Background(), run, job); err != nil {
		t.Fatalf("observePodStatus() error = %v", err)
	}

	if run.Status.PodName != "run-zzz" {
		t.Errorf("pod = %q, want the tie broken toward %q", run.Status.PodName, "run-zzz")
	}
}

func TestObservePodStatus_RecordsTheAgentContainerExitCode(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	exitCode := int32(42)
	pod := newAgentPod("run-pod", time.Now(), &exitCode)
	reconciler, _ := newRunReconciler(run, pod)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"}}

	if _, err := reconciler.observePodStatus(context.Background(), run, job); err != nil {
		t.Fatalf("observePodStatus() error = %v", err)
	}

	if run.Status.ExitCode == nil || *run.Status.ExitCode != exitCode {
		t.Errorf("exit code = %v, want %d", run.Status.ExitCode, exitCode)
	}
}

// Re-observing unchanged state must report no change, so the controller does not
// write status on every resync.
func TestObservePodStatus_ReportsNoChangeWhenAlreadyObserved(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	exitCode := int32(0)
	pod := newAgentPod("run-pod", time.Now(), &exitCode)
	reconciler, _ := newRunReconciler(run, pod)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"}}

	if _, err := reconciler.observePodStatus(context.Background(), run, job); err != nil {
		t.Fatalf("first observePodStatus() error = %v", err)
	}
	changed, err := reconciler.observePodStatus(context.Background(), run, job)

	if err != nil {
		t.Fatalf("second observePodStatus() error = %v", err)
	}
	if changed {
		t.Error("observePodStatus() = true on unchanged state, want false")
	}
}

func TestObservePodStatus_IgnoresAJobWithNoName(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler, _ := newRunReconciler(run)

	changed, err := reconciler.observePodStatus(context.Background(), run, &batchv1.Job{})

	if err != nil {
		t.Fatalf("observePodStatus() error = %v", err)
	}
	if changed {
		t.Error("observePodStatus() = true for an unnamed Job, want false")
	}
}

func TestObservePodStatus_ReportsNoChangeWithoutPods(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler, _ := newRunReconciler(run)
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"}}

	changed, err := reconciler.observePodStatus(context.Background(), run, job)

	if err != nil {
		t.Fatalf("observePodStatus() error = %v", err)
	}
	if changed {
		t.Error("observePodStatus() = true with no pods, want false")
	}
}

func TestUpdateStatus_SurfacesANonNotFoundGetFailure(t *testing.T) {
	run := testutil.NewAgentRun("default", "run")
	reconciler := newInterceptedRunReconciler(interceptor.Funcs{
		Get: func(
			_ context.Context, _ client.WithWatch, _ client.ObjectKey,
			_ client.Object, _ ...client.GetOption,
		) error {
			return errInjected
		},
	}, run)

	if err := reconciler.updateStatus(context.Background(), run); !errors.Is(err, errInjected) {
		t.Errorf("updateStatus() error = %v, want the injected failure", err)
	}
}

func TestUpdateStatus_IgnoresADeletedRun(t *testing.T) {
	reconciler, _ := newRunReconciler()
	absent := testutil.NewAgentRun("default", "absent")

	if err := reconciler.updateStatus(context.Background(), absent); err != nil {
		t.Errorf("updateStatus() error = %v, want nil for a deleted run", err)
	}
}
