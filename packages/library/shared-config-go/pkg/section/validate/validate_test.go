package validate_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/validate"
)

func TestEveryKeyBelongsToTheSection(t *testing.T) {
	section := validate.Section()
	if section.Name != validate.Name || section.Gated != false || len(section.Keys) == 0 {
		t.Fatalf("section = %+v", section)
	}
	for _, key := range section.Keys {
		if key.Section != validate.Name {
			t.Errorf("%s names the section %q", key.Name(), key.Section)
		}
	}
}
