package session_test

import (
	"context"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/session"
)

func n(v int) *int { return &v }

func TestCounts(t *testing.T) {
	subject := predicate.Subject{Obs: harness.Observation{Events: []harness.Event{
		{Kind: harness.EventToolCall, Name: "Skill"},
		{Kind: harness.EventFileRead, Text: "SKILL.md"},
		{Kind: harness.EventFileRead, Text: "ref/a.md"},
		{Kind: harness.EventCommand, Text: "rm -rf x"},
	}}}
	cases := []struct {
		name string
		p    model.Predicate
		want bool
	}{
		{"default needs one", model.Predicate{Event: "file-read", Match: `SKILL`}, true},
		{"default none", model.Predicate{Event: "file-read", Match: `nope`}, false},
		{"tool name", model.Predicate{Event: "tool-call", Match: `^Skill$`}, true},
		{"min two", model.Predicate{Event: "file-read", Match: `.`, Min: n(2)}, true},
		{"min three", model.Predicate{Event: "file-read", Match: `.`, Min: n(3)}, false},
		{"max zero holds", model.Predicate{Event: "command", Match: `curl`, Max: n(0)}, true},
		{"max zero breaks", model.Predicate{Event: "command", Match: `rm`, Max: n(0)}, false},
		{"bad pattern", model.Predicate{Event: "command", Match: `(`}, false},
	}
	e := session.New()
	if e.Kind() != "session" {
		t.Fatal(e.Kind())
	}
	for _, c := range cases {
		got, _, err := e.Evaluate(context.Background(), []model.Predicate{c.p}, subject)
		if err != nil || got[0].Passed != c.want {
			t.Errorf("%s: %+v %v", c.name, got, err)
		}
		if !got[0].Passed && got[0].Detail == "" {
			t.Errorf("%s: a failure explains itself", c.name)
		}
	}
}
