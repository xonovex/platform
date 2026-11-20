package harness_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

type stub struct{ name string }

func (s stub) Name() string { return s.name }
func (stub) Binary() string { return "b" }
func (stub) Plan(harness.Request, string) (harness.Plan, error) {
	return harness.Plan{}, nil
}
func (stub) Scan(harness.Request) harness.Scanner { return nil }

var _ = process.Result{}

func TestRegistryGetAndNames(t *testing.T) {
	r := harness.NewRegistry()
	r.Register(stub{"b"})
	r.Register(stub{"a"})
	if got := r.Names(); strings.Join(got, ",") != "a,b" {
		t.Fatal(got)
	}
	if h, err := r.Get("a"); err != nil || h.Name() != "a" {
		t.Fatalf("%v %v", h, err)
	}
	if _, err := r.Get("z"); err == nil || !strings.Contains(err.Error(), "supported: [a b]") {
		t.Fatalf("%v", err)
	}
}

func TestRegisterTwicePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("want a panic")
		}
	}()
	r := harness.NewRegistry()
	r.Register(stub{"a"})
	r.Register(stub{"a"})
}
