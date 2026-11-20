package explain_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/explain"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/layer"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/merge"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

var section = spec.Section{Name: "eval", Keys: []spec.Key{
	{Section: "eval", Pointer: "/pluginDirs", Rule: spec.Replace, Layers: spec.EveryFile},
	{Section: "eval", Pointer: "/strict", Rule: spec.Require, Layers: spec.EveryFile},
	{Section: "eval", Pointer: "/profiles", Rule: spec.Merge, Layers: spec.EveryFile},
	{Section: "eval", Pointer: "/budgets/maxModelRuns", Rule: spec.Cap, Layers: spec.EveryFile, Ceiling: 24},
}}

func resolved(t *testing.T) merge.Resolved {
	t.Helper()
	user, _ := layer.Parse(spec.User, "/u/agentic.json", []byte(`{"version":1,"eval":{"strict":false,"profiles":{"b":{"m": 1},"a":{}}}}`))
	project, _ := layer.Parse(spec.Project, "/p/agentic.json", []byte(`{"version":1,"eval":{"pluginDirs":[ "x" ],"profiles":{"a":{"harness":"claude"}}}}`))
	r, err := merge.Merge([]spec.Section{section}, []layer.Document{user, project}, []merge.Flag{{Section: "eval", Pointer: "/budgets/maxModelRuns", Value: json.RawMessage("5")}})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRenderNamesTheOriginOfEveryValue(t *testing.T) {
	var out strings.Builder
	if err := explain.Render(&out, resolved(t)); err != nil {
		t.Fatal(err)
	}
	want := `eval.pluginDirs = ["x"]  (project /p/agentic.json)
eval.strict = false  (unset)
eval.profiles.a = {"harness":"claude"}  (project /p/agentic.json)
eval.profiles.b = {"m":1}  (user /u/agentic.json)
eval.budgets.maxModelRuns = 5  (flags)
`
	if out.String() != want {
		t.Fatalf("got\n%s\nwant\n%s", out.String(), want)
	}
}

type failAfter struct{ lines int }

func (w *failAfter) Write(p []byte) (int, error) {
	if w.lines == 0 {
		return 0, errors.New("closed")
	}
	w.lines--
	return len(p), nil
}

func TestRenderReportsAWriteFault(t *testing.T) {
	for _, lines := range []int{0, 2} {
		if err := explain.Render(&failAfter{lines: lines}, resolved(t)); err == nil {
			t.Fatalf("a write fault after %d lines must surface", lines)
		}
	}
}
