package shared

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/internal/plugins"
	wsshared "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/workspace/shared"
)

// buildEnvVars builds environment variables for the agent container.
func buildEnvVars(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar) []corev1.EnvVar {
	envVars := append([]corev1.EnvVar{}, providerEnv...)

	// Add spec environment variables (these override provider env).
	envVars = append(envVars, run.Spec.Env...)

	return envVars
}

// buildAgentCommand resolves the harness command for the agent type.
func buildAgentCommand(run *agentv1alpha1.AgentRun, agentType agentv1alpha1.AgentType, providerCliArgs []string) ([]string, []string, error) {
	builder, err := plugins.GetHarnessCommand(agentType)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve harness command for agent type %q: %w", agentType, err)
	}
	command, args := builder.Command(run, providerCliArgs)
	return command, args, nil
}

// BuildInitContainers builds init containers for standalone runs (clone into the
// workspace PVC). A nil Workspace yields an empty repo (the webhook requires one
// for standalone runs); this guards the reconciler against a panic if reached.
func BuildInitContainers(run *agentv1alpha1.AgentRun, image string, wsType agentv1alpha1.WorkspaceType, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	var repo agentv1alpha1.RepositorySpec
	if run.Spec.Workspace != nil {
		repo = run.Spec.Workspace.Repository
	}
	strategy, err := plugins.GetVCSStrategy(wsType)
	if err != nil {
		return nil, fmt.Errorf("resolve VCS strategy for workspace type %q: %w", wsType, err)
	}

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
	}, nil
}

// BuildMainContainers builds the main agent container for standalone runs.
func BuildMainContainers(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, providerCliArgs []string, image string, agentType agentv1alpha1.AgentType, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	env := buildEnvVars(run, providerEnv)
	command, args, err := buildAgentCommand(run, agentType, providerCliArgs)
	if err != nil {
		return nil, err
	}

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
	}, nil
}

// BuildWorktreeInitContainers builds the init container that creates a git
// worktree (or jj workspace) for a workspace-based AgentRun.
func BuildWorktreeInitContainers(run *agentv1alpha1.AgentRun, image string, wsType agentv1alpha1.WorkspaceType, worktreeBranch, sourceBranch string, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	worktreePath := wsshared.WorktreePath(run.Name)
	strategy, err := plugins.GetVCSStrategy(wsType)
	if err != nil {
		return nil, fmt.Errorf("resolve VCS strategy for workspace type %q: %w", wsType, err)
	}
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
	}, nil
}

// BuildWorkspaceMainContainers builds the main agent container for workspace-based
// runs (working dir is the per-run worktree; shared volumes are mounted).
func BuildWorkspaceMainContainers(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, providerCliArgs []string, image string, agentType agentv1alpha1.AgentType, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	env := buildEnvVars(run, providerEnv)
	command, args, err := buildAgentCommand(run, agentType, providerCliArgs)
	if err != nil {
		return nil, err
	}
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
	}, nil
}
