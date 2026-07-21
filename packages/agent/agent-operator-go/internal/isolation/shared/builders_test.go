package shared

import (
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func mustBuildJob(t testing.TB, run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, pvcName, image string, timeout time.Duration, agentType agentv1alpha1.AgentType, workspaceType agentv1alpha1.WorkspaceType, toolchain *agentv1alpha1.ToolchainSpec, ttl *int32, workspace *WorkspaceBinding) *batchv1.Job {
	t.Helper()
	job, err := BuildJob(run, providerEnv, nil, pvcName, image, timeout, agentType, workspaceType, toolchain, ttl, workspace)
	if err != nil {
		t.Fatalf("build Job: %v", err)
	}
	return job
}

func mustBuildWorkspaceInitJob(t testing.TB, workspace *agentv1alpha1.AgentWorkspace, pvcName, image string, runtimeClassName *string) *batchv1.Job {
	t.Helper()
	job, err := BuildWorkspaceInitJob(workspace, pvcName, image, runtimeClassName)
	if err != nil {
		t.Fatalf("build workspace init Job: %v", err)
	}
	return job
}

func mustBuildInitContainers(t testing.TB, run *agentv1alpha1.AgentRun, image string, workspaceType agentv1alpha1.WorkspaceType, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	containers, err := BuildInitContainers(run, image, workspaceType, securityContext)
	if err != nil {
		t.Fatalf("build init containers: %v", err)
	}
	return containers
}

func mustBuildMainContainers(t testing.TB, run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, image string, agentType agentv1alpha1.AgentType, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	containers, err := BuildMainContainers(run, providerEnv, nil, image, agentType, securityContext)
	if err != nil {
		t.Fatalf("build main containers: %v", err)
	}
	return containers
}

func mustBuildWorktreeInitContainers(t testing.TB, run *agentv1alpha1.AgentRun, image string, workspaceType agentv1alpha1.WorkspaceType, worktreeBranch, sourceBranch string, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	containers, err := BuildWorktreeInitContainers(run, image, workspaceType, worktreeBranch, sourceBranch, securityContext)
	if err != nil {
		t.Fatalf("build worktree init containers: %v", err)
	}
	return containers
}

func mustBuildWorkspaceMainContainers(t testing.TB, run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, image string, agentType agentv1alpha1.AgentType, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	containers, err := BuildWorkspaceMainContainers(run, providerEnv, nil, image, agentType, sharedVolumes, sharedVolumePVCs, securityContext)
	if err != nil {
		t.Fatalf("build workspace main containers: %v", err)
	}
	return containers
}
