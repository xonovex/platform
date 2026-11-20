package model

// Effort is how much reasoning a profile asks of the model.
type Effort string

// Profile is one way to run a model: the harness, the model and its limits.
type Profile struct {
	Harness        string `json:"harness"`
	Model          string `json:"model,omitempty"`
	Effort         Effort `json:"effort,omitempty"`
	TimeoutSeconds int    `json:"timeoutSeconds,omitempty"`
}

// TriggerSettings are the protocol settings of the trigger eval.
type TriggerSettings struct {
	Runs      int      `json:"runs,omitempty"`
	BatchSize int      `json:"batchSize,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
}

// RoutingSettings are the protocol settings of the routing eval.
type RoutingSettings struct {
	Runs      int      `json:"runs,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
}

// OutputSettings are the protocol settings of the output eval.
type OutputSettings struct {
	Runs           int      `json:"runs,omitempty"`
	MinimumWinRate *float64 `json:"minimumWinRate,omitempty"`
}

// InstructionSettings are the protocol settings of the instruction eval.
type InstructionSettings struct {
	Runs        int      `json:"runs,omitempty"`
	MinimumLift *float64 `json:"minimumLift,omitempty"`
}

// OrchestrationSettings are the protocol settings of the orchestration eval.
type OrchestrationSettings struct {
	Runs int `json:"runs,omitempty"`
}

// ProtocolSettings groups the settings of every protocol.
type ProtocolSettings struct {
	Trigger       *TriggerSettings       `json:"trigger,omitempty"`
	Routing       *RoutingSettings       `json:"routing,omitempty"`
	Output        *OutputSettings        `json:"output,omitempty"`
	Instruction   *InstructionSettings   `json:"instruction,omitempty"`
	Orchestration *OrchestrationSettings `json:"orchestration,omitempty"`
}

// Budgets bound the model work of one run. A zero member means the default.
type Budgets struct {
	MaxModelRuns int     `json:"maxModelRuns,omitempty"`
	MaxCostUSD   float64 `json:"maxCostUsd,omitempty"`
}

// Sandbox is the confinement every harness run starts in. The zero value runs
// the harness on the host, as it runs without the tool.
type Sandbox struct {
	Isolation       string   `json:"isolation,omitempty"`
	Provision       string   `json:"provision,omitempty"`
	Network         string   `json:"network,omitempty"`
	Image           string   `json:"image,omitempty"`
	Runtime         string   `json:"runtime,omitempty"`
	HostPassthrough bool     `json:"hostPassthrough,omitempty"`
	InitCommands    []string `json:"initCommands,omitempty"`
	NixPackages     []string `json:"nixPackages,omitempty"`
	NixRev          string   `json:"nixRev,omitempty"`
	AllowHosts      []string `json:"allowHosts,omitempty"`
	Credentials     []string `json:"credentials,omitempty"`
	CustomEnv       []string `json:"customEnv,omitempty"`

	RequirePinnedProvision      bool `json:"requirePinnedProvision,omitempty"`
	RequireHostToolsUnreachable bool `json:"requireHostToolsUnreachable,omitempty"`
	RequireEgressRestricted     bool `json:"requireEgressRestricted,omitempty"`
	RequireKernelIsolation      bool `json:"requireKernelIsolation,omitempty"`
}

// The axis values that leave a run unconfined.
const (
	unconfinedIsolation = "none"
	unconfinedProvision = "none"
	unconfinedNetwork   = "host"
)

// Unconfined reports whether the sandbox asks for nothing: every axis at the
// value that confines nothing, and no input and no guarantee.
func (s Sandbox) Unconfined() bool {
	axes := (s.Isolation == "" || s.Isolation == unconfinedIsolation) &&
		(s.Provision == "" || s.Provision == unconfinedProvision) &&
		(s.Network == "" || s.Network == unconfinedNetwork)
	inputs := s.Image == "" && s.Runtime == "" && !s.HostPassthrough && s.NixRev == "" &&
		len(s.InitCommands)+len(s.NixPackages)+len(s.AllowHosts)+len(s.Credentials)+len(s.CustomEnv) == 0
	guarantees := !s.RequirePinnedProvision && !s.RequireHostToolsUnreachable && !s.RequireEgressRestricted && !s.RequireKernelIsolation
	return axes && inputs && guarantees
}

// Project is the merged eval section of agentic.json: where the plugins and
// instructions live, the profiles that run them, the protocol settings, the
// budgets and the sandbox of every harness run.
type Project struct {
	PluginDirs       []string           `json:"pluginDirs,omitempty"`
	InstructionRoots []string           `json:"instructionRoots,omitempty"`
	DefaultProfile   string             `json:"defaultProfile,omitempty"`
	Profiles         map[string]Profile `json:"profiles"`
	Protocols        ProtocolSettings   `json:"protocols,omitempty"`
	Budgets          Budgets            `json:"budgets"`
	Sandbox          Sandbox            `json:"sandbox,omitzero"`
}
