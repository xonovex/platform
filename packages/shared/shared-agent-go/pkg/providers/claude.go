package providers

import (
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func geminiProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "gemini",
		DisplayName:         "Google Gemini 3.x",
		AgentType:           types.AgentClaude,
		CredentialSourceEnv: "CLI_PROXY_API_KEY",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Portable:            false,
		Environment: map[string]string{
			"ANTHROPIC_BASE_URL":                       "http://127.0.0.1:8317",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "gemini-3-pro-preview",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "gemini-3-flash-preview",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "gemini-2.5-flash-lite",
			"CLAUDE_CODE_SUBAGENT_MODEL":               "gemini-3-flash-preview",
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
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "gemini-claude-opus-4-5-thinking",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "gemini-claude-sonnet-4-5-thinking",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "gemini-2.5-flash-lite",
			"CLAUDE_CODE_SUBAGENT_MODEL":               "gemini-claude-sonnet-4-5",
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
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "GLM-4.7",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "GLM-4.7",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "GLM-4.5-Air",
		},
		CliArgs: []string{},
	}
}

func gpt5CodexProvider() *types.ModelProvider {
	return &types.ModelProvider{
		Name:                "gpt5-codex",
		DisplayName:         "OpenAI GPT-5.2 Codex",
		AgentType:           types.AgentClaude,
		CredentialSourceEnv: "CLI_PROXY_API_KEY",
		CredentialTargetEnv: "ANTHROPIC_AUTH_TOKEN",
		Portable:            false,
		Environment: map[string]string{
			"ANTHROPIC_BASE_URL":                       "http://127.0.0.1:8317",
			"API_TIMEOUT_MS":                           "3000000",
			"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":             "gpt-5.2-codex(high)",
			"ANTHROPIC_DEFAULT_SONNET_MODEL":           "gpt-5.2-codex(medium)",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "gpt-5.2-codex(low)",
			"CLAUDE_CODE_SUBAGENT_MODEL":               "gpt-5.2-codex(medium)",
		},
		CliArgs: []string{},
	}
}
