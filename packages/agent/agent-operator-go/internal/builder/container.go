package builder

import (
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const workspaceMountPath = "/workspace"

// BuildInitContainers builds the init container for git clone
func BuildInitContainers(run *agentv1alpha1.AgentRun, image string) []corev1.Container {
	cloneArgs := []string{
		"-c",
		buildCloneScript(run),
	}

	return []corev1.Container{
		{
			Name:    "git-clone",
			Image:   image,
			Command: []string{"sh"},
			Args:    cloneArgs,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workspace",
					MountPath: workspaceMountPath,
				},
			},
		},
	}
}

// BuildMainContainers builds the main agent container
func BuildMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string) []corev1.Container {
	env := BuildEnvVars(run, providerEnv)

	command, args := buildAgentCommand(run)

	return []corev1.Container{
		{
			Name:       "agent",
			Image:      image,
			Command:    command,
			Args:       args,
			Env:        env,
			WorkingDir: workspaceMountPath,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workspace",
					MountPath: workspaceMountPath,
				},
			},
		},
	}
}

func buildCloneScript(run *agentv1alpha1.AgentRun) string {
	script := "set -e\n"
	script += "cd " + workspaceMountPath + "\n"

	// Clone the repository
	script += "git clone"
	if run.Spec.Repository.Branch != "" {
		script += " --branch " + run.Spec.Repository.Branch
	}
	script += " --single-branch --depth 1"
	script += " " + run.Spec.Repository.URL + " .\n"

	// Checkout specific commit if specified
	if run.Spec.Repository.Commit != "" {
		script += "git fetch origin " + run.Spec.Repository.Commit + "\n"
		script += "git checkout " + run.Spec.Repository.Commit + "\n"
	}

	// Setup worktree if specified
	if run.Spec.Worktree != nil {
		sourceBranch := run.Spec.Worktree.SourceBranch
		if sourceBranch == "" {
			sourceBranch = "HEAD"
		}
		script += "git worktree add /workspace-wt -b " + run.Spec.Worktree.Branch + " " + sourceBranch + "\n"
	}

	return script
}

func buildAgentCommand(run *agentv1alpha1.AgentRun) ([]string, []string) {
	switch run.Spec.Agent {
	case agentv1alpha1.AgentTypeClaude:
		args := []string{"--permission-mode", "bypassPermissions"}
		if run.Spec.Prompt != "" {
			args = append(args, "--print", "--prompt", run.Spec.Prompt)
		}
		return []string{"claude"}, args
	case agentv1alpha1.AgentTypeOpencode:
		var args []string
		if run.Spec.Provider != nil && len(run.Spec.Provider.CliArgs) > 0 {
			args = append(args, run.Spec.Provider.CliArgs...)
		}
		return []string{"opencode"}, args
	default:
		return []string{"claude"}, nil
	}
}
