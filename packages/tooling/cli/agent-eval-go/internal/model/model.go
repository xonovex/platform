// Package model holds the typed form of every eval document: the trigger,
// output, instruction and orchestration evals, the project file, results,
// baselines and the cost history. It is pure data with no behavior beyond
// decoding and encoding, so every other package can depend on it. The JSON
// schemas in shared-contract-go define what a valid document is; this package
// only gives a valid document a shape.
package model

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/condition"
)

// Version is the contract version every document carries.
const Version = 1

// File names that a skill directory uses for its eval documents.
const (
	TriggerFile = "eval-trigger.json"
	OutputFile  = "eval-output.json"
)

// ProjectFile is the default name of the eval project file.
const ProjectFile = "agentic.eval.json"

// Protocol names the kind of eval a document or a result belongs to.
type Protocol string

// The protocols.
const (
	Trigger       Protocol = "trigger"
	Routing       Protocol = "routing"
	Output        Protocol = "output"
	Instruction   Protocol = "instruction"
	Orchestration Protocol = "orchestration"
)

// Protocols lists every protocol.
func Protocols() []Protocol {
	return []Protocol{Trigger, Routing, Output, Instruction, Orchestration}
}

// Condition is the state of the thing under test during a run. For an
// instruction it is the state of the rule; for a skill output it is present
// with the skill and absent without it.
type Condition = condition.Condition

// The conditions.
const (
	Present = condition.Present
	Absent  = condition.Absent
	Ablated = condition.Ablated
)

// Prompt holds either an inline prompt or the path of a Markdown file that
// holds a long one. Text is the resolved prompt, filled by the loader and never
// written back.
type Prompt struct {
	Prompt     string `json:"prompt,omitempty"`
	PromptFile string `json:"promptFile,omitempty"`
	Text       string `json:"-"`
}

// Usage is the model usage of one run or of a sum of runs.
type Usage struct {
	InputTokens  int     `json:"inputTokens"`
	OutputTokens int     `json:"outputTokens"`
	CostUSD      float64 `json:"costUsd"`
}

// Add returns the sum of two usages.
func (u Usage) Add(other Usage) Usage {
	return Usage{
		InputTokens:  u.InputTokens + other.InputTokens,
		OutputTokens: u.OutputTokens + other.OutputTokens,
		CostUSD:      u.CostUSD + other.CostUSD,
	}
}

// Predicate is one check on the outcome of a run. Kind selects which of the
// other members apply: filesystem uses Path, Check and Value; session uses
// Event, Match, Min and Max; llm uses Criterion and Threshold.
type Predicate struct {
	Kind      string   `json:"kind"`
	Path      string   `json:"path,omitempty"`
	Check     string   `json:"check,omitempty"`
	Value     string   `json:"value,omitempty"`
	Event     string   `json:"event,omitempty"`
	Match     string   `json:"match,omitempty"`
	Min       *int     `json:"min,omitempty"`
	Max       *int     `json:"max,omitempty"`
	Criterion string   `json:"criterion,omitempty"`
	Threshold *float64 `json:"threshold,omitempty"`
}

// The predicate kinds.
const (
	KindFilesystem = "filesystem"
	KindSession    = "session"
	KindLLM        = "llm"
)

// Decode parses a JSON document into v. Unknown members are ignored: the
// schema, not the decoder, decides which members are allowed.
func Decode(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("decode document: %w", err)
	}
	return nil
}

// Encode renders v as indented JSON that ends in a newline, so a document that
// is written and read again stays byte-identical.
func Encode(v any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(v); err != nil {
		return nil, fmt.Errorf("encode document: %w", err)
	}
	return buffer.Bytes(), nil
}
