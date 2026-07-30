package providers

import (
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// GetProvider retrieves a fresh provider configuration by name and agent type.
func GetProvider(name string, agentType types.AgentType) (*types.ModelProvider, error) {
	var provider *types.ModelProvider
	switch {
	case agentType == types.AgentClaude && name == "gemini":
		provider = geminiProvider()
	case agentType == types.AgentClaude && name == "gemini-claude":
		provider = geminiClaudeProvider()
	case agentType == types.AgentClaude && name == "glm":
		provider = glmProvider()
	case agentType == types.AgentClaude && name == "gpt5-codex":
		provider = gpt5CodexProvider()
	case agentType == types.AgentOpencode && name == "gemini":
		provider = geminiOpencodeProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %s for agent %s", name, agentType)
	}
	provider.Environment = maps.Clone(provider.Environment)
	provider.CliArgs = slices.Clone(provider.CliArgs)
	return provider, nil
}

// GetPortableProvider retrieves a preset that is safe to use in a remote workload.
func GetPortableProvider(name string, agentType types.AgentType) (*types.ModelProvider, error) {
	provider, err := GetProvider(name, agentType)
	if err != nil {
		return nil, err
	}
	if !provider.Portable {
		return nil, fmt.Errorf("provider preset %q for agent %q is local-only", name, agentType)
	}
	return provider, nil
}

// GetProviderNames returns all provider names for an agent type.
func GetProviderNames(agentType types.AgentType) []string {
	switch agentType {
	case types.AgentClaude:
		return []string{"gemini", "gemini-claude", "glm", "gpt5-codex"}
	case types.AgentOpencode:
		return []string{"gemini"}
	default:
		return []string{}
	}
}

// BuildProviderEnv builds environment variables from provider config
func BuildProviderEnv(provider *types.ModelProvider) (map[string]string, error) {
	env := maps.Clone(provider.Environment)
	if env == nil {
		env = make(map[string]string)
	}

	if provider.CredentialSourceEnv != "" || provider.CredentialTargetEnv != "" {
		if provider.CredentialSourceEnv == "" || provider.CredentialTargetEnv == "" {
			return nil, fmt.Errorf("provider %q must define both credential source and target environment variables", provider.Name)
		}
		authToken := os.Getenv(provider.CredentialSourceEnv)
		if authToken == "" {
			return nil, fmt.Errorf("missing authentication token: %s environment variable is not set", provider.CredentialSourceEnv)
		}
		env[provider.CredentialTargetEnv] = authToken
	}

	return env, nil
}

// GetProviderCliArgs gets CLI arguments for a provider
func GetProviderCliArgs(provider *types.ModelProvider) []string {
	return slices.Clone(provider.CliArgs)
}
