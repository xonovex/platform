// Package sarif writes findings as a SARIF 2.1.0 log, including the execution
// notifications that report an infrastructure fault apart from a finding.
package sarif

import (
	"encoding/json"
	"io"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
)

const (
	// Version is the SARIF version the writer emits.
	Version = "2.1.0"
	// SchemaURI names the SARIF 2.1.0 schema.
	SchemaURI = "https://json.schemastore.org/sarif-2.1.0.json"
)

// Notification is a message about the run itself, such as a transport that
// could not be reached. It is not a finding about the checked content.
type Notification struct {
	Level   string
	Code    string
	Message string
}

// Run is one tool execution.
type Run struct {
	ToolName       string
	ToolVersion    string
	InformationURI string
	Findings       []findings.Finding
	Notifications  []Notification
	// ExecutionFailed marks a run that did not finish, so it has no verdict.
	ExecutionFailed bool
}

// Level returns the SARIF level a finding level maps to, and false for pass,
// which a SARIF log does not report as a result.
func Level(l findings.Level) (string, bool) {
	switch l {
	case findings.Fail:
		return "error", true
	case findings.Drift:
		return "warning", true
	case findings.Warn:
		return "note", true
	default:
		return "", false
	}
}

// Write encodes the run as an indented SARIF log. Results keep the order of the
// findings, and rules are sorted by id so equal input gives equal output.
func Write(w io.Writer, run Run) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(build(run))
}

type log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

type run struct {
	Tool        tool         `json:"tool"`
	Invocations []invocation `json:"invocations"`
	Results     []result     `json:"results"`
}

type tool struct {
	Driver driver `json:"driver"`
}

type driver struct {
	Name           string `json:"name"`
	Version        string `json:"version,omitempty"`
	InformationURI string `json:"informationUri,omitempty"`
	Rules          []rule `json:"rules"`
}

type rule struct {
	ID               string  `json:"id"`
	ShortDescription message `json:"shortDescription"`
}

type message struct {
	Text string `json:"text"`
}

type invocation struct {
	ExecutionSuccessful        bool           `json:"executionSuccessful"`
	ToolExecutionNotifications []notification `json:"toolExecutionNotifications"`
}

type notification struct {
	Level      string      `json:"level"`
	Message    message     `json:"message"`
	Descriptor *descriptor `json:"descriptor,omitempty"`
}

type descriptor struct {
	ID string `json:"id"`
}

type result struct {
	RuleID    string     `json:"ruleId"`
	Level     string     `json:"level"`
	Message   message    `json:"message"`
	Locations []location `json:"locations,omitempty"`
}

type location struct {
	PhysicalLocation physicalLocation `json:"physicalLocation"`
}

type physicalLocation struct {
	ArtifactLocation artifactLocation `json:"artifactLocation"`
	Region           *region          `json:"region,omitempty"`
}

type artifactLocation struct {
	URI string `json:"uri"`
}

type region struct {
	StartLine int `json:"startLine"`
}

func build(in Run) log {
	results := make([]result, 0, len(in.Findings))
	seen := map[string]struct{}{}
	rules := []rule{}
	for _, f := range in.Findings {
		level, reported := Level(f.Level)
		if !reported {
			continue
		}
		results = append(results, toResult(f, level))
		if _, ok := seen[f.Code]; !ok {
			seen[f.Code] = struct{}{}
			rules = append(rules, rule{ID: f.Code, ShortDescription: message{Text: f.Code}})
		}
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].ID < rules[j].ID })

	notifications := make([]notification, 0, len(in.Notifications))
	for _, n := range in.Notifications {
		item := notification{Level: n.Level, Message: message{Text: n.Message}}
		if n.Code != "" {
			item.Descriptor = &descriptor{ID: n.Code}
		}
		notifications = append(notifications, item)
	}

	return log{
		Schema:  SchemaURI,
		Version: Version,
		Runs: []run{{
			Tool: tool{Driver: driver{
				Name:           in.ToolName,
				Version:        in.ToolVersion,
				InformationURI: in.InformationURI,
				Rules:          rules,
			}},
			Invocations: []invocation{{
				ExecutionSuccessful:        !in.ExecutionFailed,
				ToolExecutionNotifications: notifications,
			}},
			Results: results,
		}},
	}
}

func toResult(f findings.Finding, level string) result {
	r := result{RuleID: f.Code, Level: level, Message: message{Text: f.Message}}
	if f.Path == "" {
		return r
	}
	physical := physicalLocation{ArtifactLocation: artifactLocation{URI: f.Path}}
	if f.Line > 0 {
		physical.Region = &region{StartLine: f.Line}
	}
	r.Locations = []location{{PhysicalLocation: physical}}
	return r
}
