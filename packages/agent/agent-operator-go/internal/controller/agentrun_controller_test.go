package controller

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

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
