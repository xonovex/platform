package provider

import (
	"context"
	"errors"
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	sharedtypes "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

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

// claudeEnv renders the provider environment of route for the claude harness,
// which reads the provider environment.
func claudeEnv(t testing.TB, route Route) []corev1.EnvVar {
	t.Helper()
	agent, err := agents.GetAgent(sharedtypes.AgentClaude)
	if err != nil {
		t.Fatalf("harness facts: %v", err)
	}
	return Environment(agent, route)
}

func envValue(t testing.TB, env []corev1.EnvVar, name string) string {
	t.Helper()
	return findEnvVar(t, env, name).Value
}

func TestResolveProvider_NoProvider(t *testing.T) {
	c := testutil.NewMemoryStore()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if len(claudeEnv(t, env)) != 0 {
		t.Errorf("env = %v, want empty", env)
	}
}

func TestResolveProvider_InlineProvider(t *testing.T) {
	c := testutil.NewMemoryStore()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{
				Environment: map[string]string{
					"ANTHROPIC_BASE_URL": "http://proxy:8080",
					"API_TIMEOUT_MS":     "3000000",
				},
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, claudeEnv(t, env), "ANTHROPIC_BASE_URL"); got != "http://proxy:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", got, "http://proxy:8080")
	}
	if got := envValue(t, claudeEnv(t, env), "API_TIMEOUT_MS"); got != "3000000" {
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

	c := testutil.NewMemoryStore(secret)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{
				AuthTokenEnv: "ANTHROPIC_AUTH_TOKEN",
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

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	auth := findEnvVar(t, claudeEnv(t, env), "ANTHROPIC_AUTH_TOKEN")
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
	c := testutil.NewMemoryStore(secret)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{Provider: &agentv1alpha1.ProviderSpec{
			AuthSecretRef: &agentv1alpha1.SecretKeyRef{Name: secret.Name, Key: "token"},
			AuthTokenEnv:  "ANTHROPIC_AUTH_TOKEN",
			Environment:   map[string]string{"ANTHROPIC_BASE_URL": "http://proxy:8080"},
		}},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")

	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	for _, variable := range claudeEnv(t, env) {
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

	c := testutil.NewMemoryStore(provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "gemini-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, claudeEnv(t, env), "ANTHROPIC_BASE_URL"); got != "http://proxy:8080" {
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
			AuthTokenEnv: "ANTHROPIC_AUTH_TOKEN",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "provider-secret",
				Key:  "api-key",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
			},
		},
	}

	c := testutil.NewMemoryStore(secret, provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "gemini-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	auth := findEnvVar(t, claudeEnv(t, env), "ANTHROPIC_AUTH_TOKEN")
	if auth.Value != "" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN value = %q, want empty", auth.Value)
	}
	if auth.ValueFrom == nil || auth.ValueFrom.SecretKeyRef == nil || auth.ValueFrom.SecretKeyRef.Name != "provider-secret" || auth.ValueFrom.SecretKeyRef.Key != "api-key" {
		t.Errorf("ANTHROPIC_AUTH_TOKEN source = %#v, want provider-secret/api-key SecretKeyRef", auth.ValueFrom)
	}
}

func TestResolveProvider_ProviderRefNotFound(t *testing.T) {
	c := testutil.NewMemoryStore()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "nonexistent",
		},
	}

	_, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err == nil {
		t.Error("ResolveProvider() expected error for nonexistent provider ref")
	}
}

func TestResolveProvider_SecretNotFound(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AuthTokenEnv: "ANTHROPIC_AUTH_TOKEN",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "nonexistent-secret",
				Key:  "token",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
			},
		},
	}

	c := testutil.NewMemoryStore(provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "provider",
		},
	}

	_, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
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
			AuthTokenEnv: "ANTHROPIC_AUTH_TOKEN",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "my-secret",
				Key:  "expected-key",
			},
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://proxy:8080",
			},
		},
	}

	c := testutil.NewMemoryStore(secret, provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "provider",
		},
	}

	_, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
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

	c := testutil.NewMemoryStore(provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "default-claude-provider")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, claudeEnv(t, env), "FROM_HARNESS"); got != "true" {
		t.Errorf("FROM_HARNESS = %q, want %q", got, "true")
	}
}

func TestResolveProvider_PresetRef(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "preset-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: testutil.PresetPortable,
		},
	}

	c := testutil.NewMemoryStore(provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "preset-provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, claudeEnv(t, env), "ANTHROPIC_BASE_URL"); got == "" {
		t.Error("expected ANTHROPIC_BASE_URL from preset, got empty")
	}
}

func TestResolveProvider_PresetRefOverriddenByEnvironment(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "preset-override", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: testutil.PresetPortable,
			Environment: map[string]string{
				"ANTHROPIC_BASE_URL": "http://custom:9999",
			},
		},
	}

	c := testutil.NewMemoryStore(provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "preset-override",
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, claudeEnv(t, env), "ANTHROPIC_BASE_URL"); got != "http://custom:9999" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want %q", got, "http://custom:9999")
	}
	// Preset's other env vars should still be present
	if got := envValue(t, claudeEnv(t, env), "API_TIMEOUT_MS"); got == "" {
		t.Error("expected API_TIMEOUT_MS from preset, got empty")
	}
}

func TestResolveProvider_UnknownPresetReturnsError(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "unknown-preset", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: "not-defined",
			Environment: map[string]string{
				"CUSTOM_VAR": "value",
			},
		},
	}

	c := testutil.NewMemoryStore(provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "unknown-preset",
		},
	}

	if _, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, ""); err == nil {
		t.Fatal("ResolveProvider() error = nil, want unknown-preset error")
	}
}

func TestResolveProvider_InlinePresetRef(t *testing.T) {
	c := testutil.NewMemoryStore()

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{
				PresetRef: testutil.PresetPortable,
				Environment: map[string]string{
					"ANTHROPIC_BASE_URL": "http://override:8080",
				},
			},
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	if got := envValue(t, claudeEnv(t, env), "ANTHROPIC_BASE_URL"); got != "http://override:8080" {
		t.Errorf("ANTHROPIC_BASE_URL = %q, want override", got)
	}
	if got := envValue(t, claudeEnv(t, env), "API_TIMEOUT_MS"); got == "" {
		t.Error("expected API_TIMEOUT_MS from preset")
	}
}

func TestResolveProvider_InjectsAuthTokenIntoExplicitDestinationWithoutBaseURL(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-key", Namespace: "default"},
		Data: map[string][]byte{
			"token": []byte("secret-token"),
		},
	}
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			AuthTokenEnv: "CUSTOM_PROVIDER_TOKEN",
			AuthTokenSecretRef: &agentv1alpha1.SecretKeyRef{
				Name: "api-key",
				Key:  "token",
			},
			Environment: map[string]string{
				"SOME_OTHER_VAR": "value",
			},
		},
	}

	c := testutil.NewMemoryStore(secret, provider)

	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			ProviderRef: "provider",
		},
	}

	env, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")
	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	auth := findEnvVar(t, claudeEnv(t, env), "CUSTOM_PROVIDER_TOKEN")
	if auth.ValueFrom == nil || auth.ValueFrom.SecretKeyRef == nil {
		t.Fatal("CUSTOM_PROVIDER_TOKEN must reference the configured Secret")
	}
}

func TestResolveProvider_MergesPresetAndExplicitCliArgs(t *testing.T) {
	provider := &agentv1alpha1.AgentProvider{
		ObjectMeta: metav1.ObjectMeta{Name: "opencode-provider", Namespace: "default"},
		Spec: agentv1alpha1.AgentProviderSpec{
			PresetRef: testutil.PresetArgs,
			AgentType: "opencode",
			CliArgs:   []string{"--model", "custom/model"},
		},
	}
	c := testutil.NewMemoryStore(provider)
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{ProviderRef: provider.Name},
	}

	resolved, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")

	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	want := []string{"--model", testutil.PresetArgsModel, "--model", "custom/model"}
	if !slices.Equal(resolved.Provider.CliArgs, want) {
		t.Errorf("CliArgs = %v, want %v", resolved.Provider.CliArgs, want)
	}
}

func TestResolveProvider_InlinePresetContributesCliArgs(t *testing.T) {
	c := testutil.NewMemoryStore()
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{Provider: &agentv1alpha1.ProviderSpec{
			PresetRef: testutil.PresetArgs,
			AgentType: "opencode",
		}},
	}

	resolved, err := ResolveProvider(context.Background(), c, testutil.NewPresets(), run, "")

	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	want := []string{"--model", testutil.PresetArgsModel}
	if !slices.Equal(resolved.Provider.CliArgs, want) {
		t.Errorf("CliArgs = %v, want %v", resolved.Provider.CliArgs, want)
	}
}

func TestResolveProvider_PresetRefWithoutAnyDefinedProviderFailsClearly(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{PresetRef: testutil.PresetPortable},
		},
	}

	_, err := ResolveProvider(context.Background(), testutil.NewMemoryStore(), providers.Catalog{}, run, "")

	if !errors.Is(err, providers.ErrNotDefined) {
		t.Errorf("ResolveProvider() error = %v, want it to wrap ErrNotDefined", err)
	}
}

func TestResolveProvider_LocalOnlyPresetIsRefused(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec: agentv1alpha1.AgentRunSpec{
			Provider: &agentv1alpha1.ProviderSpec{PresetRef: testutil.PresetLocal},
		},
	}

	_, err := ResolveProvider(context.Background(), testutil.NewMemoryStore(), testutil.NewPresets(), run, "")

	if !errors.Is(err, providers.ErrLocalOnly) {
		t.Errorf("ResolveProvider() error = %v, want it to wrap ErrLocalOnly", err)
	}
}

func TestResolveProvider_BuildsTheRouteOfTheSharedPlan(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
		Spec:       agentv1alpha1.AgentRunSpec{Provider: &agentv1alpha1.ProviderSpec{PresetRef: "portable-route"}},
	}

	route, err := ResolveProvider(context.Background(), testutil.NewMemoryStore(), testutil.NewPresets(), run, "")

	if err != nil {
		t.Fatalf("ResolveProvider() error = %v", err)
	}
	p := route.Provider
	if p.Name != "portable-route" || p.AgentType != sharedtypes.AgentClaude || !p.Portable {
		t.Errorf("route = %+v, want the portable-route preset for claude", p)
	}
	if p.CredentialTargetEnv != "" || p.CredentialSourceEnv != "" || route.AuthSecret != nil {
		t.Errorf("credential = %q from %q, secret %v, want none: no Secret holds the credential", p.CredentialTargetEnv, p.CredentialSourceEnv, route.AuthSecret)
	}
}

func TestResolveProvider_NamesTheProviderOfARouteAndAnUnnamedInlineRoute(t *testing.T) {
	stored := &agentv1alpha1.AgentProvider{ObjectMeta: metav1.ObjectMeta{Name: "team", Namespace: "default"}}
	reader := testutil.NewMemoryStore(stored)
	byRef := &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: agentv1alpha1.AgentRunSpec{ProviderRef: "team"}}
	inline := &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}, Spec: agentv1alpha1.AgentRunSpec{Provider: &agentv1alpha1.ProviderSpec{}}}

	referenced, err := ResolveProvider(context.Background(), reader, testutil.NewPresets(), byRef, "")
	if err != nil || referenced.Provider.Name != "team" {
		t.Errorf("referenced route = %+v, %v, want the name team", referenced.Provider, err)
	}
	unnamed, err := ResolveProvider(context.Background(), reader, testutil.NewPresets(), inline, "")
	if err != nil || unnamed.Provider.Name != "inline" {
		t.Errorf("inline route = %+v, %v, want the name inline", unnamed.Provider, err)
	}
}

func TestEnvironment_RendersNothingForAnAgentThatReadsNoProviderEnvironment(t *testing.T) {
	agent, err := agents.GetAgent(sharedtypes.AgentOpencode)
	if err != nil {
		t.Fatalf("harness facts: %v", err)
	}
	route := Route{
		Provider:   &sharedtypes.ModelProvider{Name: "p", Environment: map[string]string{"A": "1"}, CredentialTargetEnv: "TOKEN"},
		AuthSecret: &agentv1alpha1.SecretKeyRef{Name: "s", Key: "k"},
	}

	if env := Environment(agent, route); len(env) != 0 {
		t.Errorf("Environment() = %v, want none", env)
	}
}

func TestEnvironment_RendersACredentialTargetOnlyFromASecret(t *testing.T) {
	route := Route{Provider: &sharedtypes.ModelProvider{Name: "p", CredentialTargetEnv: "TOKEN"}}

	if env := claudeEnv(t, route); len(env) != 0 {
		t.Errorf("Environment() = %v, want none: no Secret holds the credential", env)
	}
}
