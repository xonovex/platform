package audit_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/audit"
)

func TestEveryKeyBelongsToTheSection(t *testing.T) {
	section := audit.Section()
	if section.Name != audit.Name || section.Gated != false || len(section.Keys) == 0 {
		t.Fatalf("section = %+v", section)
	}
	for _, key := range section.Keys {
		if key.Section != audit.Name {
			t.Errorf("%s names the section %q", key.Name(), key.Section)
		}
	}
}
