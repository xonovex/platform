package report_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/report"
)

var sample = []findings.Finding{
	{Check: "skill", Level: findings.Fail, Code: "x", Path: "a.md", Line: 3, Message: "bad"},
	{Check: "skill", Level: findings.Pass, Code: "ok", Message: "fine"},
	{Check: "text", Level: findings.Warn, Code: "w", Path: "b.md", Message: "meh"},
}

func TestParseFormat(t *testing.T) {
	if f, err := report.ParseFormat("sarif"); err != nil || f != report.SARIF {
		t.Fatal("sarif")
	}
	if _, err := report.ParseFormat("xml"); err == nil {
		t.Fatal("unknown format")
	}
}

func TestWriteText(t *testing.T) {
	var out bytes.Buffer
	if err := report.Write(&out, report.Text, report.Meta{}, []string{"skill", "text", "idle"}, sample); err != nil {
		t.Fatal(err)
	}
	want := "[skill] a.md:3: fail x: bad\n[text] b.md: warn w: meh\nskill: 1 fail, 0 drift, 0 warn\ntext: 0 fail, 0 drift, 1 warn\nidle: 0 fail, 0 drift, 0 warn\n"
	if out.String() != want {
		t.Fatalf("got %q", out.String())
	}
}

func TestWriteSARIF(t *testing.T) {
	var out bytes.Buffer
	meta := report.Meta{Tool: "agent-validate", Version: "1", URI: "https://example.com"}
	if err := report.Write(&out, report.SARIF, meta, nil, sample); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"name": "agent-validate"`) || !strings.Contains(out.String(), `"ruleId"`) {
		t.Fatalf("got %s", out.String())
	}
}

func TestWriteFailure(t *testing.T) {
	var out bytes.Buffer
	if err := report.WriteFailure(&out, report.Text, report.Meta{}, errors.New("boom")); err != nil || out.Len() != 0 {
		t.Fatalf("text leaves the message to stderr: %q %v", out.String(), err)
	}
	if err := report.WriteFailure(&out, report.SARIF, report.Meta{Tool: "agent-validate"}, errors.New("boom")); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"boom", "validate.failed", `"executionSuccessful": false`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in %s", want, out.String())
		}
	}
}
