package provider

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

func lookupEnvVar(env []corev1.EnvVar, name string) (corev1.EnvVar, bool) {
	for _, variable := range env {
		if variable.Name == name {
			return variable, true
		}
	}
	return corev1.EnvVar{}, false
}

func findEnvVar(t testing.TB, env []corev1.EnvVar, name string) corev1.EnvVar {
	t.Helper()
	variable, found := lookupEnvVar(env, name)
	if found {
		return variable
	}
	t.Fatalf("environment variable %q not found", name)
	return corev1.EnvVar{}
}

func envValue(t testing.TB, env []corev1.EnvVar, name string) string {
	t.Helper()
	return findEnvVar(t, env, name).Value
}

func TestResolveProvider_NoProvider(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
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
			Provider: &agentv1alpha1.ProviderSpec{
				Type: "gemini",
				Environment: map[string]string{
					"ANTHROPIC_BASE_URL": "http://proxy:8080",
					"API_TIMEOUT_MS":     "3000000",
				},
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, env, "ANTHROPIC_BASE_URL"); got != "http://proxy:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", got, "http://proxy:8080")
	}
	if got := envValue(t, env, "API_TIMEOUT_MS"); got != "3000000" {
		t.Errorf("API_TIMEOUT_MS = %q, want %q", got, "3000000")
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
			Provider: &agentv1alpha1.ProviderSpec{
				Type: "gemini",
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

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	auth := findEnvVar(t, env, "ANTHROPIC_AUTH_TOKEN")
	if auth.Value != "" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN value = %q, want empty", auth.Value)
	}
	if auth.ValueFrom == nil || auth.ValueFrom.SecretKeyRef == nil || auth.ValueFrom.SecretKeyRef.Name != "api-key" || auth.ValueFrom.SecretKeyRef.Key != "token" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN source = %#v, want api-key/token SecretKeyRef", auth.ValueFrom)
	}
}

func TestResolveProvider_DoesNotMaterializeSecretValue(t *testing.T) {
	const secretValue = "must-not-enter-the-job-spec"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-key", Namespace: "default"},
		Data:       map[string][]byte{"token": []byte(secretValue)},
	}
	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(secret).Build()
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{Provider: &agentv1alpha1.ProviderSpec{
			AuthSecretRef: &agentv1alpha1.SecretKeyRef{Name: secret.Name, Key: "token"},
			Environment:   map[string]string{"ANTHROPIC_BASE_URL": "http://proxy:8080"},
		}},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")

	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	for _, variable := range env {
		if variable.Value == secretValue {
			t.Fatalf("resolved environment materialized secret value in %q", variable.Name)
		}
	}
}

func TestResolveProvider_ProviderRef(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "gemini-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
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
			ProviderRef: "gemini-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, env, "ANTHROPIC_BASE_URL"); got != "http://proxy:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", got, "http://proxy:8080")
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
			ProviderRef: "gemini-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	auth := findEnvVar(t, env, "ANTHROPIC_AUTH_TOKEN")
	if auth.Value != "" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN value = %q, want empty", auth.Value)
	}
	if auth.ValueFrom == nil || auth.ValueFrom.SecretKeyRef == nil || auth.ValueFrom.SecretKeyRef.Name != "provider-secret" || auth.ValueFrom.SecretKeyRef.Key != "api-key" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN source = %#v, want provider-secret/api-key SecretKeyRef", auth.ValueFrom)
	}
}

func TestResolveProvider_ProviderRefNotFound(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "nonexistent",
		},
	}

	_, err := ResolveProvider(context.Background(), c, run, "")
	if err == nil {
		t.Error("ResolveProvider() expected error for nonexistent provider ref")
	}
}

func TestResolveProvider_SecretNotFound(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
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
			ProviderRef: "provider",
		},
	}

	_, err := ResolveProvider(context.Background(), c, run, "")
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
			ProviderRef: "provider",
		},
	}

	_, err := ResolveProvider(context.Background(), c, run, "")
	if err == nil {
		t.Error("ResolveProvider() expected error for missing key in secret")
	}
}

func TestResolveProvider_DefaultFromHarness(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "default-claude-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			Environment: map[string]string{
				"FROM_HARNESS": "true",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
	}

	env, err := ResolveProvider(context.Background(), c, run, "default-claude-provider")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, env, "FROM_HARNESS"); got != "true" {
		t.Errorf("FROM_HARNESS = %q, want %q", got, "true")
	}
}

func TestResolveProvider_PresetRef(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "preset-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: "gemini",
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "preset-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, env, "ANTHROPIC_BASE_URL"); got == "" {
		t.Error("expected ANTHROPIC_BASE_URL from preset, got empty")
	}
}

func TestResolveProvider_PresetRefOverriddenByEnvironment(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "preset-override", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: "gemini",
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://custom:9999",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "preset-override",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, env, "ANTHROPIC_BASE_URL"); got != "http://custom:9999" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", got, "http://custom:9999")
	}
	// Preset's other env vars should still be present
	if got := envValue(t, env, "API_TIMEOUT_MS"); got == "" {
		t.Error("expected API_TIMEOUT_MS from preset, got empty")
	}
}

func TestResolveProvider_UnknownPresetReturnsError(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "unknown-preset", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: "nonexistent-preset",
			Environment: map[string]string{
				"CUSTOM_VAR": "value",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(provider).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "unknown-preset",
		},
	}

	if _, err := ResolveProvider(context.Background(), c, run, ""); err == nil {
		t.Fatal("ResolveProvider() error = nil, want unknown-preset error")
	}
}

func TestResolveProvider_InlinePresetRef(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(newScheme()).Build()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{
				PresetRef: "gemini",
				Environment: map[string]string{
					"ANTHROPIC_BASE_URL": "http://override:8080",
				},
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, env, "ANTHROPIC_BASE_URL"); got != "http://override:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want override", got)
	}
	if got := envValue(t, env, "API_TIMEOUT_MS"); got == "" {
		t.Error("expected API_TIMEOUT_MS from preset")
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
			ProviderRef: "provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	// Without ANTHROPIC_BASE_URL, the token should NOT be injected as ANTHROPIC_AUTH_TOKEN
	if _, has := lookupEnvVar(env, "ANTHROPIC_AUTH_TOKEN"); has {
		t.Error("ANTHROPIC_AUTH_TOKEN should not be set without ANTHROPIC_BASE_URL")
	}
}
