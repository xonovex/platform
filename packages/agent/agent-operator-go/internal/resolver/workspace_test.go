package resolver

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func TestResolveWorkspace_Found(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "default",
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
		},
		Status: agentv1alpha1.AgentWorkspaceStatus{
			Phase:        agentv1alpha1.AgentWorkspacePhaseReady,
			WorkspacePVC: "my-workspace-ws",
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ws).
		Build()

	result, err := ResolveWorkspace(context.Background(), c, "default", "my-workspace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "my-workspace" {
		t.Errorf("expected workspace name my-workspace, got %s", result.Name)
	}
	if result.Status.WorkspacePVC != "my-workspace-ws" {
		t.Errorf("expected workspace PVC my-workspace-ws, got %s", result.Status.WorkspacePVC)
	}
}

func TestResolveWorkspace_NotFound(t *testing.T) {
	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		Build()

	_, err := ResolveWorkspace(context.Background(), c, "default", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent workspace")
	}
}

func TestResolveWorkspace_WrongNamespace(t *testing.T) {
	ws := &agentv1alpha1.AgentWorkspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workspace",
			Namespace: "other-ns",
		},
		Spec: agentv1alpha1.AgentWorkspaceSpec{
			Repository: agentv1alpha1.RepositorySpec{URL: "https://github.com/org/repo.git"},
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(testutil.NewScheme()).
		WithObjects(ws).
		Build()

	_, err := ResolveWorkspace(context.Background(), c, "default", "my-workspace")
	if err == nil {
		t.Fatal("expected error for workspace in wrong namespace")
	}
}
