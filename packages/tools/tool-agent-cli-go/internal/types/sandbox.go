package types

// SandboxMethod represents the sandbox execution method
type SandboxMethod string

const (
	SandboxNone    SandboxMethod = "none"
	SandboxBwrap   SandboxMethod = "bwrap"
	SandboxDocker  SandboxMethod = "docker"
	SandboxCompose SandboxMethod = "compose"
	SandboxNix     SandboxMethod = "nix"
)

// SandboxConfig holds sandbox configuration
type SandboxConfig struct {
	AgentID             string
	Method              SandboxMethod
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
