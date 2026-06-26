package types

// SandboxMethod represents the sandbox execution method
type SandboxMethod string

const (
	SandboxNone     SandboxMethod = "none"
	SandboxBwrap    SandboxMethod = "bwrap"
	SandboxDocker   SandboxMethod = "docker"
	SandboxCompose  SandboxMethod = "compose"
	SandboxNix      SandboxMethod = "nix"
	SandboxNixFlake SandboxMethod = "nixflake"
)

// SandboxPolicy expresses the isolation guarantees the caller demands of the
// selected tier, independent of which method was requested.
type SandboxPolicy struct {
	// RequirePinnedToolchain mandates a tier whose toolchain comes entirely from
	// a pinned source (a flake.lock devShell or a nix-built closure) with no host
	// /usr,/lib,/bin bound and no host PATH appended. Tiers that leak host system
	// directories (none, bwrap) are rejected at selection time.
	RequirePinnedToolchain bool
}

// SandboxConfig holds sandbox configuration
type SandboxConfig struct {
	AgentID             string
	Method              SandboxMethod
	Policy              SandboxPolicy
	Agent               *AgentConfig
	HomeDir             string
	Image               string
	ComposeFile         string
	Service             string
	WorkDir             string
	RepoDir             string
	Network             bool
	BindPaths           []string
	RoBindPaths         []string
	CustomEnv           []string
	Provider            *ModelProvider
	AgentArgs           []string
	SandboxInitCommands []string
	Verbose             bool
	Debug               bool
	DryRun              bool
}

// SandboxExecutor defines the interface for sandbox implementations
type SandboxExecutor interface {
	IsAvailable() (bool, error)
	Execute(config *SandboxConfig) (int, error)
	GetCommand(config *SandboxConfig) []string
}
