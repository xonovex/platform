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

// Project is the merged eval section of agentic.json: where the plugins and
// instructions live, the profiles that run them, the protocol settings and the
// budgets.
type Project struct {
	PluginDirs       []string           `json:"pluginDirs,omitempty"`
	InstructionRoots []string           `json:"instructionRoots,omitempty"`
	DefaultProfile   string             `json:"defaultProfile,omitempty"`
	Profiles         map[string]Profile `json:"profiles"`
	Protocols        ProtocolSettings   `json:"protocols,omitempty"`
	Budgets          Budgets            `json:"budgets"`
}
