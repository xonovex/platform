package run_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/run"
)

func TestEveryKeyBelongsToTheSection(t *testing.T) {
	section := run.Section()
	if section.Name != run.Name || section.Gated != true || len(section.Keys) == 0 {
		t.Fatalf("section = %+v", section)
	}
	for _, key := range section.Keys {
		if key.Section != run.Name {
			t.Errorf("%s names the section %q", key.Name(), key.Section)
		}
	}
}
