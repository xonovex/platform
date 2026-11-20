package model_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

func TestEncodeEndsWithNewlineAndKeepsHTML(t *testing.T) {
	got, err := model.Encode(map[string]string{"a": "<b>&"})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{\n  \"a\": \"<b>&\"\n}\n" {
		t.Fatalf("got %q", got)
	}
}

func TestEncodeReportsAnUnencodableValue(t *testing.T) {
	if _, err := model.Encode(make(chan int)); err == nil {
		t.Fatal("want an error")
	}
}

func TestDecodeReportsBadJSON(t *testing.T) {
	var v model.TriggerEval
	err := model.Decode([]byte("{"), &v)
	if err == nil || !strings.Contains(err.Error(), "decode document") {
		t.Fatalf("got %v", err)
	}
}

func TestTriggerEvalRoundTripsFlattenedPrompt(t *testing.T) {
	in := `{"version":1,"skill":"s","queries":[{"id":"q1","prompt":"add a route","shouldTrigger":true,"split":"train","rationale":"why"},{"id":"q2","promptFile":"p.md","shouldTrigger":false,"split":"validation"}]}`
	var doc model.TriggerEval
	if err := model.Decode([]byte(in), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Queries[0].Prompt.Prompt != "add a route" || doc.Queries[1].PromptFile != "p.md" {
		t.Fatalf("%+v", doc.Queries)
	}
	out, err := model.Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	var again model.TriggerEval
	if err := model.Decode(out, &again); err != nil {
		t.Fatal(err)
	}
	if again.Queries[1].PromptFile != "p.md" || again.Queries[0].Rationale != "why" {
		t.Fatalf("%+v", again)
	}
}

func TestUsageAdd(t *testing.T) {
	got := model.Usage{InputTokens: 1, OutputTokens: 2, CostUSD: 0.5}.Add(model.Usage{InputTokens: 3, OutputTokens: 4, CostUSD: 0.25})
	if got != (model.Usage{InputTokens: 4, OutputTokens: 6, CostUSD: 0.75}) {
		t.Fatalf("%+v", got)
	}
}

func TestExpectationsForAndConditions(t *testing.T) {
	p := model.Predicate{Kind: model.KindLLM, Criterion: "c"}
	c := model.InstructionCase{Expect: model.Expectations{Present: []model.Predicate{p}, Ablated: []model.Predicate{p}}}
	if got := c.Conditions(); len(got) != 2 || got[0] != model.Present || got[1] != model.Ablated {
		t.Fatalf("%v", got)
	}
	if len(c.Expect.For(model.Absent)) != 0 || len(c.Expect.For(model.Condition("x"))) != 0 {
		t.Fatal("absent and unknown hold no predicates")
	}
}

func TestCaseKeyAndDone(t *testing.T) {
	r := model.Results{Cases: []model.CaseResult{{ID: "a"}, {ID: "b", Condition: model.Absent}}}
	done := r.Done()
	if _, ok := done["a"]; !ok {
		t.Fatal("missing a")
	}
	if _, ok := done["b@absent"]; !ok {
		t.Fatal("missing b@absent")
	}
	if model.CaseKey("x", "") != "x" {
		t.Fatal("no condition keeps the id")
	}
}

func TestProtocolsListsEveryProtocol(t *testing.T) {
	if len(model.Protocols()) != 5 {
		t.Fatal("five protocols")
	}
}

func TestASandboxIsUnconfinedOnlyWhenItAsksForNothing(t *testing.T) {
	if !(model.Sandbox{}).Unconfined() || !(model.Sandbox{Isolation: "none", Provision: "none", Network: "host"}).Unconfined() {
		t.Error("no sandbox, or every axis at its unconfined value, runs the harness on the host")
	}
	for name, sandbox := range map[string]model.Sandbox{
		"isolation":   {Isolation: "docker"},
		"provision":   {Provision: "nix"},
		"network":     {Network: "none"},
		"image":       {Image: "registry.example/eval:1"},
		"runtime":     {Runtime: "runsc"},
		"passthrough": {HostPassthrough: true},
		"init":        {InitCommands: []string{"true"}},
		"nix package": {NixPackages: []string{"git"}},
		"nix rev":     {NixRev: "0123"},
		"allow":       {AllowHosts: []string{"api.example.com"}},
		"credential":  {Credentials: []string{"KEY=env:KEY"}},
		"env":         {CustomEnv: []string{"MODE=ci"}},
		"pinned":      {RequirePinnedProvision: true},
		"hidden host": {RequireHostToolsUnreachable: true},
		"egress":      {RequireEgressRestricted: true},
		"kernel":      {RequireKernelIsolation: true},
	} {
		if sandbox.Unconfined() {
			t.Errorf("%s: a sandbox that asks for %s is not unconfined", name, name)
		}
	}
}
