package model

// Split is the part of a trigger set that a query belongs to.
type Split string

// The splits.
const (
	Train      Split = "train"
	Validation Split = "validation"
	Test       Split = "test"
)

// TriggerEval is the document of prompts that should or should not make an
// agent load one skill.
type TriggerEval struct {
	Schema  string         `json:"$schema,omitempty"`
	Version int            `json:"version"`
	Skill   string         `json:"skill"`
	Queries []TriggerQuery `json:"queries"`
}

// TriggerQuery is one prompt of a trigger eval.
type TriggerQuery struct {
	ID string `json:"id"`
	Prompt
	ShouldTrigger bool     `json:"shouldTrigger"`
	Split         Split    `json:"split"`
	Tags          []string `json:"tags,omitempty"`
	Rationale     string   `json:"rationale,omitempty"`
}

// Tier sets how strict the quality gate of an output eval is.
type Tier string

// The tiers.
const (
	Aggressive   Tier = "aggressive"
	Moderate     Tier = "moderate"
	Conservative Tier = "conservative"
)

// Gate says whether a failing case fails the run.
type Gate string

// The gates.
const (
	Required Gate = "required"
	Advisory Gate = "advisory"
)

// OutputEval is the document of cases that run a prompt with and without a
// skill and judge the two outputs.
type OutputEval struct {
	Schema  string       `json:"$schema,omitempty"`
	Version int          `json:"version"`
	Skill   string       `json:"skill"`
	Tier    Tier         `json:"tier,omitempty"`
	Judge   *JudgeSpec   `json:"judge,omitempty"`
	Cases   []OutputCase `json:"cases"`
}

// JudgeSpec names the harness and model that grade the outputs of a case.
type JudgeSpec struct {
	Harness string `json:"harness"`
	Model   string `json:"model,omitempty"`
	Rubric  string `json:"rubric,omitempty"`
}

// Fixture is a file placed in the workspace before a run. Text is the resolved
// content, filled by the loader and never written back.
type Fixture struct {
	Path        string `json:"path"`
	Content     string `json:"content,omitempty"`
	ContentFile string `json:"contentFile,omitempty"`
	Text        string `json:"-"`
}

// OutputCase is one prompt of an output eval.
type OutputCase struct {
	ID string `json:"id"`
	Prompt
	Gate     Gate        `json:"gate"`
	Fixtures []Fixture   `json:"fixtures,omitempty"`
	Expect   []Predicate `json:"expect"`
}

// Anchor names the rule of an instruction file that a case tests, by a heading
// or by an exact quotation.
type Anchor struct {
	Heading string `json:"heading,omitempty"`
	Text    string `json:"text,omitempty"`
}

// Expectations lists the predicates that must hold under each condition.
type Expectations struct {
	Present []Predicate `json:"present,omitempty"`
	Absent  []Predicate `json:"absent,omitempty"`
	Ablated []Predicate `json:"ablated,omitempty"`
}

// For returns the predicates of one condition.
func (e Expectations) For(c Condition) []Predicate {
	switch c {
	case Present:
		return e.Present
	case Absent:
		return e.Absent
	case Ablated:
		return e.Ablated
	}
	return nil
}

// InstructionEval is the document of cases that test one rule of an
// instruction file under each condition.
type InstructionEval struct {
	Schema      string            `json:"$schema,omitempty"`
	Version     int               `json:"version"`
	Instruction string            `json:"instruction"`
	Cases       []InstructionCase `json:"cases"`
}

// InstructionCase is one prompt of an instruction eval.
type InstructionCase struct {
	ID string `json:"id"`
	Prompt
	Anchor Anchor       `json:"anchor"`
	Expect Expectations `json:"expect"`
	Ablate string       `json:"ablate,omitempty"`
}

// Conditions lists the conditions the case declares expectations for, in the
// order present, absent, ablated.
func (c InstructionCase) Conditions() []Condition {
	var out []Condition
	for _, condition := range []Condition{Present, Absent, Ablated} {
		if len(c.Expect.For(condition)) > 0 {
			out = append(out, condition)
		}
	}
	return out
}

// OrchestrationEval is the document of a flow in which several agents each run
// a step after the steps they depend on.
type OrchestrationEval struct {
	Schema      string      `json:"$schema,omitempty"`
	Version     int         `json:"version"`
	ID          string      `json:"id"`
	Description string      `json:"description,omitempty"`
	Agents      []Agent     `json:"agents"`
	Steps       []Step      `json:"steps"`
	Budget      *FlowBudget `json:"budget,omitempty"`
}

// Agent is one participant of an orchestration.
type Agent struct {
	ID      string `json:"id"`
	Harness string `json:"harness"`
	Model   string `json:"model,omitempty"`
	Role    string `json:"role,omitempty"`
}

// Step is one prompt of an orchestration, run by one agent.
type Step struct {
	ID string `json:"id"`
	Prompt
	Agent  string      `json:"agent"`
	After  []string    `json:"after,omitempty"`
	Expect []Predicate `json:"expect,omitempty"`
}

// FlowBudget bounds one orchestration.
type FlowBudget struct {
	MaxModelRuns   int     `json:"maxModelRuns,omitempty"`
	MaxCostUSD     float64 `json:"maxCostUsd,omitempty"`
	TimeoutSeconds int     `json:"timeoutSeconds,omitempty"`
}
