package providers

import (
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/types"
)

var geminiOpencodeProvider = &types.ModelProvider{
	Name:         "gemini",
	DisplayName:  "Google Gemini",
	AgentType:    types.AgentOpencode,
	AuthTokenEnv: "",
	Environment:  map[string]string{},
	CliArgs:      []string{"--model", "google/gemini-2.5-pro"},
}
