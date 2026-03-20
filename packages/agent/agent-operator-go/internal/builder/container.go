package builder

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const (
	workspaceMountPath = "/workspace"
	nixVolumeName      = "nix-env"
	nixMountPath       = "/nix"
	nixProfileBinPath  = "/nix/var/nix/profiles/agent/bin"
	defaultNixImage    = "nixos/nix:latest"
)

// BuildInitContainers builds the init containers for git clone and optional Nix package installation
func BuildInitContainers(run *agentv1alpha1.AgentRun, image string) []corev1.Container {
	containers := []corev1.Container{
		{
			Name:    "git-clone",
			Image:   image,
			Command: []string{"sh"},
			Args:    []string{"-c", buildCloneScript(run)},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "workspace",
					MountPath: workspaceMountPath,
				},
			},
		},
	}

	if nix := BuildNixInitContainer(run.Spec.Nix); nix != nil {
		containers = append(containers, *nix)
	}

	return containers
}

// BuildMainContainers builds the main agent container
func BuildMainContainers(run *agentv1alpha1.AgentRun, providerEnv map[string]string, image string) []corev1.Container {
	env := BuildEnvVars(run, providerEnv)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "workspace",
			MountPath: workspaceMountPath,
		},
	}

	if run.Spec.Nix != nil && len(run.Spec.Nix.Packages) > 0 {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      nixVolumeName,
			MountPath: nixMountPath,
		})
		env = append(env, corev1.EnvVar{
			Name:  "PATH",
			Value: nixProfileBinPath + ":/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		})
	}

	command, args := buildAgentCommand(run)

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

	// Initialize jj colocated mode on top of the git checkout
	if run.Spec.VCS == agentv1alpha1.VCSJujutsu {
		script += "jj git init --colocate\n"
	}

	// Setup worktree if specified
	if run.Spec.Worktree != nil {
		sourceBranch := run.Spec.Worktree.SourceBranch
		if sourceBranch == "" {
			sourceBranch = "HEAD"
		}
		if run.Spec.VCS == agentv1alpha1.VCSJujutsu {
			script += "jj workspace add /workspace-wt --revision " + sourceBranch + "\n"
		} else {
			script += "git worktree add /workspace-wt -b " + run.Spec.Worktree.Branch + " " + sourceBranch + "\n"
		}
	}

	return script
}

// BuildNixInitContainer creates the init container that installs Nix packages into a shared volume.
// Returns nil if no Nix packages are configured.
func BuildNixInitContainer(nix *agentv1alpha1.NixSpec) *corev1.Container {
	if nix == nil || len(nix.Packages) == 0 {
		return nil
	}

	nixImage := defaultNixImage
	if nix.Image != "" {
		nixImage = nix.Image
	}

	return &corev1.Container{
		Name:    "nix-env",
		Image:   nixImage,
		Command: []string{"sh"},
		Args:    []string{"-c", buildNixInstallScript(nix.Packages)},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      nixVolumeName,
				MountPath: "/nix-env",
			},
		},
	}
}

func buildNixInstallScript(packages []string) string {
	// Build nixpkgs# prefixed package refs
	var pkgRefs []string
	for _, pkg := range packages {
		pkgRefs = append(pkgRefs, "nixpkgs#"+pkg)
	}

	// The nixos/nix image has /nix pre-populated. We mount the emptyDir at
	// /nix-env (not /nix) so the image's Nix binary works. After installing
	// packages, we copy the full /nix to /nix-env. The main container then
	// mounts /nix-env at /nix so all store paths resolve correctly.
	script := "set -e\n"
	script += "cp -a /nix/. /nix-env/\n"
	script += fmt.Sprintf("nix --extra-experimental-features \"nix-command flakes\" profile install --profile /nix/var/nix/profiles/agent %s\n", strings.Join(pkgRefs, " "))
	script += "cp -a /nix/. /nix-env/\n"
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
