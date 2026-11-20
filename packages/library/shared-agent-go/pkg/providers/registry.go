package providers

import (
	"fmt"
	"maps"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// Getenv resolves an environment variable and returns the empty string when it is unset.
type Getenv func(key string) string

// presets returns the preset constructors of an agent type by name. A new agent
// type adds one file that defines its presets and one case here.
func presets(agentType types.AgentType) map[string]func() *types.ModelProvider {
	switch agentType {
	case types.AgentClaude:
		return claudePresets()
	case types.AgentCodex:
		return codexPresets()
	case types.AgentOpencode:
		return opencodePresets()
	default:
		return nil
	}
}

// GetProvider retrieves a fresh provider configuration by name and agent type.
func GetProvider(name string, agentType types.AgentType) (*types.ModelProvider, error) {
	build, ok := presets(agentType)[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s for agent %s", name, agentType)
	}
	provider := build()
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

// GetProviderNames returns all provider names for an agent type in sorted order.
func GetProviderNames(agentType types.AgentType) []string {
	names := slices.AppendSeq([]string{}, maps.Keys(presets(agentType)))
	slices.Sort(names)
	return names
}

// BuildProviderEnv builds environment variables from provider config, reading
// the credential through getenv.
func BuildProviderEnv(provider *types.ModelProvider, getenv Getenv) (map[string]string, error) {
	env := maps.Clone(provider.Environment)
	if env == nil {
		env = make(map[string]string)
	}

	if provider.CredentialSourceEnv != "" || provider.CredentialTargetEnv != "" {
		if provider.CredentialSourceEnv == "" || provider.CredentialTargetEnv == "" {
			return nil, fmt.Errorf("provider %q must define both credential source and target environment variables", provider.Name)
		}
		authToken := getenv(provider.CredentialSourceEnv)
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
