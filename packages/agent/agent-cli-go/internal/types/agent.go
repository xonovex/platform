package types

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentClaude   AgentType = "claude"
	AgentOpencode AgentType = "opencode"
)

// AgentConfig defines the configuration for an AI agent
type AgentConfig struct {
	Type        AgentType
	DisplayName string
	Binary      string
	NixPackage  string
}

// AgentExecOptions provides options for agent execution
type AgentExecOptions struct {
	Sandbox         bool
	ProviderCliArgs []string
}

// BuildArgsFunc builds CLI arguments for the agent
type BuildArgsFunc func(baseArgs []string, options AgentExecOptions) []string

// BuildEnvFunc builds environment variables for the agent
type BuildEnvFunc func(providerEnv map[string]string) map[string]string
