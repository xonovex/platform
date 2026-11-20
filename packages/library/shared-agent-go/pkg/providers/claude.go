package providers

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// claudePresets returns the preset constructors for the Claude agent by name.
func claudePresets() map[string]func() *types.ModelProvider {
	return map[string]func() *types.ModelProvider{
		"gemini":        geminiProvider,
		"gemini-claude": geminiClaudeProvider,
		"glm":           glmProvider,
		"gpt5-codex":    gpt5CodexProvider,
	}
}

func geminiProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "gemini",
		DisplayName:         "Google Gemini",
		AgentType:           types.AgentClaude,
		CredentialSourceEnv: "CLI_PROXY_API_KEY",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Portable:            false,
		Environment: map[string]string{
			"ANTHROPIC_BASE_URL":                       "http://127.0.0.1:8317",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "gemini-4-carbon",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "gemini-3.8-flash",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "gemini-3.5-flash-lite",
			"CLAUDE_CODE_SUBAGENT_MODEL":               "gemini-3.8-flash",
		},
		CliArgs: []string{},
	}
}

func geminiClaudeProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "gemini-claude",
		DisplayName:         "Gemini-Claude Thinking",
		AgentType:           types.AgentClaude,
		CredentialSourceEnv: "CLI_PROXY_API_KEY",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Portable:            false,
		Environment: map[string]string{
			"ANTHROPIC_BASE_URL":                       "http://127.0.0.1:8317",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "gemini-claude-opus-5-5-thinking",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "gemini-claude-sonnet-5-5-thinking",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "gemini-3.5-flash-lite",
			"CLAUDE_CODE_SUBAGENT_MODEL":               "gemini-claude-sonnet-5-5",
		},
		CliArgs: []string{},
	}
}

func glmProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "glm",
		DisplayName:         "Zhipu AI GLM",
		AgentType:           types.AgentClaude,
		CredentialSourceEnv: "ZAI_AUTH_TOKEN",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Portable:            true,
		Environment: map[string]string{
			"ANTHROPIC_BASE_URL":                       "https://api.z.ai/api/anthropic",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "glm-5.3",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "glm-5.3",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "glm-5.3-flash",
		},
		CliArgs: []string{},
	}
}

func gpt5CodexProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "gpt5-codex",
		DisplayName:         "OpenAI GPT-6.2 Sol",
		AgentType:           types.AgentClaude,
		CredentialSourceEnv: "CLI_PROXY_API_KEY",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Portable:            false,
		Environment: map[string]string{
			"ANTHROPIC_BASE_URL":                       "http://127.0.0.1:8317",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "gpt-6.2-sol(high)",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "gpt-6.2-sol(medium)",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "gpt-6.2-sol(low)",
			"CLAUDE_CODE_SUBAGENT_MODEL":               "gpt-6.2-sol(medium)",
		},
		CliArgs: []string{},
	}
}
