//go:build integration

package integration

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

func requireAdmissionRejection(t *testing.T, objectError error, expected string) {
	t.Helper()
	if objectError == nil {
		t.Fatalf("admission accepted invalid object; expected error containing %q", expected)
	}
	if !strings.Contains(objectError.Error(), expected) {
		t.Fatalf("admission error = %q, want substring %q", objectError, expected)
	}
}

func TestAdmissionRejectsInvalidAgentRun(t *testing.T) {
	ns := createNamespace(t, "admission-run")
	run := testutil.NewAgentRun(ns, "invalid-run", testutil.WithRuntimeClassName("runc"))

	err := k8sClient.Create(ctx, run)

	requireAdmissionRejection(t, err, "is not allowed by the namespace AgentPolicy")
}

func TestAdmissionRejectsInvalidAgentHarness(t *testing.T) {
	ns := createNamespace(t, "admission-harness")
	harness := testutil.NewAgentHarness(ns, "invalid-harness", testutil.WithHarnessType("unknown"))

	err := k8sClient.Create(ctx, harness)

	requireAdmissionRejection(t, err, "invalid agent type")
}

func TestAdmissionRejectsInvalidAgentProvider(t *testing.T) {
	ns := createNamespace(t, "admission-provider")
	provider := testutil.NewAgentProvider(ns, "invalid-provider", testutil.WithEnvironment(map[string]string{
		"NODE_OPTIONS": "--require=/tmp/inject.js",
	}))

	err := k8sClient.Create(ctx, provider)

	requireAdmissionRejection(t, err, "blocked prefix")
}

func TestAdmissionRejectsInvalidAgentToolchain(t *testing.T) {
	ns := createNamespace(t, "admission-toolchain")
	toolchain := &agentv1alpha1.AgentToolchain{
		ObjectMeta: metav1.ObjectMeta{Name: "invalid-toolchain", Namespace: ns},
		Spec:       agentv1alpha1.ToolchainSpec{Type: "unknown"},
	}

	err := k8sClient.Create(ctx, toolchain)

	requireAdmissionRejection(t, err, "invalid toolchain type")
}

func TestAdmissionRejectsInvalidAgentWorkspace(t *testing.T) {
	ns := createNamespace(t, "admission-workspace")
	workspace := testutil.NewAgentWorkspace(ns, "invalid-workspace", testutil.WithWorkspaceType("unknown"))

	err := k8sClient.Create(ctx, workspace)

	requireAdmissionRejection(t, err, "invalid workspace type")
}

func TestAdmissionDefaultsAgentWorkspaceStorage(t *testing.T) {
	ns := createNamespace(t, "admission-defaults")
	workspace := testutil.NewAgentWorkspace(ns, "defaulted-workspace", testutil.WithSharedVolumes(
		agentv1alpha1.SharedVolumeSpec{Name: "config", MountPath: "/config"},
	))

	err := k8sClient.Create(ctx, workspace, client.DryRunAll)
	if err != nil {
		t.Fatalf("create AgentWorkspace: %v", err)
	}

	if workspace.Spec.StorageSize != "10Gi" {
		t.Errorf("storageSize = %q, want 10Gi", workspace.Spec.StorageSize)
	}
	if workspace.Spec.SharedVolumes[0].StorageSize != "1Gi" {
		t.Errorf("shared volume storageSize = %q, want 1Gi", workspace.Spec.SharedVolumes[0].StorageSize)
	}
}
