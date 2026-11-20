package types

// AgentType represents the type of AI agent
type AgentType string

const (
	AgentClaude   AgentType = "claude"
	AgentCodex    AgentType = "codex"
	AgentCopilot  AgentType = "copilot"
	AgentKiro     AgentType = "kiro"
	AgentOpencode AgentType = "opencode"
	AgentPi       AgentType = "pi"
)

// Protocol is the model API wire protocol an agent speaks.
type Protocol string

const (
	// ProtocolNone means the agent declares no protocol: it chooses its own
	// model service or takes any provider.
	ProtocolNone Protocol = ""
	// ProtocolMessages is the Anthropic Messages API.
	ProtocolMessages Protocol = "messages"
	// ProtocolResponses is the OpenAI Responses API.
	ProtocolResponses Protocol = "responses"
)

// ModelAccess is how an agent obtains model access.
type ModelAccess string

const (
	// AccessProvider means a model provider preset supplies the access, through
	// the routes the agent declares.
	AccessProvider ModelAccess = "provider"
	// AccessLogin means the agent uses the login of its own vendor and takes no
	// provider.
	AccessLogin ModelAccess = "native-login"
)

// AgentConfig declares the facts of an AI agent. The planner and the argument
// builders compare these facts and name no agent.
type AgentConfig struct {
	Type        AgentType
	DisplayName string
	Binary      string
	NixPackage  string
	Protocol    Protocol
	Access      ModelAccess
	// ProviderEnv is true when the agent reads the provider environment.
	ProviderEnv bool
	// ProviderArgs is true when the agent takes the provider arguments.
	ProviderArgs bool
	// LaunchArgs are the arguments every launch passes first, such as a subcommand.
	LaunchArgs []string
	// SandboxArgs are the arguments a launch inside an isolation adds, such as
	// the one that turns off the command sandbox or the permission prompts of the
	// agent.
	SandboxArgs []string
	// VersionArgs are the arguments that make the binary print its version.
	VersionArgs []string
}

// AgentExecOptions provides options for agent execution
type AgentExecOptions struct {
	Sandbox         bool
	ProviderCliArgs []string
}
