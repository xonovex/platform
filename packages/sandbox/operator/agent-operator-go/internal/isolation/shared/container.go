package shared

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	harnessshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/harness/shared"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// buildEnvVars builds environment variables for the agent container.
func buildEnvVars(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar) []corev1.EnvVar {
	envVars := append([]corev1.EnvVar{}, providerEnv...)

	// Add spec environment variables (these override provider env).
	envVars = append(envVars, run.Spec.Env...)

	return envVars
}

// buildAgentCommand builds the agent command with the harness command builder.
func buildAgentCommand(run *agentv1alpha1.AgentRun, harness harnessshared.CommandBuilder, providerCliArgs []string) ([]string, []string, error) {
	if harness == nil {
		return nil, nil, fmt.Errorf("build agent command: no harness command builder")
	}
	command, args, err := harness.Command(run, providerCliArgs)
	if err != nil {
		return nil, nil, fmt.Errorf("build agent command: %w", err)
	}
	return command, args, nil
}

// buildInitContainers builds init containers for standalone runs (clone into the
// workspace PVC). A nil Workspace yields an empty repo (the webhook requires one
// for standalone runs); this guards the reconciler against a panic if reached.
func buildInitContainers(run *agentv1alpha1.AgentRun, image string, strategy wsshared.VCSStrategy, sc *corev1.SecurityContext) []corev1.Container {
	var repo agentv1alpha1.RepositorySpec
	if run.Spec.Workspace != nil {
		repo = run.Spec.Workspace.Repository
	}

	volumeMounts := []corev1.VolumeMount{
		{Name: wsshared.WorkspaceVolumeName, MountPath: wsshared.WorkspaceMountPath},
	}
	if repo.CredentialsSecretRef != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      repositoryCredentialsVolumeName,
			MountPath: wsshared.RepositoryCredentialsMountPath,
			ReadOnly:  true,
		})
	}

	return []corev1.Container{
		{
			Name:            "git-clone",
			Image:           image,
			Command:         []string{"sh"},
			Args:            []string{"-c", wsshared.CloneScript(repo, strategy)},
			VolumeMounts:    volumeMounts,
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}
}

// buildMainContainers builds the main agent container for standalone runs.
func buildMainContainers(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, providerCliArgs []string, image string, harness harnessshared.CommandBuilder, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	env := buildEnvVars(run, providerEnv)
	command, args, err := buildAgentCommand(run, harness, providerCliArgs)
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

// buildWorktreeInitContainers builds the init container that creates the per-run
// worktree of a workspace-based AgentRun with the VCS strategy.
func buildWorktreeInitContainers(run *agentv1alpha1.AgentRun, image string, strategy wsshared.VCSStrategy, worktreeBranch, sourceBranch string, sc *corev1.SecurityContext) []corev1.Container {
	worktreePath := wsshared.WorktreePath(run.Name)
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

// buildWorkspaceMainContainers builds the main agent container for workspace-based
// runs (working dir is the per-run worktree; shared volumes are mounted).
func buildWorkspaceMainContainers(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, providerCliArgs []string, image string, harness harnessshared.CommandBuilder, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	env := buildEnvVars(run, providerEnv)
	command, args, err := buildAgentCommand(run, harness, providerCliArgs)
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
