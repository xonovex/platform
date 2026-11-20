// Package render writes the outcome of an audit for a person or a tool. It
// holds no audit logic: it formats findings that other packages produced.
package render

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sarif"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/audit"
)

// Format names an output format.
type Format string

// Output formats.
const (
	Text  Format = "text"
	JSON  Format = "json"
	SARIF Format = "sarif"
)

// Formats lists the output formats.
func Formats() []Format { return []Format{Text, JSON, SARIF} }

// Parse returns the format a name stands for.
func Parse(name string) (Format, error) {
	for _, format := range Formats() {
		if string(format) == name {
			return format, nil
		}
	}
	return "", fmt.Errorf("unknown output format %q: use text, json or sarif", name)
}

// Outcome is everything an audit run produced, ready to be written.
type Outcome struct {
	Catalogs []audit.CatalogReport
	// Compare holds the findings of a comparison against a base ref.
	Compare []findings.Finding
	// Err is the fault that stopped the run, if one did.
	Err error
}

// Findings returns every finding of the outcome in a stable order: the catalogs
// in the order given, then the comparison.
func (o Outcome) Findings() []findings.Finding {
	var all []findings.Finding
	for _, catalog := range o.Catalogs {
		all = append(all, catalog.Findings...)
	}
	return append(all, o.Compare...)
}

// Write writes the outcome in the format. The tool name and version go into the
// SARIF run.
func Write(w io.Writer, format Format, outcome Outcome, tool, version string) error {
	switch format {
	case JSON:
		return writeJSON(w, outcome)
	case SARIF:
		return writeSARIF(w, outcome, tool, version)
	default:
		return writeText(w, outcome)
	}
}

func writeText(w io.Writer, outcome Outcome) error {
	all := outcome.Findings()
	for _, f := range all {
		if f.Level == findings.Pass {
			continue
		}
		if _, err := fmt.Fprintln(w, f); err != nil {
			return err
		}
	}
	if outcome.Err != nil {
		if _, err := fmt.Fprintf(w, "error: %v\n", outcome.Err); err != nil {
			return err
		}
	}
	return writeSummary(w, all)
}

func writeSummary(w io.Writer, all []findings.Finding) error {
	counts := map[findings.Level]int{}
	for _, f := range all {
		counts[f.Level]++
	}
	_, err := fmt.Fprintf(w, "%d pass, %d warn, %d drift, %d fail\n",
		counts[findings.Pass], counts[findings.Warn], counts[findings.Drift], counts[findings.Fail])
	return err
}

type jsonDocument struct {
	Catalogs []audit.CatalogReport `json:"catalogs"`
	Compare  []findings.Finding    `json:"compare,omitempty"`
	Summary  []findings.Summary    `json:"summary"`
	Error    string                `json:"error,omitempty"`
}

func writeJSON(w io.Writer, outcome Outcome) error {
	document := jsonDocument{
		Catalogs: outcome.Catalogs,
		Compare:  outcome.Compare,
		Summary:  findings.Summarize(outcome.Findings()),
	}
	if document.Catalogs == nil {
		document.Catalogs = []audit.CatalogReport{}
	}
	if outcome.Err != nil {
		document.Error = outcome.Err.Error()
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(document)
}

func writeSARIF(w io.Writer, outcome Outcome, tool, version string) error {
	run := sarif.Run{ToolName: tool, ToolVersion: version, Findings: outcome.Findings()}
	if outcome.Err != nil {
		run.ExecutionFailed = true
		run.Notifications = []sarif.Notification{{Level: "error", Code: "audit.fault", Message: outcome.Err.Error()}}
	}
	return sarif.Write(w, run)
}
