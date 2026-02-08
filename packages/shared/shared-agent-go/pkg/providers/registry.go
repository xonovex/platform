package providers

import (
	"fmt"
	"os"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// providerRegistry uses composite keys: "agentType:providerName"
var providerRegistry = map[string]*types.ModelProvider{
	// Claude providers
	"claude:gemini":        geminiProvider,
	"claude:gemini-claude": geminiClaudeProvider,
	"claude:glm":           glmProvider,
	"claude:gpt5-codex":    gpt5CodexProvider,
	// OpenCode providers
	"opencode:gemini": geminiOpencodeProvider,
}

// GetProvider retrieves a provider by name and agent type
func GetProvider(name string, agentType types.AgentType) (*types.ModelProvider, error) {
	// Try direct lookup with composite key
	key := string(agentType) + ":" + name
	if provider, ok := providerRegistry[key]; ok {
		return provider, nil
	}

	// Search all providers for matching name (fallback)
	for _, provider := range providerRegistry {
		if provider.Name == name && provider.AgentType == agentType {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("unknown provider: %s for agent %s", name, agentType)
}

// GetProviderNames returns all provider names for an agent type
func GetProviderNames(agentType types.AgentType) []string {
	names := make([]string, 0)
	for name, provider := range providerRegistry {
		if provider.AgentType == agentType {
			names = append(names, name)
		}
	}
	return names
}

// BuildProviderEnv builds environment variables from provider config
func BuildProviderEnv(provider *types.ModelProvider) (map[string]string, error) {
	// Copy all environment variables from provider
	env := make(map[string]string)
	for k, v := range provider.Environment {
		env[k] = v
	}

	// If auth token is required, inject it
	if provider.AuthTokenEnv != "" {
		authToken := os.Getenv(provider.AuthTokenEnv)
		if authToken == "" {
			return nil, fmt.Errorf("missing authentication token: %s environment variable is not set", provider.AuthTokenEnv)
		}

		// Inject auth token for Anthropic-compatible providers (Claude)
		if _, hasBaseURL := env["ANTHROPIC_BASE_URL"]; hasBaseURL {
			env["ANTHROPIC_AUTH_TOKEN"] = authToken
		}
	}

	return env, nil
}

// GetProviderCliArgs gets CLI arguments for a provider
func GetProviderCliArgs(provider *types.ModelProvider) []string {
	return provider.CliArgs
}
