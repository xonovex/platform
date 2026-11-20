// Package report renders the findings of a run as text or as SARIF.
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	contractreport "github.com/xonovex/platform/packages/library/shared-contract-go/pkg/report"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sarif"
)

// Formats lists the output formats that Write renders.
func Formats() []contractreport.Format {
	return []contractreport.Format{contractreport.Text, contractreport.SARIF}
}

// Meta names the tool that produced the run.
type Meta struct {
	Tool    string
	Version string
	URI     string
}

// Write renders the findings. A text report lists each finding that is not a
// pass, then one summary line per check that ran.
func Write(w io.Writer, format contractreport.Format, meta Meta, ran []string, all []findings.Finding) error {
	if format == contractreport.SARIF {
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
func WriteFailure(w io.Writer, format contractreport.Format, meta Meta, cause error) error {
	if format != contractreport.SARIF {
		return nil
	}
	return sarif.Write(w, sarif.Run{
		ToolName: meta.Tool, ToolVersion: meta.Version, InformationURI: meta.URI,
		Notifications:   []sarif.Notification{{Level: "error", Code: "validate.failed", Message: cause.Error()}},
		ExecutionFailed: true,
	})
}
