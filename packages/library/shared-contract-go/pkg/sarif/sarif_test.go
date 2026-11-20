package sarif_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sarif"
)

type decoded struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []struct {
		Tool struct {
			Driver struct {
				Name  string `json:"name"`
				Rules []struct {
					ID string `json:"id"`
				} `json:"rules"`
			} `json:"driver"`
		} `json:"tool"`
		Invocations []struct {
			ExecutionSuccessful        bool `json:"executionSuccessful"`
			ToolExecutionNotifications []struct {
				Level      string                `json:"level"`
				Message    struct{ Text string } `json:"message"`
				Descriptor *struct{ ID string }  `json:"descriptor"`
			} `json:"toolExecutionNotifications"`
		} `json:"invocations"`
		Results []struct {
			RuleID    string `json:"ruleId"`
			Level     string `json:"level"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct{ URI string }     `json:"artifactLocation"`
					Region           *struct{ StartLine int } `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
		} `json:"results"`
	} `json:"runs"`
}

func write(t *testing.T, run sarif.Run) decoded {
	t.Helper()
	var buf bytes.Buffer
	if err := sarif.Write(&buf, run); err != nil {
		t.Fatal(err)
	}
	var out decoded
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	return out
}

func TestWriteMapsFindings(t *testing.T) {
	out := write(t, sarif.Run{
		ToolName: "agent-validate",
		Findings: []findings.Finding{
			{Level: findings.Fail, Code: "skill.name", Path: "a/SKILL.md", Line: 4, Message: "bad name"},
			{Level: findings.Pass, Code: "skill.ok", Message: "fine"},
			{Level: findings.Drift, Code: "source.stale", Path: "SOURCES.json", Message: "stale"},
			{Level: findings.Warn, Code: "skill.name", Message: "again"},
		},
	})
	if out.Version != sarif.Version || out.Schema != sarif.SchemaURI {
		t.Fatalf("header = %q %q", out.Version, out.Schema)
	}
	run := out.Runs[0]
	if run.Tool.Driver.Name != "agent-validate" {
		t.Fatalf("tool = %q", run.Tool.Driver.Name)
	}
	if len(run.Results) != 3 {
		t.Fatalf("results = %d; a pass is not reported", len(run.Results))
	}
	wantLevels := []string{"error", "warning", "note"}
	for i, want := range wantLevels {
		if run.Results[i].Level != want {
			t.Errorf("result %d level = %q; want %q", i, run.Results[i].Level, want)
		}
	}
	first := run.Results[0].Locations[0].PhysicalLocation
	if first.ArtifactLocation.URI != "a/SKILL.md" || first.Region == nil || first.Region.StartLine != 4 {
		t.Fatalf("first location = %+v", first)
	}
	second := run.Results[1].Locations[0].PhysicalLocation
	if second.Region != nil {
		t.Fatal("a finding without a line has no region")
	}
	if len(run.Results[2].Locations) != 0 {
		t.Fatal("a finding without a path has no location")
	}
	if len(run.Tool.Driver.Rules) != 2 || run.Tool.Driver.Rules[0].ID != "skill.name" || run.Tool.Driver.Rules[1].ID != "source.stale" {
		t.Fatalf("rules = %+v", run.Tool.Driver.Rules)
	}
}

func TestWriteNotificationsAndExecutionState(t *testing.T) {
	out := write(t, sarif.Run{
		ToolName:        "agent-audit",
		ExecutionFailed: true,
		Notifications: []sarif.Notification{
			{Level: "error", Code: "transport.git", Message: "clone failed"},
			{Level: "note", Message: "skipped"},
		},
	})
	inv := out.Runs[0].Invocations[0]
	if inv.ExecutionSuccessful {
		t.Fatal("a failed run is not successful")
	}
	if len(inv.ToolExecutionNotifications) != 2 {
		t.Fatalf("notifications = %+v", inv.ToolExecutionNotifications)
	}
	if d := inv.ToolExecutionNotifications[0].Descriptor; d == nil || d.ID != "transport.git" {
		t.Fatalf("descriptor = %+v", d)
	}
	if inv.ToolExecutionNotifications[1].Descriptor != nil {
		t.Fatal("a notification without a code has no descriptor")
	}
}

func TestWriteEmptyRunKeepsArrays(t *testing.T) {
	var buf bytes.Buffer
	if err := sarif.Write(&buf, sarif.Run{ToolName: "t"}); err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	run := raw["runs"].([]any)[0].(map[string]any)
	if results, ok := run["results"].([]any); !ok || len(results) != 0 {
		t.Fatalf("results = %v; want an empty array", run["results"])
	}
	if !out(run) {
		t.Fatal("a clean run is successful")
	}
}

func out(run map[string]any) bool {
	return run["invocations"].([]any)[0].(map[string]any)["executionSuccessful"].(bool)
}

func TestWriteIsDeterministic(t *testing.T) {
	in := sarif.Run{ToolName: "t", Findings: []findings.Finding{
		{Level: findings.Fail, Code: "b", Message: "m"},
		{Level: findings.Fail, Code: "a", Message: "m"},
	}}
	var first, second bytes.Buffer
	_ = sarif.Write(&first, in)
	_ = sarif.Write(&second, in)
	if first.String() != second.String() {
		t.Fatal("equal input must give equal output")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }

func TestWritePropagatesWriterError(t *testing.T) {
	if err := sarif.Write(failingWriter{}, sarif.Run{ToolName: "t"}); err == nil {
		t.Fatal("a failing writer must fail Write")
	}
}

func TestLevel(t *testing.T) {
	if _, ok := sarif.Level(findings.Pass); ok {
		t.Fatal("pass has no SARIF level")
	}
	if got, ok := sarif.Level(findings.Fail); !ok || got != "error" {
		t.Fatalf("Level(fail) = %q, %v", got, ok)
	}
}
