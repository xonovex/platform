package observe_test

import (
	"context"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
)

func TestHash(t *testing.T) {
	const empty = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := observe.Hash(nil); got != empty {
		t.Fatalf("Hash(nil) = %s", got)
	}
}

func TestAggregate(t *testing.T) {
	a := observe.Document{Locator: "https://a.test", Status: observe.OK, SHA256: observe.Hash([]byte("a"))}
	b := observe.Document{Locator: "https://b.test", Status: observe.OK, SHA256: observe.Hash([]byte("b"))}

	one, ok := observe.Aggregate([]observe.Document{a})
	if !ok || !strings.HasPrefix(one, observe.DigestPrefix) || len(one) != len(observe.DigestPrefix)+64 {
		t.Fatalf("Aggregate(a) = %q, %v", one, ok)
	}
	ab, _ := observe.Aggregate([]observe.Document{a, b})
	ba, _ := observe.Aggregate([]observe.Document{b, a})
	if ab == ba || ab == one {
		t.Fatal("the digest must depend on the order and on every document")
	}

	for name, docs := range map[string][]observe.Document{
		"no documents":   nil,
		"missing":        {a, {Locator: "x", Status: observe.Missing}},
		"restricted":     {{Locator: "x", Status: observe.Restricted}},
		"no body digest": {{Locator: "x", Status: observe.OK}},
	} {
		if got, ok := observe.Aggregate(docs); ok || got != "" {
			t.Errorf("%s: Aggregate = %q, %v; want no digest", name, got, ok)
		}
	}
}

type named struct{ name string }

func (n named) Name() string                  { return n.name }
func (named) Remote() bool                    { return false }
func (n named) Applies(s sources.Source) bool { return s.Kind == sources.Kind(n.name) }
func (named) Observe(context.Context, observe.Request) (observe.Observation, error) {
	return observe.Observation{}, nil
}

func TestRegistryReturnsTheApplicableTransportsInOrder(t *testing.T) {
	registry := observe.NewRegistry().Register(named{"web"}).Register(named{"git"}).Register(named{"web"})
	got := registry.For(sources.Source{Kind: sources.Web})
	if len(got) != 2 || got[0].Name() != "web" {
		t.Fatalf("For(web) = %v", got)
	}
	if got := registry.For(sources.Source{Kind: sources.Original}); len(got) != 0 {
		t.Fatalf("For(original) = %v; want none", got)
	}
}
