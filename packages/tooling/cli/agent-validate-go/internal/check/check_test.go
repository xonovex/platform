package check_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

func TestDecodeOptions(t *testing.T) {
	type options struct {
		Limit int    `json:"limit"`
		Name  string `json:"name"`
	}
	got := options{Limit: 5, Name: "default"}
	if err := check.DecodeOptions(nil, &got); err != nil || got.Limit != 5 {
		t.Fatalf("no options keep the defaults: %v %+v", err, got)
	}
	if err := check.DecodeOptions(json.RawMessage(`{"limit": 9}`), &got); err != nil || got.Limit != 9 || got.Name != "default" {
		t.Fatalf("an option overrides one default: %v %+v", err, got)
	}
	err := check.DecodeOptions(json.RawMessage(`{"limt": 9}`), &got)
	if err == nil || !strings.Contains(err.Error(), "limt") {
		t.Fatalf("a misspelled option is an error: %v", err)
	}
	if err := check.DecodeOptions(json.RawMessage(`{"limit": "x"}`), &got); err == nil {
		t.Fatal("a wrong type is an error")
	}
}

func TestReporter(t *testing.T) {
	r := check.NewReporter("demo")
	r.Fail("demo.a", "x.md", 3, "failed %d", 1)
	r.Warn("demo.b", "y.md", 0, "warned")
	r.Drift("demo.c", "z.md", 1, "drifted")
	r.Merge([]findings.Finding{{Check: "other", Level: findings.Fail, Code: "other.d"}})
	got := r.Findings()
	if len(got) != 4 {
		t.Fatalf("findings = %v", got)
	}
	if got[0].Level != findings.Fail || got[0].Check != "demo" || got[0].Line != 3 || got[0].Message != "failed 1" {
		t.Fatalf("first = %+v", got[0])
	}
	if got[1].Level != findings.Warn || got[2].Level != findings.Drift || got[3].Check != "other" {
		t.Fatalf("levels = %+v", got)
	}
}
