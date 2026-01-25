package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/tools/tool-agent-cli-go/internal/types"
)

// Default Docker image
const defaultDockerImage = "node:trixie-slim"

// User config paths that should be bind mounted into sandboxes (relative to home)
var userConfigPaths = []string{
	".claude",
	".claude.json",
	".gitconfig",
	".gitignore_global",
	".ssh",
	".config",
	".npmrc",
	".npm",
	".npm-global",
	".cargo",
	".rustup",
	".local",
	".cache",
}

// Executor implements Docker sandbox
type Executor struct{}

// NewExecutor creates a new docker executor
func NewExecutor() *Executor {
	return &Executor{}
}

// IsAvailable checks if docker is installed and running
func (e *Executor) IsAvailable() (bool, error) {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil, nil
}

// Execute runs the agent in Docker container
func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
	args := e.buildDockerArgs(config)
	env := os.Environ()

	return sandboxutil.SpawnSandbox("docker", args, env, "Docker sandbox", config.Verbose)
}

// GetCommand returns the docker command
func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
	args := e.buildDockerArgs(config)
	result := make([]string, 0, len(args)+1)
	result = append(result, "docker")
	result = append(result, args...)
	return result
}

func (e *Executor) buildDockerArgs(config *types.SandboxConfig) []string {
	args := []string{"run", "--rm", "-it"}

	homeDir, _ := os.UserHomeDir()

	// Network
	if !config.Network {
		args = append(args, "--network", "none")
	}

	// Working directory
	args = append(args, "-w", config.WorkDir)

	// User mapping (run as current user)
	uid := os.Getuid()
	gid := os.Getgid()
	args = append(args, "-u", fmt.Sprintf("%d:%d", uid, gid))

	// Environment variables from provider
	agentEnv, _ := sandboxutil.BuildProviderEnv(config)
	customEnv := sandboxutil.ParseCustomEnv(config.CustomEnv)
	mergedEnv := sandboxutil.MergeEnvMaps(agentEnv, customEnv)

	// Add sandbox environment
	sandboxEnv := map[string]string{
		"HOME":   homeDir,
		"TMPDIR": "/tmp",
		"SHELL":  "/bin/bash",
	}
	mergedEnv = sandboxutil.MergeEnvMaps(sandboxEnv, mergedEnv)

	for k, v := range mergedEnv {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Home directory mount
	args = append(args, "-v", fmt.Sprintf("%s:%s", homeDir, homeDir))

	// User config bind mounts
	for _, configPath := range userConfigPaths {
		sourcePath := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(sourcePath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s", sourcePath, sourcePath))
		}
	}

	// Work directory mount
	args = append(args, "-v", fmt.Sprintf("%s:%s", config.WorkDir, config.WorkDir))

	// Repo directory mount (for worktrees)
	if config.RepoDir != "" && config.RepoDir != config.WorkDir {
		args = append(args, "-v", fmt.Sprintf("%s:%s:ro", config.RepoDir, config.RepoDir))
	}

	// Additional bind mounts
	for _, path := range config.BindPaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s", absPath, absPath))
		}
	}

	for _, path := range config.RoBindPaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s:ro", absPath, absPath))
		}
	}

	// Image
	image := config.Image
	if image == "" {
		image = defaultDockerImage
	}
	args = append(args, image)

	// Agent command (wrapped with init commands if present)
	agentCmd := sandboxutil.BuildAgentCommand(config, "")
	fullCmd := sandboxutil.WrapWithInitCommands(agentCmd, config.SandboxInitCommands)
	args = append(args, fullCmd...)

	return args
}
