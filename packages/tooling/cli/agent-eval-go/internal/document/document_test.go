package document_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/memfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

func reader(t *testing.T, files map[string]string) document.Reader {
	t.Helper()
	v, err := wiring.NewValidator()
	if err != nil {
		t.Fatal(err)
	}
	return document.Reader{FS: memfs.New(files), Validator: v}
}

func invalidOf(t *testing.T, err error) *document.InvalidError {
	t.Helper()
	var bad *document.InvalidError
	if !errors.As(err, &bad) {
		t.Fatalf("want InvalidError, got %v", err)
	}
	return bad
}

const trigger = `{"version":1,"skill":"s","queries":[
 {"id":"a","prompt":"p1","shouldTrigger":true,"split":"train"},
 {"id":"b","promptFile":"long.md","shouldTrigger":false,"split":"validation"}]}`

func TestTriggerResolvesPromptFiles(t *testing.T) {
	r := reader(t, map[string]string{"s/eval-trigger.json": trigger, "s/long.md": "long prompt\n"})
	doc, err := r.Trigger("s/eval-trigger.json")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Queries[0].Text != "p1" || doc.Queries[1].Text != "long prompt" {
		t.Fatalf("%+v", doc.Queries)
	}
}

func TestTriggerMissingPromptFileIsAFault(t *testing.T) {
	r := reader(t, map[string]string{"s/eval-trigger.json": trigger})
	_, err := r.Trigger("s/eval-trigger.json")
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("got %v", err)
	}
}

func TestMissingDocumentIsAFault(t *testing.T) {
	r := reader(t, nil)
	if _, err := r.Trigger("none.json"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("got %v", err)
	}
}

func TestTriggerRejectsDuplicateIDsAndSchemaViolations(t *testing.T) {
	dup := `{"version":1,"skill":"s","queries":[
 {"id":"a","prompt":"p1","shouldTrigger":true,"split":"train"},
 {"id":"a","prompt":"p2","shouldTrigger":false,"split":"train"}]}`
	bad := invalidOf(t, func() error {
		_, err := reader(t, map[string]string{"d.json": dup}).Trigger("d.json")
		return err
	}())
	if !strings.Contains(bad.Error(), "duplicate query id a") {
		t.Fatalf("%v", bad)
	}
	bad = invalidOf(t, func() error {
		_, err := reader(t, map[string]string{"d.json": `{"version":1}`}).Trigger("d.json")
		return err
	}())
	if len(bad.Violations) == 0 {
		t.Fatal("want violations")
	}
}

func TestInvalidErrorNamesTheRoot(t *testing.T) {
	e := &document.InvalidError{Path: "f", Violations: []jsonschema.Violation{{Message: "m"}, {Pointer: "/x", Message: "n"}}}
	if got := e.Error(); !strings.Contains(got, "/: m") || !strings.Contains(got, "/x: n") {
		t.Fatal(got)
	}
}

const output = `{"version":1,"skill":"s","cases":[{"id":"c","prompt":"p","gate":"required",
 "fixtures":[{"path":"a.txt","content":"x"},{"path":"b.txt","contentFile":"fx/b.txt"}],
 "expect":[{"kind":"llm","criterion":"ok"}]}]}`

func TestOutputResolvesFixtures(t *testing.T) {
	r := reader(t, map[string]string{"s/eval-output.json": output, "s/fx/b.txt": "bee"})
	doc, err := r.Output("s/eval-output.json")
	if err != nil {
		t.Fatal(err)
	}
	f := doc.Cases[0].Fixtures
	if f[0].Text != "x" || f[1].Text != "bee" {
		t.Fatalf("%+v", f)
	}
}

func TestOutputFaults(t *testing.T) {
	if _, err := reader(t, map[string]string{"s/eval-output.json": output}).Output("s/eval-output.json"); err == nil {
		t.Fatal("missing fixture file is a fault")
	}
	missingPrompt := strings.Replace(output, `"prompt":"p"`, `"promptFile":"nope.md"`, 1)
	if _, err := reader(t, map[string]string{"o.json": missingPrompt}).Output("o.json"); err == nil {
		t.Fatal("missing prompt file is a fault")
	}
	dup := `{"version":1,"skill":"s","cases":[
{"id":"c","prompt":"p","gate":"required","expect":[{"kind":"llm","criterion":"ok"}]},
{"id":"c","prompt":"p","gate":"required","expect":[{"kind":"llm","criterion":"ok"}]}]}`
	invalidOf(t, func() error {
		_, err := reader(t, map[string]string{"o.json": dup}).Output("o.json")
		return err
	}())
}

const instruction = `{"version":1,"instruction":"AGENTS.md","cases":[
 {"id":"c","prompt":"do","anchor":{"heading":"Commits"},"expect":{"present":[{"kind":"llm","criterion":"x"}]}}]}`

func TestInstruction(t *testing.T) {
	doc, err := reader(t, map[string]string{"i.json": instruction}).Instruction("i.json")
	if err != nil || doc.Cases[0].Text != "do" {
		t.Fatalf("%v %+v", err, doc)
	}
	if _, err := reader(t, map[string]string{"i.json": strings.Replace(instruction, `"prompt":"do"`, `"promptFile":"x.md"`, 1)}).Instruction("i.json"); err == nil {
		t.Fatal("missing prompt file")
	}
	dup := strings.Replace(instruction, `]}`, `,{"id":"c","prompt":"p","anchor":{"text":"t"},"expect":{"present":[{"kind":"llm","criterion":"x"}]}}]}`, 1)
	invalidOf(t, func() error {
		_, err := reader(t, map[string]string{"i.json": dup}).Instruction("i.json")
		return err
	}())
	if _, err := reader(t, nil).Instruction("i.json"); err == nil {
		t.Fatal("missing file")
	}
}

const flow = `{"version":1,"id":"f","agents":[{"id":"a","harness":"claude"}],"steps":[
 {"id":"s1","agent":"a","prompt":"one"},
 {"id":"s2","agent":"a","promptFile":"two.md","after":["s1"]}]}`

func TestOrchestration(t *testing.T) {
	doc, err := reader(t, map[string]string{"f.json": flow, "two.md": "two"}).Orchestration("f.json")
	if err != nil || doc.Steps[1].Text != "two" {
		t.Fatalf("%v %+v", err, doc)
	}
	if _, err := reader(t, map[string]string{"f.json": flow}).Orchestration("f.json"); err == nil {
		t.Fatal("missing prompt file")
	}
	if _, err := reader(t, nil).Orchestration("f.json"); err == nil {
		t.Fatal("missing file")
	}
	unknown := strings.Replace(flow, `"after":["s1"]`, `"after":["zz"]`, 1)
	if bad := invalidOf(t, func() error {
		_, err := reader(t, map[string]string{"f.json": unknown, "two.md": "x"}).Orchestration("f.json")
		return err
	}()); !strings.Contains(bad.Error(), "unknown step zz") {
		t.Fatal(bad)
	}
}

func TestFlowProblems(t *testing.T) {
	doc := model.OrchestrationEval{
		Agents: []model.Agent{{ID: "a"}, {ID: "a"}},
		Steps: []model.Step{
			{ID: "x", Agent: "ghost"},
			{ID: "x", Agent: "a"},
		},
	}
	got := strings.Join(document.FlowProblems(doc), "|")
	for _, want := range []string{"duplicate agent id a", "duplicate step id x", "unknown agent ghost"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	cycle := model.OrchestrationEval{
		Agents: []model.Agent{{ID: "a"}},
		Steps:  []model.Step{{ID: "p", Agent: "a", After: []string{"q"}}, {ID: "q", Agent: "a", After: []string{"p"}}},
	}
	if got := document.FlowProblems(cycle); len(got) != 1 || !strings.Contains(got[0], "cycle") {
		t.Fatal(got)
	}
}

func TestOrderKeepsDocumentOrderAmongIndependentSteps(t *testing.T) {
	steps := []model.Step{
		{ID: "c", After: []string{"a"}},
		{ID: "a"},
		{ID: "b"},
	}
	got, err := document.Order(steps)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, s := range got {
		ids = append(ids, s.ID)
	}
	if strings.Join(ids, "") != "abc" {
		t.Fatalf("%v", ids)
	}
	if _, err := document.Order([]model.Step{{ID: "a", After: []string{"a"}}}); !errors.Is(err, document.ErrCycle) {
		t.Fatalf("%v", err)
	}
}

func TestResultsBaselineCostHistoryAndValidate(t *testing.T) {
	digest := "sha256:" + strings.Repeat("0", 64)
	r := reader(t, map[string]string{
		"r.json": `{"version":1,"runId":"r","protocol":"trigger","profile":"p","inputsDigest":"` + digest + `","startedAt":"2026-01-01T00:00:00Z","status":"incomplete","cases":[]}`,
		"b.json": `{"version":1,"protocol":"output","profile":"p","inputsDigest":"` + digest + `","recordedAt":"2026-01-01T00:00:00Z","tolerance":0.1,"cases":[{"id":"c","passRate":1,"runs":1}]}`,
		"c.json": `{"version":1,"entries":[]}`,
	})
	if res, err := r.Results("r.json"); err != nil || res.RunID != "r" {
		t.Fatalf("%v %+v", err, res)
	}
	if b, err := r.Baseline("b.json"); err != nil || b.Cases[0].ID != "c" {
		t.Fatalf("%v %+v", err, b)
	}
	if c, err := r.CostHistory("c.json"); err != nil || c.Version != 1 {
		t.Fatalf("%v %+v", err, c)
	}
	if v, err := r.Validate("https://schemas.xonovex.com/agent/cost-history.schema.json", []byte(`{}`)); err != nil || len(v) == 0 {
		t.Fatalf("%v %v", err, v)
	}
	if _, err := r.Validator.Validate("urn:none", []byte(`{}`)); err == nil {
		t.Fatal("unknown schema is a fault")
	}
	strict := document.Reader{FS: memfs.New(map[string]string{"x.json": "{}"}), Validator: r.Validator}
	if _, err := strict.Results("x.json"); err == nil {
		t.Fatal("violation")
	}
}

func TestUnknownSchemaIsAFault(t *testing.T) {
	r := document.Reader{FS: memfs.New(map[string]string{"x": "{}"}), Validator: failing{}}
	if _, err := r.Results("x"); err == nil || !strings.Contains(err.Error(), "validate x") {
		t.Fatalf("%v", err)
	}
}

type failing struct{}

func (failing) Validate(string, []byte) ([]jsonschema.Violation, error) {
	return nil, jsonschema.ErrUnknownSchema
}
