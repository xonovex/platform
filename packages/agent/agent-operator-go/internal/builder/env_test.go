package builder

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestBuildEnvVars_Empty(t *testing.T) {
	run := &agentv1alpha1.AgentRun{}
	envVars := BuildEnvVars(run, nil)

	if len(envVars) != 0 {
		t.Errorf("len(envVars) = %d, want 0", len(envVars))
	}
}

func TestBuildEnvVars_ProviderOnly(t *testing.T) {
	run := &agentv1alpha1.AgentRun{}
	providerEnv := map[string]string{
		"KEY1": "val1",
		"KEY2": "val2",
	}

	envVars := BuildEnvVars(run, providerEnv)

	envMap := make(map[string]string)
	for _, e := range envVars {
		envMap[e.Name] = e.Value
	}

	if envMap["KEY1"] != "val1" {
		t.Errorf("KEY1 = %q, want %q", envMap["KEY1"], "val1")
	}
	if envMap["KEY2"] != "val2" {
		t.Errorf("KEY2 = %q, want %q", envMap["KEY2"], "val2")
	}
}

func TestBuildEnvVars_SpecEnv(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Env: []corev1.EnvVar{
				{Name: "CUSTOM", Value: "value"},
			},
		},
	}

	envVars := BuildEnvVars(run, nil)

	if len(envVars) != 1 {
		t.Fatalf("len(envVars) = %d, want 1", len(envVars))
	}
	if envVars[0].Name != "CUSTOM" || envVars[0].Value != "value" {
		t.Errorf("env var = %v, want {CUSTOM value}", envVars[0])
	}
}

func TestBuildEnvVars_CombinesProviderAndSpec(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Env: []corev1.EnvVar{
				{Name: "SPEC_VAR", Value: "from-spec"},
			},
		},
	}
	providerEnv := map[string]string{
		"PROVIDER_VAR": "from-provider",
	}

	envVars := BuildEnvVars(run, providerEnv)

	envMap := make(map[string]string)
	for _, e := range envVars {
		envMap[e.Name] = e.Value
	}

	if envMap["PROVIDER_VAR"] != "from-provider" {
		t.Errorf("PROVIDER_VAR = %q, want %q", envMap["PROVIDER_VAR"], "from-provider")
	}
	if envMap["SPEC_VAR"] != "from-spec" {
		t.Errorf("SPEC_VAR = %q, want %q", envMap["SPEC_VAR"], "from-spec")
	}
}
