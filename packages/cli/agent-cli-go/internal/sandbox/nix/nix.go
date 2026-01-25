package nix

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/nixenv"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandboxutil"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
	"github.com/xonovex/platform/packages/lib/core-go/pkg/scriptlib"
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

// Executor implements Nix sandbox using nix-build + bubblewrap
type Executor struct{}

// NewExecutor creates a new nix executor
func NewExecutor() *Executor {
	return &Executor{}
}

// IsAvailable checks if nix-build and bwrap are available
func (e *Executor) IsAvailable() (bool, error) {
	_, nixErr := exec.LookPath("nix-build")
	_, bwrapErr := exec.LookPath("bwrap")

	if nixErr != nil {
		scriptlib.LogError("nix-build is not available")
		return false, nil
	}
	if bwrapErr != nil {
		scriptlib.LogError("bubblewrap (bwrap) is not available")
		return false, nil
	}

	// Check /nix/store exists
	if _, err := os.Stat("/nix/store"); os.IsNotExist(err) {
		scriptlib.LogError("/nix/store does not exist")
		return false, nil
	}

	return true, nil
}

// Execute runs the agent in Nix environment with bubblewrap
func (e *Executor) Execute(config *types.SandboxConfig) (int, error) {
	// Parse nix config
	nixConfig := e.parseNixConfig(config.Image)

	// Build EnvSpec
	envSpec := e.buildEnvSpec(nixConfig, config.Agent)

	if config.Verbose {
		scriptlib.LogInfo("Building Nix environment...")
		scriptlib.LogInfo("Packages: " + strings.Join(envSpec.Packages, ", "))
		scriptlib.LogInfo("Nixpkgs pin: " + envSpec.NixpkgsPin)
	}

	// Build the environment
	_, result, err := nixenv.BuildEnv(envSpec, nixenv.BuildOptions{
		Verbose: config.Verbose,
		Debug:   config.Debug,
	})
	if err != nil {
		return 1, err
	}

	if !result.Success {
		scriptlib.LogError("Failed to build Nix environment: " + result.Error)
		return 1, nil
	}

	if config.Verbose {
		scriptlib.LogInfo("Environment ready: " + result.StorePath)
	}

	// Create per-agent directories
	agentDirs := e.ensureAgentDirs(config.AgentID)

	// Build bwrap arguments
	bwrapArgs := e.buildBwrapArgs(config, result.StorePath, agentDirs)

	// Build agent command with /env/bin prefix (wrapped with init commands if present)
	agentCmd := sandboxutil.BuildAgentCommand(config, "/env/bin")
	fullCmd := sandboxutil.WrapWithInitCommands(agentCmd, config.SandboxInitCommands)

	// Add command separator and agent command
	bwrapArgs = append(bwrapArgs, "--")
	bwrapArgs = append(bwrapArgs, fullCmd...)

	if config.Debug {
		scriptlib.LogDebug(config.Debug, "bwrap "+strings.Join(bwrapArgs, " "))
	}

	if config.Verbose {
		scriptlib.LogInfo("Starting sandboxed " + config.Agent.DisplayName + " with Nix + bubblewrap")
	}

	// Build environment
	agentEnv, _ := sandboxutil.BuildProviderEnv(config)
	customEnv := sandboxutil.ParseCustomEnv(config.CustomEnv)
	mergedEnv := sandboxutil.MergeEnvMaps(agentEnv, customEnv)

	env := os.Environ()
	env = append(env, sandboxutil.EnvMapToSlice(mergedEnv)...)

	return sandboxutil.SpawnSandbox("bwrap", bwrapArgs, env, "Nix sandbox", config.Verbose)
}

// GetCommand returns the bwrap command (for display purposes)
func (e *Executor) GetCommand(config *types.SandboxConfig) []string {
	nixConfig := e.parseNixConfig(config.Image)
	envSpec := e.buildEnvSpec(nixConfig, config.Agent)

	return []string{
		"# 1. Build environment:",
		"#    nix-build ~/.local/share/agent-nix/specs/<envId>.nix -o ~/.local/share/agent-nix/envs/<envId>",
		"# 2. Run with bubblewrap:",
		"bwrap --ro-bind /nix/store /nix/store --ro-bind $ENV_OUT /env ...",
		"# Packages: " + strings.Join(envSpec.Packages, ", "),
	}
}

// AgentDirs holds per-agent runtime directories
type AgentDirs struct {
	Root string
	Work string
	Tmp  string
	Home string
}

// ensureAgentDirs creates per-agent directories
func (e *Executor) ensureAgentDirs(agentID string) *AgentDirs {
	if agentID == "" {
		agentID = "default"
	}

	root := filepath.Join(nixenv.GetAgentsDir(), agentID)
	dirs := &AgentDirs{
		Root: root,
		Work: filepath.Join(root, "work"),
		Tmp:  filepath.Join(root, "tmp"),
		Home: filepath.Join(root, "home"),
	}

	_ = os.MkdirAll(dirs.Work, 0755)
	_ = os.MkdirAll(dirs.Tmp, 0755)
	_ = os.MkdirAll(dirs.Home, 0755)

	return dirs
}

// parseNixConfig parses nix config from sandbox config image field
func (e *Executor) parseNixConfig(image string) *nixenv.NixSandboxConfig {
	nixConfig := &nixenv.NixSandboxConfig{}

	if image == "" || !strings.HasPrefix(image, "nix:") {
		return nixConfig
	}

	spec := strings.TrimPrefix(image, "nix:")

	// Try JSON first
	if strings.HasPrefix(spec, "{") {
		var jsonSpec struct {
			Sets       string `json:"sets"`
			Packages   string `json:"packages"`
			NixpkgsPin string `json:"nixpkgs_pin"`
		}

		if err := json.Unmarshal([]byte(spec), &jsonSpec); err == nil {
			packages := []string{}

			// Expand sets if present
			if jsonSpec.Sets != "" {
				packages = append(packages, nixenv.ExpandPackageSets(strings.Split(jsonSpec.Sets, ","))...)
			}

			// Add extra packages
			if jsonSpec.Packages != "" {
				packages = append(packages, strings.Split(jsonSpec.Packages, ",")...)
			}

			if len(packages) > 0 {
				nixConfig.Packages = packages
			}
			if jsonSpec.NixpkgsPin != "" {
				nixConfig.NixpkgsPin = jsonSpec.NixpkgsPin
			}
			return nixConfig
		}
	}

	// Simple key=value parsing
	if strings.HasPrefix(spec, "preset=") {
		// Presets just use defaults
		return nixConfig
	} else if strings.HasPrefix(spec, "sets=") {
		// Expand package sets to actual packages (added to defaults)
		setNames := strings.Split(strings.TrimPrefix(spec, "sets="), ",")
		nixConfig.Packages = nixenv.ExpandPackageSets(setNames)
	} else if strings.HasPrefix(spec, "packages=") {
		nixConfig.Packages = strings.Split(strings.TrimPrefix(spec, "packages="), ",")
		nixConfig.NoDefaults = true
	}

	return nixConfig
}

// buildEnvSpec builds EnvSpec from NixSandboxConfig and agent
func (e *Executor) buildEnvSpec(nixConfig *nixenv.NixSandboxConfig, agent *types.AgentConfig) *nixenv.EnvSpec {
	var packages []string

	if nixConfig.NoDefaults && len(nixConfig.Packages) > 0 {
		packages = nixConfig.Packages
	} else {
		// Start with base packages
		packages = append(packages, nixenv.DefaultBasePackages...)

		// Add agent-specific package if available
		if agent != nil && agent.NixPackage != "" {
			packages = append(packages, agent.NixPackage)
		}

		// Add any extra packages
		if len(nixConfig.Packages) > 0 {
			packages = append(packages, nixConfig.Packages...)
		}
	}

	nixpkgsPin := nixConfig.NixpkgsPin
	if nixpkgsPin == "" {
		nixpkgsPin = nixenv.DefaultNixpkgsPin
	}

	return &nixenv.EnvSpec{
		NixpkgsPin: nixpkgsPin,
		Packages:   packages,
	}
}

// buildBwrapArgs builds bubblewrap arguments for running in the Nix environment
func (e *Executor) buildBwrapArgs(config *types.SandboxConfig, envOutPath string, agentDirs *AgentDirs) []string {
	homeDir, _ := os.UserHomeDir()

	// Ensure the mount point exists in agentDirs.Home for workDir
	e.ensureSandboxMountPoint(agentDirs.Home, config.WorkDir)

	args := []string{
		// Mount /nix/store read-only
		"--ro-bind", "/nix/store", "/nix/store",
		// Mount the environment output to /env
		"--ro-bind", envOutPath, "/env",
		// Mount per-agent writable directories
		"--bind", agentDirs.Work, "/work",
		"--bind", agentDirs.Tmp, "/tmp",
		// Mount agent home to real home path for compatibility
		"--bind", agentDirs.Home, homeDir,
	}

	// User config bind mounts
	for _, configPath := range userConfigPaths {
		sourcePath := filepath.Join(homeDir, configPath)
		if _, err := os.Stat(sourcePath); err == nil {
			args = append(args, "--bind", sourcePath, sourcePath)
		}
	}

	// Work directory bind mount
	args = append(args, "--bind", config.WorkDir, config.WorkDir)

	// Repo directory mount (for worktrees)
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

	// Minimal /proc and /dev
	args = append(args, "--proc", "/proc")
	args = append(args, "--dev", "/dev")

	// Unshare namespaces (avoid --unshare-all to prevent segfaults)
	args = append(args, "--unshare-uts", "--unshare-ipc", "--unshare-pid", "--unshare-cgroup")

	// Network
	if config.Network {
		args = append(args, "--share-net")
	} else {
		args = append(args, "--unshare-net")
	}

	// Build sandbox environment and add as --setenv args
	sandboxEnv := e.getSandboxEnvironment(homeDir, config.CustomEnv)
	for k, v := range sandboxEnv {
		args = append(args, "--setenv", k, v)
	}

	// Create /usr/bin/env symlink for scripts with #!/usr/bin/env shebang
	args = append(args, "--symlink", "/env/bin/env", "/usr/bin/env")

	// SSL certs if they exist
	if _, err := os.Stat("/etc/ssl/certs"); err == nil {
		args = append(args, "--ro-bind", "/etc/ssl/certs", "/etc/ssl/certs")
	}
	if _, err := os.Stat("/etc/resolv.conf"); err == nil {
		args = append(args, "--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf")
	}

	// Working directory
	args = append(args, "--chdir", config.WorkDir)

	// Die with parent
	args = append(args, "--die-with-parent")

	return args
}

// getSandboxEnvironment returns environment variables for inside the sandbox
func (e *Executor) getSandboxEnvironment(home string, customEnv []string) map[string]string {
	env := map[string]string{
		"HOME":              home,
		"TMPDIR":            "/tmp",
		"PATH":              "/env/bin:/usr/bin:/bin",
		"SHELL":             "/env/bin/bash",
		"NIX_SSL_CERT_FILE": "/etc/ssl/certs/ca-certificates.crt",
	}

	// Merge custom env
	for _, e := range customEnv {
		if idx := strings.Index(e, "="); idx > 0 {
			key := e[:idx]
			value := e[idx+1:]
			env[key] = value
		}
	}

	return env
}

// ensureSandboxMountPoint ensures the mount point directory exists in sandboxHome
func (e *Executor) ensureSandboxMountPoint(sandboxHome string, targetPath string) {
	homeDir, _ := os.UserHomeDir()

	// Check if targetPath is under home
	if !strings.HasPrefix(targetPath, homeDir+"/") {
		return
	}

	// Get the relative path from home to the target
	relativePath, err := filepath.Rel(homeDir, targetPath)
	if err != nil || strings.HasPrefix(relativePath, "..") {
		return
	}

	mountPointInSandbox := filepath.Join(sandboxHome, relativePath)
	_ = os.MkdirAll(mountPointInSandbox, 0755)
}
