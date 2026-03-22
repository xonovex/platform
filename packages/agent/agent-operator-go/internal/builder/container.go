package builder

import (
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const (
	workspaceMountPath = "/workspace"
)

// BuildInitContainers builds init containers for standalone runs
func BuildInitContainers(run *agentv1alpha1.AgentRun, image string, wsType agentv1alpha1.WorkspaceType, tc *agentv1alpha1.ToolchainSpec) []corev1.Container {
	containers := []corev1.Container{
		{
			Name:    "git-clone",
			Image:   image,
			Command: []string{"sh"},
			Args:    []string{"-c", buildCloneScript(run.Spec.Workspace.Repository, wsType)},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workspace",
					MountPath: workspaceMountPath,
				},
			},
		},
	}

	for _, t := range Toolchains(tc) {
		if c := t.InitContainer(); c != nil {
			containers = append(containers, *c)
		}
	}

	return containers
}

// BuildMainContainers builds the main agent container
func BuildMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string, agentType agentv1alpha1.AgentType, tc *agentv1alpha1.ToolchainSpec) []corev1.Container {
	env := BuildEnvVars(run, providerEnv)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "workspace",
			MountPath: workspaceMountPath,
		},
	}

	for _, t := range Toolchains(tc) {
		volumeMounts = append(volumeMounts, t.VolumeMounts()...)
		env = append(env, t.EnvVars()...)
	}

	command, args := buildAgentCommand(run, agentType)

	return []corev1.Container{
		{
			Name:         "agent",
			Image:        image,
			Command:      command,
			Args:         args,
			Env:          env,
			WorkingDir:   workspaceMountPath,
			VolumeMounts: volumeMounts,
		},
	}
}

func buildCloneScript(repo agentv1alpha1.RepositorySpec, wsType agentv1alpha1.WorkspaceType) string {
	script := "set -e\n"
	script += "cd " + workspaceMountPath + "\n"

	script += "git clone"
	if repo.Branch != "" {
		script += " --branch " + shellQuote(repo.Branch)
	}
	script += " --single-branch --depth 1"
	script += " -- " + shellQuote(repo.URL) + " .\n"

	if repo.Commit != "" {
		script += "git fetch origin " + shellQuote(repo.Commit) + "\n"
		script += "git checkout " + shellQuote(repo.Commit) + "\n"
	}

	if vcs, err := GetVCSStrategy(wsType); err == nil {
		script += vcs.PostCloneScript()
	}

	return script
}

func buildAgentCommand(run *agentv1alpha1.AgentRun, agentType agentv1alpha1.AgentType) ([]string, []string) {
	if b, err := GetHarnessCommand(agentType); err == nil {
		return b.Command(run)
	}
	return []string{"claude"}, nil
}
