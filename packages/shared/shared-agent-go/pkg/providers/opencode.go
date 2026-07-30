package providers

import (
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func geminiOpencodeProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:        "gemini",
		DisplayName: "Google Gemini",
		AgentType:   types.AgentOpencode,
		Portable:    true,
		Environment: map[string]string{},
		CliArgs:     []string{"--model", "google/gemini-2.5-pro"},
	}
}
