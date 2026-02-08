package bwrap

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

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

// Executor implements bubblewrap sandbox
type Executor struct{}

// NewExecutor creates a new bwrap executor
func NewExecutor() *Executor {
	return &Executor{}
}

// IsAvailable checks if bwrap is installed
func (e *Executor) IsAvailable() (bool, error) {
	_, err := exec.LookPath("bwrap")
	return err == nil, nil
}

// Execute runs the agent in bubblewrap sandbox
func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
	args := e.buildBwrapArgs(config)

	// Build environment - passed via --setenv in args, not via process env
	// But we still need to pass provider env to the parent process for token access
	agentEnv, err := sandboxutil.BuildProviderEnv(config)
	if err != nil {
		return 1, err
	}

	// Merge with custom env for the parent process
	customEnv := sandboxutil.ParseCustomEnv(config.CustomEnv)
	mergedEnv := sandboxutil.MergeEnvMaps(agentEnv, customEnv)

	// Convert to environ format for parent process
	env := os.Environ()
	env = append(env, sandboxutil.EnvMapToSlice(mergedEnv)...)

	return sandboxutil.SpawnSandbox("bwrap", args, env, "Bubblewrap sandbox", config.Verbose)
}

// GetCommand returns the bwrap command
func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
	args := e.buildBwrapArgs(config)
	result := make([]string, 0, len(args)+1)
	result = append(result, "bwrap")
	result = append(result, args...)
	return result
}

func (e *Executor) buildBwrapArgs(config *types.SandboxConfig) []string {
	// Unshare namespaces individually instead of --unshare-all
	// to avoid user namespace issues that can cause segfaults
	args := []string{
		"--unshare-uts",
		"--unshare-ipc",
		"--unshare-pid",
		"--unshare-cgroup",
		"--die-with-parent",
	}

	// Network
	if config.Network {
		args = append(args, "--share-net")
	}

	homeDir, _ := os.UserHomeDir()
	if config.HomeDir != "" {
		homeDir = config.HomeDir
	}

	// Build sandbox environment and add as --setenv args
	sandboxEnv := e.getSandboxEnvironment(homeDir, config.CustomEnv)
	for k, v := range sandboxEnv {
		args = append(args, "--setenv", k, v)
	}

	// System directories (read-only)
	systemDirs := []string{"/usr", "/lib", "/lib64", "/bin", "/etc"}
	for _, dir := range systemDirs {
		if _, err := os.Stat(dir); err == nil {
			args = append(args, "--ro-bind", dir, dir)
		}
	}

	// Special mounts
	args = append(args,
		"--dev-bind", "/dev", "/dev",
		"--proc", "/proc",
		"--tmpfs", "/tmp",
	)

	// Home directory
	args = append(args, "--bind", homeDir, homeDir)

	// User config bind mounts
	for _, configPath := range userConfigPaths {
		sourcePath := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(sourcePath); err == nil {
			args = append(args, "--bind", sourcePath, sourcePath)
		}
	}

	// Bind work directory
	args = append(args, "--bind", config.WorkDir, config.WorkDir)

	// Bind source repo if worktree
	if config.RepoDir != "" && config.RepoDir != config.WorkDir {
		args = append(args, "--ro-bind", config.RepoDir, config.RepoDir)
	}

	// Additional bind mounts
	for _, path := range config.BindPaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err == nil {
			args = append(args, "--bind", absPath, absPath)
		}
	}

	for _, path := range config.RoBindPaths {
		absPath, _ := filepath.Abs(path)
		if _, err := os.Stat(absPath); err == nil {
			args = append(args, "--ro-bind", absPath, absPath)
		}
	}

	// Set working directory
	args = append(args, "--chdir", config.WorkDir)

	// Separator and agent command (wrapped with init commands if present)
	args = append(args, "--")
	agentCmd := sandboxutil.BuildAgentCommand(config, "")
	fullCmd := sandboxutil.WrapWithInitCommands(agentCmd, config.SandboxInitCommands)
	args = append(args, fullCmd...)

	return args
}

// getSandboxEnvironment returns environment variables for inside the sandbox
func (e *Executor) getSandboxEnvironment(home string, customEnv []string) map[string]string {
	// Build PATH with common locations
	pathParts := []string{
		filepath.Join(home, ".local", "bin"),
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
	}

	// Include host PATH for binaries in non-standard locations
	if hostPath := os.Getenv("PATH"); hostPath != "" {
		pathParts = append(pathParts, hostPath)
	}

	env := map[string]string{
		"HOME":   home,
		"TMPDIR": "/tmp",
		"PATH":   joinPath(pathParts),
		"SHELL":  "/bin/bash",
	}

	// Merge custom env
	for _, e := range customEnv {
		if idx := indexOf(e, "="); idx > 0 {
			key := e[:idx]
			value := e[idx+1:]
			env[key] = value
		}
	}

	return env
}

func joinPath(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ":"
		}
		result += p
	}
	return result
}

func indexOf(s string, substr string) int {
	for i := 0; i < len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
