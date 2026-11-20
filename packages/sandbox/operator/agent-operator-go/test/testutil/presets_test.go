package testutil

import (
	"slices"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestNewPresetsDefinesTheDocumentedProviders(t *testing.T) {
	presets := NewPresets()

	if got := presets.Names(types.AgentClaude); !slices.Equal(got, []string{PresetLocal, PresetPortable}) {
		t.Errorf("claude providers = %v", got)
	}
	args, err := presets.ResolvePortable(PresetArgs, types.AgentOpencode)
	if err != nil || !slices.Equal(args.CliArgs, []string{"--model", PresetArgsModel}) {
		t.Errorf("args provider = %+v, %v", args, err)
	}
	if _, err := presets.ResolvePortable(PresetLocal, types.AgentClaude); err == nil {
		t.Error("the local provider is portable")
	}
	var empty providers.Catalog
	if len(empty.List()) != 0 {
		t.Error("the zero catalog defines a provider")
	}
}
