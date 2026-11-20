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

// AgentConfig declares the facts of an AI agent harness. The planner and the
// argument builders compare these facts and name no agent. The JSON form is the
// harness schema of shared-contract-go.
type AgentConfig struct {
	Type        AgentType `json:"name"`
	DisplayName string    `json:"displayName"`
	// Description says why the facts are as they are.
	Description string      `json:"description,omitempty"`
	Binary      string      `json:"binary"`
	NixPackage  string      `json:"nixPackage"`
	Protocol    Protocol    `json:"protocol,omitempty"`
	Access      ModelAccess `json:"access"`
	// ProviderEnv is true when the agent reads the provider environment.
	ProviderEnv bool `json:"providerEnv,omitempty"`
	// ProviderArgs is true when the agent takes the provider arguments.
	ProviderArgs bool `json:"providerArgs,omitempty"`
	// LaunchArgs are the arguments every launch passes first, such as a subcommand.
	LaunchArgs []string `json:"launchArgs,omitempty"`
	// SandboxArgs are the arguments a launch inside a boundary adds, such as
	// the one that turns off the command sandbox or the permission prompts of the
	// agent.
	SandboxArgs []string `json:"sandboxArgs,omitempty"`
	// VersionArgs are the arguments that make the binary print its version.
	VersionArgs []string `json:"versionArgs"`
	// UserConfigPaths are the smallest set of home-relative paths the agent needs
	// from the user. General credential stores and broad configuration
	// directories need an explicit read-only bind from the caller.
	UserConfigPaths []string `json:"userConfigPaths,omitempty"`
	// ConfigHome is the directory the agent keeps its configuration in, nil
	// when no variable moves it.
	ConfigHome *ConfigHome `json:"configHome,omitempty"`
	// InstructionFile is the project file the agent reads as its instructions.
	InstructionFile string `json:"instructionFile,omitempty"`
	// SkillsDir is the directory below a project or the home directory that the
	// agent reads skills from.
	SkillsDir string       `json:"skillsDir,omitempty"`
	Plugin    PluginLayout `json:"plugin,omitzero"`
	Headless  Headless     `json:"headless,omitzero"`
}

// ConfigHome is the configuration directory of an agent and the variable that
// moves it.
type ConfigHome struct {
	Env string `json:"env"`
	// Dir is the directory below the home directory.
	Dir string `json:"dir"`
}

// PluginLayout is where a plugin of the agent keeps its parts. Manifest and
// SkillsDir are relative to the plugin, Marketplace to the marketplace root.
type PluginLayout struct {
	Manifest    string `json:"manifest,omitempty"`
	Marketplace string `json:"marketplace,omitempty"`
	SkillsDir   string `json:"skillsDir,omitempty"`
}

// Headless declares the arguments of a run that takes one prompt and exits.
type Headless struct {
	// Args make the run non-interactive.
	Args []string `json:"args,omitempty"`
	// HermeticArgs keep the run from reading the configuration, sessions and
	// servers of the machine.
	HermeticArgs []string `json:"hermeticArgs,omitempty"`
	// Output holds the arguments of each output format, by format name.
	Output map[string][]string `json:"output,omitempty"`
	// Options holds named options. OptionValue in an argument stands for the
	// value a run gives the option.
	Options map[string][]string `json:"options,omitempty"`
	// SkillPrompt invokes a skill: SkillPlaceholder stands for the skill and
	// PromptPlaceholder for the request.
	SkillPrompt string `json:"skillPrompt,omitempty"`
}

// The placeholders of the headless facts.
const (
	OptionValue       = "{value}"
	SkillPlaceholder  = "{skill}"
	PromptPlaceholder = "{prompt}"
)

// AgentExecOptions are the arguments one launch adds to the declared facts of
// the agent.
type AgentExecOptions struct {
	// SandboxArgs are the sandbox arguments of the agent when the launch decided
	// they apply, and none otherwise.
	SandboxArgs     []string
	ProviderCliArgs []string
}
