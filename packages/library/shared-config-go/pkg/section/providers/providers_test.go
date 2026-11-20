package providers_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/providers"
)

func TestEveryKeyBelongsToTheSection(t *testing.T) {
	section := providers.Section()
	if section.Name != providers.Name || section.Gated != true || len(section.Keys) == 0 {
		t.Fatalf("section = %+v", section)
	}
	for _, key := range section.Keys {
		if key.Section != providers.Name {
			t.Errorf("%s names the section %q", key.Name(), key.Section)
		}
	}
}
