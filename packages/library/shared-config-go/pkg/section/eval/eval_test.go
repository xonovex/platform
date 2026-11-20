package eval_test

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/spec"
)

func TestEveryKeyBelongsToTheSection(t *testing.T) {
	section := eval.Section()
	if section.Name != eval.Name || section.Gated != true || len(section.Keys) == 0 {
		t.Fatalf("section = %+v", section)
	}
	for _, key := range section.Keys {
		if key.Section != eval.Name {
			t.Errorf("%s names the section %q", key.Name(), key.Section)
		}
	}
}

func TestOnlyTheTrustedFilesChooseTheSandbox(t *testing.T) {
	section := eval.Section()
	demands := map[string]bool{
		"/sandbox/requirePinnedProvision": true, "/sandbox/requireHostToolsUnreachable": true,
		"/sandbox/requireEgressRestricted": true, "/sandbox/requireKernelIsolation": true,
	}
	sandbox := 0
	for _, key := range section.Keys {
		if !strings.HasPrefix(key.Pointer, "/sandbox/") {
			continue
		}
		sandbox++
		if demands[key.Pointer] {
			if key.Rule != spec.Require || !key.Allows(spec.Project) {
				t.Errorf("%s: a project demands a guarantee, which every layer keeps: %+v", key.Name(), key)
			}
			continue
		}
		if key.Allows(spec.Project) {
			t.Errorf("%s: a project file may not choose the confinement or what it reaches", key.Name())
		}
	}
	if sandbox != 16 {
		t.Errorf("%d sandbox keys, want 16", sandbox)
	}
}
