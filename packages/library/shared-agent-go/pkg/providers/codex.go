package providers

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// codexPresets returns the preset constructors for the Codex agent by name.
func codexPresets() map[string]func() *types.ModelProvider {
	return map[string]func() *types.ModelProvider{
		"gpt5-codex": gpt5CodexCodexProvider,
	}
}

// gpt5CodexCodexProvider defines the provider on the command line with the
// Responses wire API, so the run depends on no config.toml. Codex reads the key
// from the variable that env_key names, so the key never appears in an argument.
func gpt5CodexCodexProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "gpt5-codex",
		DisplayName:         "OpenAI GPT-6.2 Sol",
		AgentType:           types.AgentCodex,
		CredentialSourceEnv: "CLI_PROXY_API_KEY",
		CredentialTargetEnv: "CLI_PROXY_API_KEY",
		Portable:            false,
		Environment:         map[string]string{},
		CliArgs: []string{
			"-c", `model_provider="proxy"`,
			"-c", `model_providers.proxy.name="CLI proxy"`,
			"-c", `model_providers.proxy.base_url="http://127.0.0.1:8317/v1"`,
			"-c", `model_providers.proxy.wire_api="responses"`,
			"-c", `model_providers.proxy.env_key="CLI_PROXY_API_KEY"`,
			"-c", `model="gpt-6.2-sol"`,
		},
	}
}
