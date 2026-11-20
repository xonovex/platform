// Package report renders the findings of a run as text or as SARIF.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sarif"
)

// Format is an output format.
type Format string

// The formats.
const (
	Text  Format = "text"
	SARIF Format = "sarif"
)

// ParseFormat reads a format name.
func ParseFormat(name string) (Format, error) {
	switch Format(name) {
	case Text, SARIF:
		return Format(name), nil
	}
	return "", fmt.Errorf("unknown format %q; use text or sarif", name)
}

// Meta names the tool that produced the run.
type Meta struct {
	Tool    string
	Version string
	URI     string
}

// Write renders the findings. A text report lists each finding that is not a
// pass, then one summary line per check that ran.
func Write(w io.Writer, format Format, meta Meta, ran []string, all []findings.Finding) error {
	if format == SARIF {
		return sarif.Write(w, sarif.Run{ToolName: meta.Tool, ToolVersion: meta.Version, InformationURI: meta.URI, Findings: all})
	}
	var b strings.Builder
	for _, f := range all {
		if f.Level != findings.Pass {
			fmt.Fprintf(&b, "[%s] %s\n", f.Check, f)
		}
	}
	counts := map[string]findings.Summary{}
	for _, s := range findings.Summarize(all) {
		counts[s.Check] = s
	}
	for _, name := range ran {
		s := counts[name]
		fmt.Fprintf(&b, "%s: %d fail, %d drift, %d warn\n", name, s.Fail, s.Drift, s.Warn)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// WriteFailure renders a run that did not finish. A SARIF log marks the run as
// failed, so a consumer reads no verdict from it; a text report leaves the
// message to the error stream.
func WriteFailure(w io.Writer, format Format, meta Meta, cause error) error {
	if format != SARIF {
		return nil
	}
	return sarif.Write(w, sarif.Run{
		ToolName: meta.Tool, ToolVersion: meta.Version, InformationURI: meta.URI,
		Notifications:   []sarif.Notification{{Level: "error", Code: "validate.failed", Message: cause.Error()}},
		ExecutionFailed: true,
	})
}
