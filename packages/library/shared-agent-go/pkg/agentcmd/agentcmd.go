// Package agentcmd plans the environment that a provider route gives an agent.
// The plan holds no credential value, so an executor that only references the
// credential, such as a Kubernetes Job, and one that reads it from the host
// share one plan.
package agentcmd

import (
	"fmt"
	"maps"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

// CredentialRef names the variable that receives the credential of a provider
// and the host variable that holds it. It holds no credential value.
type CredentialRef struct {
	// Env is the variable of the agent that receives the credential.
	Env string
	// SourceEnv is the host variable that holds the credential. An executor that
	// takes the credential from a store of its own ignores it.
	SourceEnv string
}

// ProviderEnvPlan is the environment a provider route gives an agent, before
// any credential is read.
type ProviderEnvPlan struct {
	// Provider is the name of the provider route, for the messages of Resolve.
	Provider string
	// Values are the plain variables the agent reads, by name.
	Values map[string]string
	// Credential is the credential of the provider; the zero value names none.
	Credential CredentialRef
}

// PlanProviderEnv plans the environment of provider for agent. A run with no
// provider, and an agent that reads no provider environment, get a plan with no
// variable and no credential.
func PlanProviderEnv(agent *types.AgentConfig, provider *types.ModelProvider) ProviderEnvPlan {
	if provider == nil || !agent.ProviderEnv {
		return ProviderEnvPlan{Values: map[string]string{}}
	}
	values := maps.Clone(provider.Environment)
	if values == nil {
		values = map[string]string{}
	}
	return ProviderEnvPlan{
		Provider:   provider.Name,
		Values:     values,
		Credential: CredentialRef{Env: provider.CredentialTargetEnv, SourceEnv: provider.CredentialSourceEnv},
	}
}

// Resolve returns the environment of the plan with the credential read from
// the host through getenv. It fails on a credential that names only one of its
// two variables, and on a host variable that is empty.
func (p ProviderEnvPlan) Resolve(getenv providers.Getenv) (map[string]string, error) {
	env := maps.Clone(p.Values)
	if env == nil {
		env = map[string]string{}
	}
	credential := p.Credential
	if credential == (CredentialRef{}) {
		return env, nil
	}
	if credential.Env == "" || credential.SourceEnv == "" {
		return nil, fmt.Errorf("failed to build provider environment: provider %q must define both credential source and target environment variables", p.Provider)
	}
	value := getenv(credential.SourceEnv)
	if value == "" {
		return nil, fmt.Errorf("failed to build provider environment: missing authentication token: %s environment variable is not set", credential.SourceEnv)
	}
	env[credential.Env] = value
	return env, nil
}
