package providers

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// opencodePresets returns the preset constructors for the OpenCode agent by name.
func opencodePresets() map[string]func() *types.ModelProvider {
	return map[string]func() *types.ModelProvider{
		"gemini": geminiOpencodeProvider,
	}
}

func geminiOpencodeProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:        "gemini",
		DisplayName: "Google Gemini",
		AgentType:   types.AgentOpencode,
		Portable:    true,
		Environment: map[string]string{},
		CliArgs:     []string{"--model", "google/gemini-4-carbon"},
	}
}
