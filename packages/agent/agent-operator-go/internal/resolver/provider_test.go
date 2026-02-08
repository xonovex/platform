package resolver

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = agentv1alpha1.AddToScheme(s)
	return s
}

func TestResolveProvider_NoProvider(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, nil)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if len(env) != 0 {
		t.Errorf("env = %v, want empty", env)
	}
}

func TestResolveProvider_InlineProvider(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
			Provider: &agentv1alpha1.ProviderSpec{
				Name: "gemini",
				Environment: map[string]string{
					"ANTHROPIC_BASE_URL": "http://proxy:8080",
					"API_TIMEOUT_MS":     "3000000",
				},
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, nil)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if env["ANTHROPIC_BASE_URL"] != "http://proxy:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", env["ANTHROPIC_BASE_URL"], "http://proxy:8080")
	}
	if env["API_TIMEOUT_MS"] != "3000000" {
		t.Errorf("API_TIMEOUT_MS = %q, want %q", env["API_TIMEOUT_MS"], "3000000")
	}
}

func TestResolveProvider_InlineProviderWithSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-key", Namespace: "default"},
		Data: map[string][]byte{
			"token": []byte("secret-token-123"),
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(secret).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
			Provider: &agentv1alpha1.ProviderSpec{
				Name: "gemini",
				AuthSecretRef: &agentv1alpha1.SecretKeyRef{
					Name: "api-key",
					Key:  "token",
				},
				Environment: map[string]string{
					"ANTHROPIC_BASE_URL": "http://proxy:8080",
				},
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, nil)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "secret-token-123" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", env["ANTHROPIC_AUTH_TOKEN"], "secret-token-123")
	}
}

func TestResolveProvider_ProviderRef(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "gemini-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
				"API_TIMEOUT_MS":     "3000000",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:       agentv1alpha1.AgentTypeClaude,
			ProviderRef: "gemini-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, nil)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if env["ANTHROPIC_BASE_URL"] != "http://proxy:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", env["ANTHROPIC_BASE_URL"], "http://proxy:8080")
	}
}

func TestResolveProvider_ProviderRefWithSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "provider-secret", Namespace: "default"},
		Data: map[string][]byte{
			"api-key": []byte("my-secret-key"),
		},
	}
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "gemini-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "provider-secret",
				Key:  "api-key",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(secret, provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:       agentv1alpha1.AgentTypeClaude,
			ProviderRef: "gemini-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, nil)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "my-secret-key" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN = %q, want %q", env["ANTHROPIC_AUTH_TOKEN"], "my-secret-key")
	}
}

func TestResolveProvider_ProviderRefNotFound(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:       agentv1alpha1.AgentTypeClaude,
			ProviderRef: "nonexistent",
		},
	}

	_, err := ResolveProvider(context.Background(), c, run, nil)
	if err == nil {
		t.Error("ResolveProvider() expected error for nonexistent provider ref")
	}
}

func TestResolveProvider_SecretNotFound(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "nonexistent-secret",
				Key:  "token",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:       agentv1alpha1.AgentTypeClaude,
			ProviderRef: "provider",
		},
	}

	_, err := ResolveProvider(context.Background(), c, run, nil)
	if err == nil {
		t.Error("ResolveProvider() expected error for missing secret")
	}
}

func TestResolveProvider_SecretMissingKey(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-secret", Namespace: "default"},
		Data: map[string][]byte{
			"wrong-key": []byte("value"),
		},
	}
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "my-secret",
				Key:  "expected-key",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(secret, provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:       agentv1alpha1.AgentTypeClaude,
			ProviderRef: "provider",
		},
	}

	_, err := ResolveProvider(context.Background(), c, run, nil)
	if err == nil {
		t.Error("ResolveProvider() expected error for missing key in secret")
	}
}

func TestResolveProvider_DefaultFromConfig(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "default-claude-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
			Environment: map[string]string{
				"FROM_CONFIG": "true",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent: agentv1alpha1.AgentTypeClaude,
		},
	}
	config := &agentv1alpha1.AgentConfig{
		Spec: agentv1alpha1.AgentConfigSpec{
			DefaultProviders: map[agentv1alpha1.AgentType]string{
				agentv1alpha1.AgentTypeClaude: "default-claude-provider",
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, config)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if env["FROM_CONFIG"] != "true" {
		t.Errorf("FROM_CONFIG = %q, want %q", env["FROM_CONFIG"], "true")
	}
}

func TestResolveProvider_NoAuthTokenInjectionWithoutBaseURL(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-key", Namespace: "default"},
		Data: map[string][]byte{
			"token": []byte("secret-token"),
		},
	}
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AgentTypes: []agentv1alpha1.AgentType{agentv1alpha1.AgentTypeClaude},
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "api-key",
				Key:  "token",
			},
			Environment: map[string]string{
				"SOME_OTHER_VAR": "value",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(secret, provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Agent:       agentv1alpha1.AgentTypeClaude,
			ProviderRef: "provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, nil)
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	// Without ANTHROPIC_BASE_URL, the token should NOT be injected as ANTHROPIC_AUTH_TOKEN
	if _, has := env["ANTHROPIC_AUTH_TOKEN"]; has {
		t.Error("ANTHROPIC_AUTH_TOKEN should not be set without ANTHROPIC_BASE_URL")
	}
}
