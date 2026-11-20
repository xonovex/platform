package agents

import (
	"fmt"
	"slices"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// BuildArgs builds the arguments of an agent from its declared facts: the launch
// arguments, the sandbox arguments when the launch runs inside an isolation, the
// provider arguments when the agent takes them, then the caller's arguments.
func BuildArgs(agent *types.AgentConfig, baseArgs []string, options types.AgentExecOptions) []string {
	args := slices.Clone(agent.LaunchArgs)
	if options.Sandbox {
		args = append(args, agent.SandboxArgs...)
	}
	if agent.ProviderArgs {
		args = append(args, options.ProviderCliArgs...)
	}
	return append(args, baseArgs...)
}

// BuildEnv returns the part of the provider environment that the agent reads,
// as a new map.
func BuildEnv(agent *types.AgentConfig, providerEnv map[string]string) map[string]string {
	env := make(map[string]string)
	if !agent.ProviderEnv {
		return env
	}
	for key, value := range providerEnv {
		env[key] = value
	}
	return env
}

// TakesProvider reports whether the agent obtains model access from a provider
// preset, and not from the login of its own vendor.
func TakesProvider(agent *types.AgentConfig) bool {
	return agent.Access == types.AccessProvider
}

// CheckRoutes fails when the provider sets a route that the agent does not read,
// because the agent would drop it silently and run without the endpoint or
// credential the provider names.
func CheckRoutes(agent *types.AgentConfig, provider *types.ModelProvider) error {
	if provider == nil {
		return nil
	}
	if !TakesProvider(agent) {
		return fmt.Errorf("agent %s uses its own login and takes no provider", agent.Type)
	}
	hasEnv := len(provider.Environment) > 0 || provider.CredentialTargetEnv != ""
	if hasEnv && !agent.ProviderEnv {
		return fmt.Errorf("provider %s sets environment that agent %s does not read", provider.Name, agent.Type)
	}
	if len(provider.CliArgs) > 0 && !agent.ProviderArgs {
		return fmt.Errorf("provider %s sets arguments that agent %s does not take", provider.Name, agent.Type)
	}
	return nil
}
