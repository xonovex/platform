package builder

import (
	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// BuildEnvVars builds environment variables for the agent container
func BuildEnvVars(run *agentv1alpha1.AgentRun, providerEnv map[string]string) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	// Add provider environment variables
	for k, v := range providerEnv {
		envVars = append(envVars, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	// Add spec environment variables (these override provider env)
	envVars = append(envVars, run.Spec.Env...)

	return envVars
}
