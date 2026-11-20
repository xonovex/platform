package shared

import (
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"

	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	sharedlaunch "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/launch"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	wsshared "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/workspace/shared"
)

// buildEnvVars builds the environment of the agent container: the provider
// environment, the environment of the planned process, then the environment of
// the run, which overrides both.
func buildEnvVars(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, process isoshared.Process) []corev1.EnvVar {
	envVars := append([]corev1.EnvVar{}, providerEnv...)
	for _, entry := range process.Env {
		name, value, _ := strings.Cut(entry, "=")
		envVars = append(envVars, corev1.EnvVar{Name: name, Value: value})
	}
	return append(envVars, run.Spec.Env...)
}

// agentCommand splits the planned argument vector into the command and the
// arguments of the agent container.
func agentCommand(process isoshared.Process) ([]string, []string, error) {
	if len(process.Argv) == 0 {
		return nil, nil, fmt.Errorf("build agent command: the plan holds no agent process")
	}
	var args []string
	if len(process.Argv) > 1 {
		args = slices.Clone(process.Argv[1:])
	}
	return []string{process.Argv[0]}, args, nil
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
func buildMainContainers(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, spec sharedlaunch.Spec, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	process := spec.Run.Process
	command, args, err := agentCommand(process)
	if err != nil {
		return nil, err
	}

	return []corev1.Container{
		{
			Name:       "agent",
			Image:      spec.Run.Image,
			Command:    command,
			Args:       args,
			Env:        buildEnvVars(run, providerEnv, process),
			WorkingDir: process.Dir,
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
// runs (the planned work directory is the per-run worktree; shared volumes are
// mounted).
func buildWorkspaceMainContainers(run *agentv1alpha1.AgentRun, providerEnv []corev1.EnvVar, spec sharedlaunch.Spec, sharedVolumes []agentv1alpha1.SharedVolumeSpec, sharedVolumePVCs map[string]string, sc *corev1.SecurityContext) ([]corev1.Container, error) {
	process := spec.Run.Process
	command, args, err := agentCommand(process)
	if err != nil {
		return nil, err
	}

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
			Image:           spec.Run.Image,
			Command:         command,
			Args:            args,
			Env:             buildEnvVars(run, providerEnv, process),
			WorkingDir:      process.Dir,
			VolumeMounts:    volumeMounts,
			SecurityContext: DefaultContainerSecurityContext(sc),
		},
	}, nil
}
