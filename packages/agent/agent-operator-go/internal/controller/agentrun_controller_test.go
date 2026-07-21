package controller

import (
	"context"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/resolver"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestReconcileExecutionResources_RejectsUnownedExistingJob(t *testing.T) {
	run := testutil.NewAgentRun("test", "run")
	run.UID = types.UID("run-uid")
	existing := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: run.Name, Namespace: run.Namespace}}
	fakeClient := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithStatusSubresource(&agentv1alpha1.AgentRun{}).
		WithObjects(run, existing).
		Build()
	reconciler := &AgentRunReconciler{Client: fakeClient, Scheme: testutil.NewScheme()}
	execution := &resolvedRunExecution{
		agentType: agentv1alpha1.AgentTypeClaude,
		defaults: resolver.ResolvedDefaults{
			Timeout:       time.Hour,
			NetworkPolicy: &agentv1alpha1.AgentNetworkPolicy{Disabled: true},
		},
		image: run.Spec.Image,
	}

	_, err := reconciler.reconcileExecutionResources(
		context.Background(), run, execution, "workspace-pvc", agentv1alpha1.WorkspaceTypeGit, nil,
	)

	if err == nil || !strings.Contains(err.Error(), "not controlled by AgentRun") {
		t.Fatalf("reconcileExecutionResources() error = %v, want ownership error", err)
	}
}

func TestAgentRunReconciler_ResolutionFailureMarksRunFailed(t *testing.T) {
	tests := []struct {
		name            string
		workspace       bool
		configureRun    func(*agentv1alpha1.AgentRun)
		wantMessagePart string
	}{
		{
			name: "missing standalone harness",
			configureRun: func(run *agentv1alpha1.AgentRun) {
				run.Spec.HarnessRef = "missing-harness"
			},
			wantMessagePart: "HarnessResolutionFailed",
		},
		{
			name: "missing standalone toolchain",
			configureRun: func(run *agentv1alpha1.AgentRun) {
				run.Spec.ToolchainRef = "missing-toolchain"
			},
			wantMessagePart: "ToolchainResolutionFailed",
		},
		{
			name:      "missing workspace harness",
			workspace: true,
			configureRun: func(run *agentv1alpha1.AgentRun) {
				run.Spec.HarnessRef = "missing-harness"
			},
			wantMessagePart: "HarnessResolutionFailed",
		},
		{
			name:      "missing workspace toolchain",
			workspace: true,
			configureRun: func(run *agentv1alpha1.AgentRun) {
				run.Spec.ToolchainRef = "missing-toolchain"
			},
			wantMessagePart: "ToolchainResolutionFailed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			run := testutil.NewAgentRun("test", "run")
			objects := []client.Object{run}
			if test.workspace {
				run.Spec.Workspace = nil
				run.Spec.WorkspaceRef = "workspace"
				objects = append(objects, &agentv1alpha1.AgentWorkspace{
					ObjectMeta: metav1.ObjectMeta{Name: "workspace", Namespace: "test"},
					Spec: agentv1alpha1.AgentWorkspaceSpec{
						Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/example/repo.git"},
					},
					Status: agentv1alpha1.AgentWorkspaceStatus{
						Phase:        agentv1alpha1.AgentWorkspacePhaseReady,
						WorkspacePVC: "workspace-pvc",
					},
				})
			}
			test.configureRun(run)

			fakeClient := fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithStatusSubresource(&agentv1alpha1.AgentRun{}).
				WithObjects(objects...).
				Build()
			reconciler := &AgentRunReconciler{Client: fakeClient}

			if test.workspace {
				_, _ = reconciler.reconcileWithWorkspace(context.Background(), run)
			} else {
				_, _ = reconciler.reconcileStandalone(context.Background(), run)
			}

			var updated agentv1alpha1.AgentRun
			if err := fakeClient.Get(context.Background(), client.ObjectKeyFromObject(run), &updated); err != nil {
				t.Fatalf("get AgentRun: %v", err)
			}
			if updated.Status.Phase != agentv1alpha1.AgentRunPhaseFailed {
				t.Fatalf("phase = %q, want %q", updated.Status.Phase, agentv1alpha1.AgentRunPhaseFailed)
			}
			if len(updated.Status.Conditions) != 1 || !strings.Contains(updated.Status.Conditions[0].Message, test.wantMessagePart) {
				t.Fatalf("conditions = %#v, want one containing %q", updated.Status.Conditions, test.wantMessagePart)
			}
		})
	}
}

func TestReconcilersSkipDeletingResources(t *testing.T) {
	now := metav1.Now()
	metadata := metav1.ObjectMeta{
		Name:              "deleting",
		Namespace:         "test",
		DeletionTimestamp: &now,
		Finalizers:        []string{"test.xonovex.com/hold"},
	}

	tests := []struct {
		name       string
		object     client.Object
		statusType client.Object
		reconcile  func(client.Client) reconcileFunc
	}{
		{
			name:       "agent run",
			object:     &agentv1alpha1.AgentRun{ObjectMeta: metadata},
			statusType: &agentv1alpha1.AgentRun{},
			reconcile: func(c client.Client) reconcileFunc {
				return (&AgentRunReconciler{Client: c}).Reconcile
			},
		},
		{
			name:       "agent workspace",
			object:     &agentv1alpha1.AgentWorkspace{ObjectMeta: metadata},
			statusType: &agentv1alpha1.AgentWorkspace{},
			reconcile: func(c client.Client) reconcileFunc {
				return (&AgentWorkspaceReconciler{Client: c}).Reconcile
			},
		},
		{
			name:       "agent provider",
			object:     &agentv1alpha1.AgentProvider{ObjectMeta: metadata},
			statusType: &agentv1alpha1.AgentProvider{},
			reconcile: func(c client.Client) reconcileFunc {
				return (&AgentProviderReconciler{Client: c}).Reconcile
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(testutil.NewScheme()).
				WithStatusSubresource(test.statusType).
				WithObjects(test.object).
				Build()
			request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(test.object)}
			if _, err := test.reconcile(fakeClient)(context.Background(), request); err != nil {
				t.Fatalf("Reconcile() error = %v", err)
			}

			var pvcs corev1.PersistentVolumeClaimList
			if err := fakeClient.List(context.Background(), &pvcs, client.InNamespace("test")); err != nil {
				t.Fatalf("list PVCs: %v", err)
			}
			var jobs batchv1.JobList
			if err := fakeClient.List(context.Background(), &jobs, client.InNamespace("test")); err != nil {
				t.Fatalf("list Jobs: %v", err)
			}
			if len(pvcs.Items) != 0 || len(jobs.Items) != 0 {
				t.Fatalf("deleting resource created %d PVCs and %d Jobs", len(pvcs.Items), len(jobs.Items))
			}
		})
	}
}

type reconcileFunc func(context.Context, ctrl.Request) (ctrl.Result, error)
