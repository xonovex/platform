package shared

import (
	"errors"
	"fmt"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

// stubHarness is an in-memory harness command builder: the binary is the agent
// type, and the args are the provider args followed by the prompt.
type stubHarness struct{ agent agentv1alpha1.AgentType }

func (h stubHarness) Command(run *agentv1alpha1.AgentRun, providerCliArgs []string) ([]string, []string, error) {
	args := append([]string{}, providerCliArgs...)
	if run.Spec.Prompt != "" {
		args = append(args, "--prompt", run.Spec.Prompt)
	}
	return []string{string(h.agent)}, args, nil
}

// failingHarness is a harness command builder that always fails.
type failingHarness struct{}

func (failingHarness) Command(*agentv1alpha1.AgentRun, []string) ([]string, []string, error) {
	return nil, nil, errors.New("harness failed")
}

// stubVCS is an in-memory VCS strategy whose scripts and init-container name
// carry the workspace type.
type stubVCS struct{ workspace agentv1alpha1.WorkspaceType }

func (v stubVCS) PostCloneScript() string { return "post-clone " + string(v.workspace) + "\n" }

func (v stubVCS) WorktreeScript(path, branch, sourceBranch string) string {
	return fmt.Sprintf("%s worktree %s %s %s\n", v.workspace, path, branch, sourceBranch)
}

func (v stubVCS) InitContainerName() string { return string(v.workspace) + "-worktree" }

func mustBuildJob(t testing.TB, run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, pvcName, image string, timeout time.Duration, agentType agentv1alpha1.AgentType, workspaceType agentv1alpha1.WorkspaceType, ttl *int32, workspace *WorkspaceBinding) *batchv1.Job {
	t.Helper()
	job, err := BuildJob(run, RunJob{
		ProviderEnv: providerEnv,
		PVCName:     pvcName,
		Image:       image,
		Timeout:     timeout,
		AgentType:   agentType,
		Harness:     stubHarness{agent: agentType},
		VCS:         stubVCS{workspace: workspaceType},
		TTL:         ttl,
		Workspace:   workspace,
	}, DefaultJobPolicy())
	if err != nil {
		t.Fatalf("build Job: %v", err)
	}
	return job
}

func mustBuildWorkspaceInitJob(t testing.TB, workspace *agentv1alpha1.AgentWorkspace, pvcName, image string, runtimeClassName *string) *batchv1.Job {
	t.Helper()
	if runtimeClassName == nil {
		defaultRuntimeClassName := "kata"
		runtimeClassName = &defaultRuntimeClassName
	}
	job, err := BuildWorkspaceInitJob(workspace, pvcName, image, runtimeClassName, stubVCS{workspace: workspace.Spec.Type}, DefaultJobPolicy())
	if err != nil {
		t.Fatalf("build workspace init Job: %v", err)
	}
	return job
}

func mustBuildInitContainers(t testing.TB, run *agentv1alpha1.AgentRun, image string, workspaceType agentv1alpha1.WorkspaceType, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	return buildInitContainers(run, image, stubVCS{workspace: workspaceType}, securityContext)
}

func mustBuildMainContainers(t testing.TB, run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, image string, agentType agentv1alpha1.AgentType, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	containers, err := buildMainContainers(run, providerEnv, nil, image, stubHarness{agent: agentType}, securityContext)
	if err != nil {
		t.Fatalf("build main containers: %v", err)
	}
	return containers
}

func mustBuildWorktreeInitContainers(t testing.TB, run *agentv1alpha1.AgentRun, image string, workspaceType agentv1alpha1.WorkspaceType, worktreeBranch, sourceBranch string, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	return buildWorktreeInitContainers(run, image, stubVCS{workspace: workspaceType}, worktreeBranch, sourceBranch, securityContext)
}

func mustBuildWorkspaceMainContainers(t testing.TB, run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, image string, agentType agentv1alpha1.AgentType, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, securityContext *corev1.SecurityContext) []corev1.Container {
	t.Helper()
	containers, err := buildWorkspaceMainContainers(run, providerEnv, nil, image, stubHarness{agent: agentType}, sharedVolumes, sharedVolumePVCs, securityContext)
	if err != nil {
		t.Fatalf("build workspace main containers: %v", err)
	}
	return containers
}
