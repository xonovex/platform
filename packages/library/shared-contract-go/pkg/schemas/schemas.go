// Package schemas embeds the JSON Schema documents of the contract. Each
// schema has an $id under Namespace, and the constants name them, so a caller
// never spells a schema URL by hand.
package schemas

import (
	"embed"
	"encoding/json"
	"fmt"
	iofs "io/fs"
	"sort"
)

// Namespace is the URL prefix of every schema id.
const Namespace = "https://schemas.xonovex.com/agent/"

// Schema ids.
const (
	Common            = Namespace + "common.schema.json"
	Predicate         = Namespace + "predicate.schema.json"
	Sources           = Namespace + "sources.schema.json"
	EvalTrigger       = Namespace + "eval-trigger.schema.json"
	EvalOutput        = Namespace + "eval-output.schema.json"
	EvalInstruction   = Namespace + "eval-instruction.schema.json"
	EvalOrchestration = Namespace + "eval-orchestration.schema.json"
	Results           = Namespace + "results.schema.json"
	Baseline          = Namespace + "baseline.schema.json"
	CostHistory       = Namespace + "cost-history.schema.json"
	Kit               = Namespace + "kit.schema.json"
)

// Ids of the agentic.json schemas: the file and the section of each tool.
const (
	Agentic          = Namespace + "agentic.schema.json"
	AgenticRun       = Namespace + "agentic-run.schema.json"
	AgenticProviders = Namespace + "agentic-providers.schema.json"
	AgenticValidate  = Namespace + "agentic-validate.schema.json"
	AgenticAudit     = Namespace + "agentic-audit.schema.json"
	AgenticEval      = Namespace + "agentic-eval.schema.json"
)

// Ids of the agent schemas: a model provider route, the facts of a harness and
// an override of some of them.
const (
	Provider        = Namespace + "provider.schema.json"
	Harness         = Namespace + "harness.schema.json"
	HarnessOverride = Namespace + "harness-override.schema.json"
)

//go:embed *.schema.json
var files embed.FS

// IDs lists the id of every embedded schema, sorted.
func IDs() []string {
	ids := []string{
		Common, Predicate, Sources, EvalTrigger, EvalOutput, EvalInstruction,
		EvalOrchestration, Results, Baseline, CostHistory, Kit,
		Agentic, AgenticRun, AgenticProviders, AgenticValidate, AgenticAudit, AgenticEval,
		Provider, Harness, HarnessOverride,
	}
	sort.Strings(ids)
	return ids
}

// Documents returns every embedded schema keyed by its $id.
func Documents() (map[string][]byte, error) {
	return read(files)
}

func read(files iofs.FS) (map[string][]byte, error) {
	entries, err := iofs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("list embedded schemas: %w", err)
	}
	documents := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		data, err := iofs.ReadFile(files, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read schema %s: %w", entry.Name(), err)
		}
		var header struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal(data, &header); err != nil {
			return nil, fmt.Errorf("parse schema %s: %w", entry.Name(), err)
		}
		if header.ID == "" {
			return nil, fmt.Errorf("schema %s has no $id", entry.Name())
		}
		if _, dup := documents[header.ID]; dup {
			return nil, fmt.Errorf("schema %s repeats the $id %s", entry.Name(), header.ID)
		}
		documents[header.ID] = data
	}
	return documents, nil
}
