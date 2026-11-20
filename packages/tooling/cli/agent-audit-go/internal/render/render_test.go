package render_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/audit"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/render"
)

func outcome() render.Outcome {
	return render.Outcome{
		Catalogs: []audit.CatalogReport{{Dir: "d", Findings: []findings.Finding{
			{Check: "audit", Level: findings.Pass, Code: "audit.ok", Message: "fine"},
			{Check: "audit", Level: findings.Drift, Code: "audit.metadata-changed", Path: "d/sources.json", Message: "moved"},
		}}},
		Compare: []findings.Finding{{Check: "audit", Level: findings.Fail, Code: "audit.review-regressed", Message: "back"}},
	}
}

func TestParse(t *testing.T) {
	for _, format := range render.Formats() {
		got, err := render.Parse(string(format))
		if err != nil || got != format {
			t.Errorf("Parse(%s) = %s, %v", format, got, err)
		}
	}
	if _, err := render.Parse("xml"); err == nil {
		t.Error("an unknown format must be an error")
	}
}

func TestFindingsOrder(t *testing.T) {
	all := outcome().Findings()
	if len(all) != 3 || all[2].Code != "audit.review-regressed" {
		t.Fatalf("Findings = %v", all)
	}
}

func TestTextHidesPassesAndSummarises(t *testing.T) {
	var out bytes.Buffer
	if err := render.Write(&out, render.Text, outcome(), "agent-audit", "1"); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{"d/sources.json: drift audit.metadata-changed: moved", "fail audit.review-regressed", "1 pass, 0 warn, 1 drift, 1 fail"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "audit.ok") {
		t.Errorf("a pass must not be listed:\n%s", text)
	}
}

func TestTextShowsTheFault(t *testing.T) {
	var out bytes.Buffer
	o := outcome()
	o.Err = errors.New("disk gone")
	if err := render.Write(&out, render.Text, o, "t", "1"); err != nil || !strings.Contains(out.String(), "error: disk gone") {
		t.Fatalf("out = %q, err = %v", out.String(), err)
	}
}

func TestJSON(t *testing.T) {
	var out bytes.Buffer
	o := outcome()
	o.Err = errors.New("late")
	if err := render.Write(&out, render.JSON, o, "t", "1"); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Catalogs []struct{ Dir string } `json:"catalogs"`
		Compare  []findings.Finding     `json:"compare"`
		Summary  []findings.Summary     `json:"summary"`
		Error    string                 `json:"error"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Catalogs) != 1 || doc.Catalogs[0].Dir != "d" || len(doc.Compare) != 1 || len(doc.Summary) != 1 || doc.Error != "late" {
		t.Errorf("doc = %+v", doc)
	}
}

func TestJSONOfNothingIsAnEmptyList(t *testing.T) {
	var out bytes.Buffer
	if err := render.Write(&out, render.JSON, render.Outcome{}, "t", "1"); err != nil || !strings.Contains(out.String(), `"catalogs": []`) {
		t.Fatalf("out = %q, err = %v", out.String(), err)
	}
}

func TestSARIF(t *testing.T) {
	var out bytes.Buffer
	o := outcome()
	o.Err = errors.New("late")
	if err := render.Write(&out, render.SARIF, o, "agent-audit", "9.9"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"version": "2.1.0"`, `"name": "agent-audit"`, `audit.metadata-changed`, `audit.fault`} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in:\n%s", want, out.String())
		}
	}
}

type failingWriter struct{ after int }

func (f *failingWriter) Write(p []byte) (int, error) {
	if f.after <= 0 {
		return 0, errors.New("closed")
	}
	f.after--
	return len(p), nil
}

func TestWriteReportsAFailingWriter(t *testing.T) {
	for _, format := range render.Formats() {
		if err := render.Write(&failingWriter{}, format, outcome(), "t", "1"); err == nil {
			t.Errorf("%s: a failing writer must be an error", format)
		}
	}
	// The summary line and the fault line fail on their own.
	o := outcome()
	o.Err = errors.New("x")
	for after := 0; after < 4; after++ {
		if err := render.Write(&failingWriter{after: after}, render.Text, o, "t", "1"); err == nil {
			t.Errorf("text after %d writes: want an error", after)
		}
	}
}
