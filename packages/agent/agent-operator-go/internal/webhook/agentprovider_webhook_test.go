package webhook

import (
	"context"
	"testing"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestAgentProviderWebhook_Validate_Valid(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Type: "anthropic",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "my-secret",
				Key:  "token",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "https://api.anthropic.com",
			},
			CliArgs: []string{"--model", "claude-sonnet-4-6"},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_Empty(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_NilSecretRef(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Type: "anthropic",
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_EmptySecretName(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "",
				Key:  "token",
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err == nil {
		t.Fatal("expected error for empty secret name")
	}
}

func TestAgentProviderWebhook_Validate_EmptySecretKey(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "my-secret",
				Key:  "",
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err == nil {
		t.Fatal("expected error for empty secret key")
	}
}

func TestAgentProviderWebhook_Validate_InvalidSecretName(t *testing.T) {
	w := &AgentProviderWebhook{}
	invalid := []string{"MySecret", "my_secret", "MY-SECRET", "my secret", ".hidden"}
	for _, name := range invalid {
		provider := &agentv1alpha1.AgentProvider{
			Spec: agentv1alpha1.AgentProviderSpec{
				AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
					Name: name,
					Key:  "token",
				},
			},
		}
		_, err := w.ValidateCreate(context.Background(), provider)
		if err == nil {
			t.Errorf("expected error for invalid secret name %q", name)
		}
	}
}

func TestAgentProviderWebhook_Validate_ValidEnvKeys(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Environment: map[string]string{
				"MY_VAR":    "value",
				"_PRIVATE":  "value",
				"SOME_VAR2": "value",
			},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_InvalidEnvKey(t *testing.T) {
	w := &AgentProviderWebhook{}
	invalid := []string{"1INVALID", "my-var", "my.var", "my var", ""}
	for _, key := range invalid {
		provider := &agentv1alpha1.AgentProvider{
			Spec: agentv1alpha1.AgentProviderSpec{
				Environment: map[string]string{key: "value"},
			},
		}
		_, err := w.ValidateCreate(context.Background(), provider)
		if err == nil {
			t.Errorf("expected error for invalid env key %q", key)
		}
	}
}

func TestAgentProviderWebhook_Validate_BlockedEnvKey(t *testing.T) {
	w := &AgentProviderWebhook{}
	blocked := []string{
		"LD_PRELOAD", "LD_LIBRARY_PATH",
		"DYLD_INSERT_LIBRARIES",
		"PYTHONPATH",
		"RUBYOPT",
		"NODE_OPTIONS",
		"JAVA_TOOL_OPTIONS",
	}
	for _, key := range blocked {
		provider := &agentv1alpha1.AgentProvider{
			Spec: agentv1alpha1.AgentProviderSpec{
				Environment: map[string]string{key: "value"},
			},
		}
		_, err := w.ValidateCreate(context.Background(), provider)
		if err == nil {
			t.Errorf("expected error for blocked env key %q", key)
		}
	}
}

func TestAgentProviderWebhook_Validate_ValidCliArgs(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			CliArgs: []string{"--model", "claude-sonnet-4-6", "--max-tokens", "4096"},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err != nil {
		t.Errorf("ValidateCreate() error = %v", err)
	}
}

func TestAgentProviderWebhook_Validate_EmptyCliArg(t *testing.T) {
	w := &AgentProviderWebhook{}
	provider := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			CliArgs: []string{"--model", ""},
		},
	}

	_, err := w.ValidateCreate(context.Background(), provider)
	if err == nil {
		t.Fatal("expected error for empty cliArg")
	}
}

func TestAgentProviderWebhook_Validate_InjectionCliArg(t *testing.T) {
	w := &AgentProviderWebhook{}
	malicious := []string{
		"--model; rm -rf /",
		"--arg | cat /etc/passwd",
		"--arg$(whoami)",
		"--arg`id`",
		"--arg && curl evil.com",
	}
	for _, arg := range malicious {
		provider := &agentv1alpha1.AgentProvider{
			Spec: agentv1alpha1.AgentProviderSpec{
				CliArgs: []string{arg},
			},
		}
		_, err := w.ValidateCreate(context.Background(), provider)
		if err == nil {
			t.Errorf("expected error for malicious cliArg %q", arg)
		}
	}
}

func TestAgentProviderWebhook_ValidateUpdate(t *testing.T) {
	w := &AgentProviderWebhook{}
	old := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{Type: "anthropic"},
	}
	updated := &agentv1alpha1.AgentProvider{
		Spec: agentv1alpha1.AgentProviderSpec{
			Type: "openai",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "new-secret",
				Key:  "api-key",
			},
		},
	}

	_, err := w.ValidateUpdate(context.Background(), old, updated)
	if err != nil {
		t.Errorf("ValidateUpdate() error = %v", err)
	}
}

func TestAgentProviderWebhook_ValidateDelete(t *testing.T) {
	w := &AgentProviderWebhook{}
	_, err := w.ValidateDelete(context.Background(), &agentv1alpha1.AgentProvider{})
	if err != nil {
		t.Errorf("ValidateDelete() error = %v", err)
	}
}
