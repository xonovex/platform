package spec_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

func TestAllows(t *testing.T) {
	key := spec.Key{Layers: spec.TrustedFiles}
	for layer, want := range map[spec.Layer]bool{
		spec.Managed: true, spec.User: true, spec.Flags: true, spec.Project: false, spec.Default: false,
	} {
		if got := key.Allows(layer); got != want {
			t.Errorf("Allows(%s) = %v", layer, got)
		}
	}
}

func TestName(t *testing.T) {
	cases := map[string]spec.Key{
		"eval.budgets.maxModelRuns": {Section: "eval", Pointer: "/budgets/maxModelRuns"},
		"providers":                 {Section: "providers"},
	}
	for want, key := range cases {
		if got := key.Name(); got != want {
			t.Errorf("Name = %q, want %q", got, want)
		}
	}
}

func TestSectionKey(t *testing.T) {
	section := spec.Section{Keys: []spec.Key{{Pointer: "/a"}, {Pointer: "/b", Doc: "b"}}}
	if key, ok := section.Key("/b"); !ok || key.Doc != "b" {
		t.Fatalf("Key(/b) = %+v, %v", key, ok)
	}
	if _, ok := section.Key("/c"); ok {
		t.Fatal("Key(/c) must not exist")
	}
}
