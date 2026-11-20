// Package settings resolves what one invocation runs with. A value comes from
// the first of: a command-line flag, an environment variable, the project file,
// and a built-in default. The package checks the hard caps that keep an eval
// run small and cheap, so no protocol has to.
package settings

import (
	"fmt"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/budget"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// Hard caps. A flag or a project file can lower them but never raise them.
const (
	MaxRuns        = 3
	MaxConcurrency = 2
	MaxTurns       = 24
	// MaxBatchModelRuns is the most model runs one batch of units may need.
	MaxBatchModelRuns = 24
)

// Per-run spend caps in US dollars, by what the run is for.
const (
	TriggerRunBudgetUSD  = 0.05
	GenerateRunBudgetUSD = 0.10
	JudgeRunBudgetUSD    = 0.10
	AgentRunBudgetUSD    = 0.25
)

// Per-run timeouts in seconds.
const (
	TriggerTimeoutSeconds  = 60
	GenerateTimeoutSeconds = 180
	AgentTimeoutSeconds    = 300
)

// Defaults of the protocols.
const (
	DefaultHarness          = "claude"
	DefaultRuns             = 1
	DefaultTriggerRuns      = 3
	DefaultThreshold        = 0.5
	DefaultMinimumLift      = 0.5
	DefaultProfileName      = "default"
	DefaultTriggerBatchSize = MaxBatchModelRuns / DefaultTriggerRuns
)

// models holds the default generation and judge model of each harness.
var models = map[string][2]string{
	"claude": {"claude-haiku-5-5", "claude-sonnet-5-5"},
	"codex":  {"gpt-5.3-codex", "gpt-5.3-codex"},
}

// environment names the variables that select a harness and its models.
var environment = map[string][2]string{
	"claude": {"CLAUDE_MODEL", "JUDGE_MODEL"},
	"codex":  {"CODEX_MODEL", "CODEX_JUDGE_MODEL"},
}

// HarnessVariable selects the harness from the environment.
const HarnessVariable = "SKILL_EVAL_HARNESS"

// Flags are the values a command line sets. A zero value means unset.
type Flags struct {
	Profile      string
	Harness      string
	Model        string
	JudgeModel   string
	Runs         int
	Threshold    *float64
	BatchSize    int
	Concurrency  int
	MaxCostUSD   float64
	MaxModelRuns int
	Split        string
}

// Settings are the resolved values of one invocation of one protocol.
type Settings struct {
	ProfileName string
	Harness     string
	Model       string
	JudgeModel  string
	Effort      model.Effort
	// TimeoutSeconds is the profile timeout, or zero for the protocol default.
	TimeoutSeconds int
	Runs           int
	Threshold      float64
	BatchSize      int
	Concurrency    int
	MinimumLift    float64
	MinimumWinRate float64
	Split          model.Split
	Limits         budget.Limits
}

// Resolve merges flags, environment and project for a protocol.
func Resolve(project model.Project, flags Flags, env map[string]string, protocol model.Protocol) (Settings, error) {
	name := first(flags.Profile, project.DefaultProfile, DefaultProfileName)
	profile, known := project.Profiles[name]
	if !known && (flags.Profile != "" || project.DefaultProfile != "") {
		return Settings{}, fmt.Errorf("profile %q is not defined in the project", name)
	}
	harness := first(flags.Harness, env[HarnessVariable], profile.Harness, DefaultHarness)
	defaults, ok := models[harness]
	if !ok {
		return Settings{}, fmt.Errorf("no eval support for harness %q", harness)
	}
	vars := environment[harness]
	s := Settings{
		ProfileName:    name,
		Harness:        harness,
		Model:          first(flags.Model, env[vars[0]], profile.Model, defaults[0]),
		JudgeModel:     first(flags.JudgeModel, env[vars[1]], defaults[1]),
		Effort:         profile.Effort,
		TimeoutSeconds: profile.TimeoutSeconds,
		Concurrency:    MaxConcurrency,
		Threshold:      DefaultThreshold,
		MinimumLift:    DefaultMinimumLift,
		Split:          model.Split(flags.Split),
	}
	s.protocol(project.Protocols, protocol, flags)
	if flags.Concurrency > 0 {
		s.Concurrency = flags.Concurrency
	}
	if flags.BatchSize > 0 {
		s.BatchSize = flags.BatchSize
	}
	s.Limits = limits(project.Budgets, flags)
	if s.Runs < 1 || s.Runs > MaxRuns {
		return Settings{}, fmt.Errorf("runs must be between 1 and %d, got %d", MaxRuns, s.Runs)
	}
	if s.Concurrency < 1 || s.Concurrency > MaxConcurrency {
		return Settings{}, fmt.Errorf("concurrency must be between 1 and %d, got %d", MaxConcurrency, s.Concurrency)
	}
	if s.Threshold < 0 || s.Threshold > 1 {
		return Settings{}, fmt.Errorf("threshold must be between 0 and 1, got %v", s.Threshold)
	}
	switch s.Split {
	case "", model.Train, model.Validation, model.Test:
	default:
		return Settings{}, fmt.Errorf("split must be train, validation or test, got %q", s.Split)
	}
	return s, nil
}

// protocol applies the settings of one protocol, then the runs and threshold
// flags that apply to every protocol.
func (s *Settings) protocol(p model.ProtocolSettings, protocol model.Protocol, flags Flags) {
	s.Runs = DefaultRuns
	switch protocol {
	case model.Trigger:
		s.Runs, s.BatchSize = DefaultTriggerRuns, DefaultTriggerBatchSize
		if t := p.Trigger; t != nil {
			s.Runs, s.BatchSize, s.Threshold = pick(t.Runs, s.Runs), pick(t.BatchSize, s.BatchSize), ptr(t.Threshold, s.Threshold)
		}
	case model.Routing:
		s.Runs = DefaultTriggerRuns
		if r := p.Routing; r != nil {
			s.Runs, s.Threshold = pick(r.Runs, s.Runs), ptr(r.Threshold, s.Threshold)
		}
	case model.Output:
		s.MinimumWinRate = 0
		if o := p.Output; o != nil {
			s.Runs, s.MinimumWinRate = pick(o.Runs, s.Runs), ptr(o.MinimumWinRate, 0)
		}
	case model.Instruction:
		if i := p.Instruction; i != nil {
			s.Runs, s.MinimumLift = pick(i.Runs, s.Runs), ptr(i.MinimumLift, s.MinimumLift)
		}
	case model.Orchestration:
		if o := p.Orchestration; o != nil {
			s.Runs = pick(o.Runs, s.Runs)
		}
	}
	if flags.Runs > 0 {
		s.Runs = flags.Runs
	}
	if flags.Threshold != nil {
		s.Threshold = *flags.Threshold
	}
}

func limits(project *model.Budgets, flags Flags) budget.Limits {
	var l budget.Limits
	if project != nil {
		l = budget.Limits{MaxModelRuns: project.MaxModelRuns, MaxCostUSD: project.MaxCostUSD}
	}
	if flags.MaxModelRuns > 0 {
		l.MaxModelRuns = flags.MaxModelRuns
	}
	if flags.MaxCostUSD > 0 {
		l.MaxCostUSD = flags.MaxCostUSD
	}
	return l
}

func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func pick(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func ptr(value *float64, fallback float64) float64 {
	if value != nil {
		return *value
	}
	return fallback
}
