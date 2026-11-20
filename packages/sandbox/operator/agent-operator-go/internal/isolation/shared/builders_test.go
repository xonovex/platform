package shared

import (
	"fmt"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	sharedlaunch "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// podDouble is a confinement that bounds the agent like a pod, so the shared
// planner adds the sandbox arguments of the harness.
type podDouble struct{}

func (podDouble) PinnedProvision(provision.ProvisionMethod, bool, string) bool { return true }
func (podDouble) HidesHost(bool, string) bool                                  { return true }
func (podDouble) KernelIsolated(string) bool                                   { return true }
func (podDouble) ConfinesNetwork() bool                                        { return true }
func (podDouble) EgressTransport() network.EgressTransport                     { return network.TransportNone }
func (podDouble) Bounded() bool                                                { return true }
func (podDouble) RequiresImage() bool                                          { return true }
func (podDouble) AcceptsHostBinds() bool                                       { return false }

// imageDouble is a provisioner that leaves every tool to the image.
type imageDouble struct{}

func (imageDouble) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (imageDouble) Pinned() bool { return false }
func (imageDouble) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}

// planFor plans run for agentType in image with the shared planner and the
// shipped harness facts, so a Job carries the argument vector the operator runs.
func planFor(t testing.TB, run *agentv1alpha1.AgentRun, agentType agentv1alpha1.AgentType, image string) sharedlaunch.Spec {
	t.Helper()
	agent, err := agents.GetAgent(types.AgentType(agentType))
	if err != nil {
		t.Fatalf("harness facts: %v", err)
	}
	var args []string
	if run.Spec.Prompt != "" {
		if args, err = agents.PromptArgs(agent, run.Spec.Prompt); err != nil {
			t.Fatalf("prompt arguments: %v", err)
		}
	}
	runtime := ""
	if run.Spec.RuntimeClassName != nil {
		runtime = *run.Spec.RuntimeClassName
	}
	registry := sandbox.NewRegistry[isoshared.Confinement]().
		RegisterIsolator("pod", func() isoshared.Confinement { return podDouble{} }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return imageDouble{} })
	request := sharedlaunch.Request{
		Agent: string(agentType), Isolation: "pod", Provision: string(provision.ProvisionNone), Network: string(network.NetworkNone),
		ExplicitAxes: true, IsolationDockerRuntime: runtime, Image: image, AgentArgs: args,
	}
	workDir := wsshared.WorkDir(run)
	spec, _, err := sharedlaunch.PlanSpec(request, agent, nil, workDir, workDir, AgentHome, registry, false)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	return spec
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
		Spec:        planFor(t, run, agentType, image),
		ProviderEnv: providerEnv,
		PVCName:     pvcName,
		Timeout:     timeout,
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
	containers, err := buildMainContainers(run, providerEnv, planFor(t, run, agentType, image), securityContext)
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
	containers, err := buildWorkspaceMainContainers(run, providerEnv, planFor(t, run, agentType, image), sharedVolumes, sharedVolumePVCs, securityContext)
	if err != nil {
		t.Fatalf("build workspace main containers: %v", err)
	}
	return containers
}
