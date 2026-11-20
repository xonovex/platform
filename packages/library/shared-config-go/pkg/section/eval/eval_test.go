package eval_test

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/section/eval"
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
