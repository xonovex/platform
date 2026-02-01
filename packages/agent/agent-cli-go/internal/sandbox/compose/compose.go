package compose

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
	"github.com/xonovex/platform/packages/shared/shared-core-go/pkg/scriptlib"
)

// Default compose file location
const defaultComposeFile = "stacks/ai-agent.yaml"

// Default service name
const defaultService = "ai-agent"

// Bash reserved environment variables that should be filtered
var bashReservedEnvVars = map[string]bool{
	"UID":    true,
	"EUID":   true,
	"GID":    true,
	"GROUPS": true,
}

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

// Executor implements Docker Compose sandbox
type Executor struct{}

// NewExecutor creates a new compose executor
func NewExecutor() *Executor {
	return &Executor{}
}

// IsAvailable checks if docker compose is installed
func (e *Executor) IsAvailable() (bool, error) {
	cmd := exec.Command("docker", "compose", "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	return err == nil, nil
}

// Execute runs the agent via Docker Compose
func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
	// Look for compose file in repoDir (original repo) when using worktree
	composeSearchDir := config.WorkDir
	if config.RepoDir != "" {
		composeSearchDir = config.RepoDir
	}

	composeFile := e.findComposeFile(composeSearchDir, config.ComposeFile)
	if composeFile == "" {
		file := config.ComposeFile
		if file == "" {
			file = defaultComposeFile
		}
		scriptlib.LogError("Docker Compose file not found: " + file)
		return 1, nil
	}

	args := e.buildComposeArgs(config, composeFile)

	// Build environment with bash reserved vars filtered
	env := e.buildComposeEnv(config)

	// Add provider environment
	agentEnv, _ := sandboxutil.BuildProviderEnv(config)
	for k, v := range agentEnv {
		env = append(env, k+"="+v)
	}

	if config.Debug {
		scriptlib.LogDebug(config.Debug, "AGENT_WORK_DIR="+config.WorkDir+" docker "+strings.Join(args, " "))
	}

	if config.Verbose {
		service := config.Service
		if service == "" {
			service = defaultService
		}
		scriptlib.LogInfo("Starting " + config.Agent.DisplayName + " via Docker Compose service: " + service)
		scriptlib.LogInfo("Compose file: " + composeFile)
	}

	// Warn about provider override when using compose with pre-configured services
	if config.Provider != nil && config.Service != "" && strings.Contains(config.Service, "-glm") {
		scriptlib.LogWarning("Provider override may conflict with pre-configured service environment")
	}

	return sandboxutil.SpawnSandbox("docker", args, env, "Docker Compose sandbox", config.Verbose)
}

// GetCommand returns the docker compose command
func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
	composeSearchDir := config.WorkDir
	if config.RepoDir != "" {
		composeSearchDir = config.RepoDir
	}

	composeFile := e.findComposeFile(composeSearchDir, config.ComposeFile)
	if composeFile == "" {
		return []string{"# Error: Compose file not found"}
	}

	repoDir := config.RepoDir
	if repoDir == "" {
		repoDir = config.WorkDir
	}

	args := e.buildComposeArgs(config, composeFile)
	result := []string{
		"AGENT_WORK_DIR=" + config.WorkDir,
		"AGENT_REPO_DIR=" + repoDir,
		"docker",
	}
	result = append(result, args...)
	return result
}

// findComposeFile searches for the compose file, from workDir up to root
func (e *Executor) findComposeFile(workDir string, configFile string) string {
	if configFile != "" {
		absolutePath := configFile
		if !filepath.IsAbs(configFile) {
			absolutePath = filepath.Join(workDir, configFile)
		}
		if _, err := os.Stat(absolutePath); err == nil {
			return absolutePath
		}
		return ""
	}

	// Search from workDir up to find the default compose file
	currentDir := workDir
	for {
		candidatePath := filepath.Join(currentDir, defaultComposeFile)
		if _, err := os.Stat(candidatePath); err == nil {
			return candidatePath
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root
			break
		}
		currentDir = parentDir
	}

	return ""
}

// buildComposeArgs builds docker compose run arguments
func (e *Executor) buildComposeArgs(config *types.SandboxConfig, composeFile string) []string {
	service := config.Service
	if service == "" {
		service = defaultService
	}

	homeDir, _ := os.UserHomeDir()

	args := []string{
		"compose",
		"-f", composeFile,
		"run",
		"--rm",
	}

	// Add sandbox environment variables
	sandboxEnv := map[string]string{
		"HOME":   homeDir,
		"TMPDIR": "/tmp",
		"SHELL":  "/bin/bash",
	}
	customEnv := sandboxutil.ParseCustomEnv(config.CustomEnv)
	for k, v := range customEnv {
		sandboxEnv[k] = v
	}
	for k, v := range sandboxEnv {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add application-level volumes on top of compose file volumes
	// Work directory
	args = append(args, "-v", fmt.Sprintf("%s:%s", config.WorkDir, config.WorkDir))

	// User config bind mounts
	for _, configPath := range userConfigPaths {
		sourcePath := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(sourcePath); err == nil {
			args = append(args, "-v", fmt.Sprintf("%s:%s", sourcePath, sourcePath))
		}
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

	// Service name
	args = append(args, service)

	// Agent command (wrapped with init commands if present)
	agentCmd := sandboxutil.BuildAgentCommand(config, "")
	fullCmd := sandboxutil.WrapWithInitCommands(agentCmd, config.SandboxInitCommands)
	args = append(args, fullCmd...)

	return args
}

// buildComposeEnv builds environment variables for docker compose
func (e *Executor) buildComposeEnv(config *types.SandboxConfig) []string {
	env := []string{}

	// Copy process.env but skip read-only bash variables
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]

		// Skip bash reserved vars
		if bashReservedEnvVars[key] {
			continue
		}

		env = append(env, e)
	}

	// Add compose-specific variables
	repoDir := config.RepoDir
	if repoDir == "" {
		repoDir = config.WorkDir
	}

	env = append(env, "AGENT_WORK_DIR="+config.WorkDir)
	env = append(env, "AGENT_REPO_DIR="+repoDir)
	env = append(env, fmt.Sprintf("AGENT_UID=%d", os.Getuid()))
	env = append(env, fmt.Sprintf("AGENT_GID=%d", os.Getgid()))

	// Add agent ID if present
	if config.AgentID != "" {
		env = append(env, "AGENT_ID="+config.AgentID)
	}

	return env
}
