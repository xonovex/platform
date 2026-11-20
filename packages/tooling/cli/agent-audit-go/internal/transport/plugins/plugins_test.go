package plugins_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/plugins"
)

func TestDefaultRegistryHoldsEveryTransport(t *testing.T) {
	registry := plugins.DefaultRegistry(plugins.Ports{})
	cases := map[string]struct {
		source sources.Source
		want   []string
	}{
		"path":     {sources.Source{Path: "a"}, []string{"filesystem"}},
		"checkout": {sources.Source{Checkout: "c"}, []string{"git"}},
		"urls":     {sources.Source{URLs: []string{"https://e.test"}}, []string{"http"}},
		"all":      {sources.Source{Path: "a", Checkout: "c", URLs: []string{"https://e.test"}}, []string{"filesystem", "git", "http"}},
		"none":     {sources.Source{Provenance: "here"}, nil},
	}
	for name, tc := range cases {
		got := registry.For(tc.source)
		if len(got) != len(tc.want) {
			t.Fatalf("%s: %d transports; want %v", name, len(got), tc.want)
		}
		for i, tr := range got {
			if tr.Name() != tc.want[i] {
				t.Errorf("%s: transport %d = %s; want %s", name, i, tr.Name(), tc.want[i])
			}
		}
	}
}
