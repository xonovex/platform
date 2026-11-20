package agentcmd

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

var (
	readsEnv   = &types.AgentConfig{Type: "reads", ProviderEnv: true}
	ignoresEnv = &types.AgentConfig{Type: "ignores"}
)

func hostWith(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestPlanProviderEnvWithNoProviderPlansNothing(t *testing.T) {
	plan := PlanProviderEnv(readsEnv, nil)

	if len(plan.Values) != 0 || plan.Credential != (CredentialRef{}) {
		t.Fatalf("PlanProviderEnv(nil) = %+v, want an empty plan", plan)
	}
	env, err := plan.Resolve(hostWith(nil))
	if err != nil || env == nil || len(env) != 0 {
		t.Fatalf("Resolve() = %v, %v, want an empty map", env, err)
	}
}

func TestPlanProviderEnvReferencesTheCredentialWithoutItsValue(t *testing.T) {
	provider := &types.ModelProvider{
		Name:                "route",
		Environment:         map[string]string{"BASE_URL": "https://models.example.test"},
		CredentialSourceEnv: "HOST_KEY",
		CredentialTargetEnv: "AGENT_KEY",
	}

	plan := PlanProviderEnv(readsEnv, provider)

	if plan.Provider != "route" || plan.Values["BASE_URL"] != "https://models.example.test" || len(plan.Values) != 1 {
		t.Fatalf("plan = %+v, want the provider environment", plan)
	}
	if plan.Credential != (CredentialRef{Env: "AGENT_KEY", SourceEnv: "HOST_KEY"}) {
		t.Fatalf("plan credential = %+v, want the reference of the provider credential", plan.Credential)
	}
	plan.Values["BASE_URL"] = "changed"
	if provider.Environment["BASE_URL"] != "https://models.example.test" {
		t.Error("PlanProviderEnv shares the environment of the provider")
	}
}

func TestPlanProviderEnvForAnAgentThatReadsNoProviderEnvironmentPlansNothing(t *testing.T) {
	provider := &types.ModelProvider{Name: "route", Environment: map[string]string{"A": "1"}, CredentialSourceEnv: "S", CredentialTargetEnv: "T"}

	plan := PlanProviderEnv(ignoresEnv, provider)

	if len(plan.Values) != 0 || plan.Credential != (CredentialRef{}) {
		t.Fatalf("plan = %+v, want no variable and no credential", plan)
	}
}

func TestPlanProviderEnvOfAProviderWithNoEnvironmentHasAnEmptyMap(t *testing.T) {
	plan := PlanProviderEnv(readsEnv, &types.ModelProvider{Name: "bare"})

	if plan.Values == nil || len(plan.Values) != 0 {
		t.Fatalf("Values = %v, want an empty map", plan.Values)
	}
}

func TestResolveReadsTheCredentialFromTheHost(t *testing.T) {
	plan := ProviderEnvPlan{Provider: "route", Values: map[string]string{"A": "1"}, Credential: CredentialRef{Env: "AGENT_KEY", SourceEnv: "HOST_KEY"}}

	env, err := plan.Resolve(hostWith(map[string]string{"HOST_KEY": "secret"}))

	if err != nil || env["AGENT_KEY"] != "secret" || env["A"] != "1" {
		t.Fatalf("Resolve() = %v, %v", env, err)
	}
	if _, set := plan.Values["AGENT_KEY"]; set {
		t.Error("Resolve() wrote the credential into the plan")
	}
}

func TestResolveOfAPlanWithNoValuesReturnsAMap(t *testing.T) {
	env, err := ProviderEnvPlan{}.Resolve(hostWith(nil))

	if err != nil || env == nil {
		t.Fatalf("Resolve() = %v, %v, want an empty map", env, err)
	}
}

func TestResolveReportsAMissingCredential(t *testing.T) {
	plan := ProviderEnvPlan{Provider: "route", Credential: CredentialRef{Env: "AGENT_KEY", SourceEnv: "HOST_KEY"}}

	_, err := plan.Resolve(hostWith(nil))

	if err == nil || !strings.Contains(err.Error(), "failed to build provider environment") || !strings.Contains(err.Error(), "HOST_KEY") {
		t.Fatalf("Resolve() err = %v, want the missing host variable reported", err)
	}
}

func TestResolveRefusesAHalfCredential(t *testing.T) {
	for _, credential := range []CredentialRef{{Env: "AGENT_KEY"}, {SourceEnv: "HOST_KEY"}} {
		plan := ProviderEnvPlan{Provider: "half", Credential: credential}

		_, err := plan.Resolve(hostWith(map[string]string{"HOST_KEY": "x"}))

		if err == nil || !strings.Contains(err.Error(), "both") || !strings.Contains(err.Error(), `"half"`) {
			t.Errorf("Resolve(%+v) err = %v, want both variables required", credential, err)
		}
	}
}
