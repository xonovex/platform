package shared

import (
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
)

// buildEnvVars builds environment variables for the agent container.
func buildEnvVars(run *agentv1alpha1.AgentRun, providerEnv map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	// Add provider environment variables.
	for k, v := range providerEnv {
		envVars = append(envVars, corev1.EnvVar{Name: k, Value: v})
	}

	// Add spec environment variables (these override provider env).
	envVars = append(envVars, run.Spec.Env...)

	return envVars
}

// buildAgentCommand resolves the harness command for the agent type.
func buildAgentCommand(run *agentv1alpha1.AgentRun, agentType agentv1alpha1.AgentType) ([]string, []string) {
	if b, err := plugins.GetHarnessCommand(agentType); err == nil {
		return b.Command(run)
	}
	return []string{"claude"}, nil
}

// BuildInitContainers builds init containers for standalone runs (clone into the
// workspace PVC). A nil Workspace yields an empty repo (the webhook requires one
// for standalone runs); this guards the reconciler against a panic if reached.
func BuildInitContainers(run *agentv1alpha1.AgentRun, image string, wsType agentv1alpha1.WorkspaceType, sc *corev1.SecurityContext) []corev1.Container {
	var repo agentv1alpha1.RepositorySpec
	if run.Spec.Workspace != nil {
		repo = run.Spec.Workspace.Repository
	}
	strategy, _ := plugins.GetVCSStrategy(wsType)

	return []corev1.Container{
		{
			Name:    "git-clone",
			Image:   image,
			Command: []string{"sh"},
			Args:    []string{"-c", wsshared.CloneScript(repo, strategy)},
			VolumeMounts: []corev1.VolumeMount{
				{Name: wsshared.WorkspaceVolumeName, MountPath: wsshared.WorkspaceMountPath},
			},
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}

// BuildMainContainers builds the main agent container for standalone runs.
func BuildMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string, agentType agentv1alpha1.AgentType, sc *corev1.SecurityContext) []corev1.Container {
	env := buildEnvVars(run, providerEnv)
	command, args := buildAgentCommand(run, agentType)

	return []corev1.Container{
		{
			Name:       "agent",
			Image:      image,
			Command:    command,
			Args:       args,
			Env:        env,
			WorkingDir: wsshared.WorkspaceMountPath,
			VolumeMounts: []corev1.VolumeMount{
				{Name: wsshared.WorkspaceVolumeName, MountPath: wsshared.WorkspaceMountPath},
				{Name: "tmp", MountPath: "/tmp"},
			},
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}

// BuildWorktreeInitContainers builds the init container that creates a git
// worktree (or jj workspace) for a workspace-based AgentRun.
func BuildWorktreeInitContainers(run *agentv1alpha1.AgentRun, image string, wsType agentv1alpha1.WorkspaceType, worktreeBranch, sourceBranch string, sc *corev1.SecurityContext) []corev1.Container {
	worktreePath := wsshared.WorktreePath(run.Name)
	strategy, _ := plugins.GetVCSStrategy(wsType)
	script, name := wsshared.WorktreeScriptAndName(strategy, worktreePath, worktreeBranch, sourceBranch)

	return []corev1.Container{
		{
			Name:    name,
			Image:   image,
			Command: []string{"sh"},
			Args:    []string{"-c", script},
			VolumeMounts: []corev1.VolumeMount{
				{Name: wsshared.WorkspaceVolumeName, MountPath: wsshared.WorkspaceMountPath},
			},
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}

// BuildWorkspaceMainContainers builds the main agent container for workspace-based
// runs (working dir is the per-run worktree; shared volumes are mounted).
func BuildWorkspaceMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string, agentType agentv1alpha1.AgentType, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, sc *corev1.SecurityContext) []corev1.Container {
	env := buildEnvVars(run, providerEnv)
	command, args := buildAgentCommand(run, agentType)
	worktreePath := wsshared.WorktreePath(run.Name)

	volumeMounts := []corev1.VolumeMount{
		{Name: wsshared.WorkspaceVolumeName, MountPath: wsshared.WorkspaceMountPath},
	}
	for _, vol := range sharedVolumes {
		if _, ok := sharedVolumePVCs[vol.Name]; ok {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: vol.Name, MountPath: vol.MountPath})
		}
	}
	volumeMounts = append(volumeMounts, corev1.VolumeMount{Name: "tmp", MountPath: "/tmp"})

	return []corev1.Container{
		{
			Name:            "agent",
			Image:           image,
			Command:         command,
			Args:            args,
			Env:             env,
			WorkingDir:      worktreePath,
			VolumeMounts:    volumeMounts,
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}
