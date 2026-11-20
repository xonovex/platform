package filesystem_test

import (
	"context"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate/filesystem"
)

func TestChecks(t *testing.T) {
	subject := predicate.Subject{Obs: harness.Observation{Files: map[string]string{"a/b.md": "hello world"}}}
	cases := []struct {
		name string
		p    model.Predicate
		want bool
	}{
		{"exists", model.Predicate{Path: "a/b.md", Check: "exists"}, true},
		{"exists clean", model.Predicate{Path: "./a/b.md", Check: "exists"}, true},
		{"exists missing", model.Predicate{Path: "x", Check: "exists"}, false},
		{"absent", model.Predicate{Path: "x", Check: "absent"}, true},
		{"absent present", model.Predicate{Path: "a/b.md", Check: "absent"}, false},
		{"equals", model.Predicate{Path: "a/b.md", Check: "equals", Value: "hello world"}, true},
		{"equals differs", model.Predicate{Path: "a/b.md", Check: "equals", Value: "x"}, false},
		{"contains", model.Predicate{Path: "a/b.md", Check: "contains", Value: "lo wo"}, true},
		{"contains none", model.Predicate{Path: "a/b.md", Check: "contains", Value: "zzz"}, false},
		{"contains empty", model.Predicate{Path: "a/b.md", Check: "contains", Value: ""}, true},
		{"contains missing file", model.Predicate{Path: "x", Check: "contains", Value: "a"}, false},
		{"matches", model.Predicate{Path: "a/b.md", Check: "matches", Value: "^hello"}, true},
		{"matches no", model.Predicate{Path: "a/b.md", Check: "matches", Value: "^world"}, false},
		{"bad pattern", model.Predicate{Path: "a/b.md", Check: "matches", Value: "("}, false},
		{"unknown", model.Predicate{Path: "a/b.md", Check: "weird"}, false},
	}
	e := filesystem.New()
	if e.Kind() != "filesystem" {
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
