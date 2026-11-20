package providers

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
)

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name      string
		agentType types.AgentType
	}{
		{name: "gemini", agentType: types.AgentClaude},
		{name: "gemini-claude", agentType: types.AgentClaude},
		{name: "glm", agentType: types.AgentClaude},
		{name: "gpt5-codex", agentType: types.AgentClaude},
		{name: "gemini", agentType: types.AgentOpencode},
		{name: "gpt5-codex", agentType: types.AgentCodex},
	}
	for _, test := range tests {
		t.Run(string(test.agentType)+"/"+test.name, func(t *testing.T) {
			provider, err := GetProvider(test.name, test.agentType)

			if err != nil {
				t.Fatalf("GetProvider() error = %v", err)
			}
			if provider.Name != test.name || provider.AgentType != test.agentType {
				t.Fatalf("GetProvider() = %+v, want %s provider %q", provider, test.agentType, test.name)
			}
		})
	}
}

func TestGetProviderReturnsIndependentConfigurations(t *testing.T) {
	first, err := GetProvider("gemini", types.AgentClaude)
	if err != nil {
		t.Fatalf("GetProvider() error = %v", err)
	}
	first.Environment["ANTHROPIC_BASE_URL"] = "changed"

	second, err := GetProvider("gemini", types.AgentClaude)
	if err != nil {
		t.Fatalf("GetProvider() error = %v", err)
	}
	if second.Environment["ANTHROPIC_BASE_URL"] == "changed" {
		t.Fatal("GetProvider() returned shared mutable environment")
	}
}

func TestGetPortableProviderRejectsLocalOnlyPreset(t *testing.T) {
	provider, err := GetPortableProvider("gemini", types.AgentClaude)
	if err == nil || provider != nil {
		t.Fatalf("GetPortableProvider() = (%+v, %v), want local-only error", provider, err)
	}

	provider, err = GetPortableProvider("glm", types.AgentClaude)
	if err != nil || provider == nil {
		t.Fatalf("GetPortableProvider() = (%+v, %v), want GLM provider", provider, err)
	}
}

func TestGetProviderRejectsUnknownCombination(t *testing.T) {
	provider, err := GetProvider("glm", types.AgentOpencode)
	if err == nil || provider != nil {
		t.Fatalf("GetProvider() = (%+v, %v), want nil provider and error", provider, err)
	}
}

func TestGetProviderNamesReturnsSortedProviderNames(t *testing.T) {
	tests := []struct {
		agentType types.AgentType
		want      []string
	}{
		{agentType: types.AgentClaude, want: []string{"gemini", "gemini-claude", "glm", "gpt5-codex"}},
		{agentType: types.AgentCodex, want: []string{"gpt5-codex"}},
		{agentType: types.AgentOpencode, want: []string{"gemini"}},
		{agentType: types.AgentPi, want: []string{}},
		{agentType: types.AgentCopilot, want: []string{}},
		{agentType: types.AgentKiro, want: []string{}},
		{agentType: types.AgentType("unknown"), want: []string{}},
	}
	for _, test := range tests {
		t.Run(string(test.agentType), func(t *testing.T) {
			got := GetProviderNames(test.agentType)

			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("GetProviderNames() = %v, want %v", got, test.want)
			}
		})
	}
}

func mapGetenv(env map[string]string) Getenv {
	return func(key string) string { return env[key] }
}

func TestBuildProviderEnvInjectsAuthenticationTokenIntoCopy(t *testing.T) {
	provider := geminiProvider()
	env, err := BuildProviderEnv(provider, mapGetenv(map[string]string{"CLI_PROXY_API_KEY": "secret"}))
	if err != nil {
		t.Fatalf("BuildProviderEnv() error = %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "secret" {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN = %q, want secret", env["ANTHROPIC_AUTH_TOKEN"])
	}
	env["ANTHROPIC_BASE_URL"] = "changed"
	if provider.Environment["ANTHROPIC_BASE_URL"] == "changed" {
		t.Fatal("BuildProviderEnv() mutated provider environment")
	}
}

func TestBuildProviderEnvRejectsMissingAuthenticationToken(t *testing.T) {
	_, err := BuildProviderEnv(glmProvider(), mapGetenv(nil))
	if err == nil || !strings.Contains(err.Error(), "ZAI_AUTH_TOKEN") {
		t.Fatalf("BuildProviderEnv() error = %v, want missing-token error", err)
	}
}

func TestBuildProviderEnvRejectsIncompleteCredentialMapping(t *testing.T) {
	provider := &types.ModelProvider{
		Name:                "incomplete",
		CredentialSourceEnv: "SOURCE_TOKEN",
	}

	_, err := BuildProviderEnv(provider, mapGetenv(nil))

	if err == nil {
		t.Fatal("BuildProviderEnv() error = nil, want incomplete credential mapping error")
	}
}

func TestBuildProviderEnvSupportsCredentialWithoutPresetEnvironment(t *testing.T) {
	provider := &types.ModelProvider{
		Name:                "custom",
		CredentialSourceEnv: "CUSTOM_SOURCE_TOKEN",
		CredentialTargetEnv: "CUSTOM_TARGET_TOKEN",
	}

	env, err := BuildProviderEnv(provider, mapGetenv(map[string]string{"CUSTOM_SOURCE_TOKEN": "secret"}))

	if err != nil {
		t.Fatalf("BuildProviderEnv() error = %v", err)
	}
	if env["CUSTOM_TARGET_TOKEN"] != "secret" {
		t.Fatalf("CUSTOM_TARGET_TOKEN = %q, want secret", env["CUSTOM_TARGET_TOKEN"])
	}
}

func TestBuildProviderEnvAllowsProviderWithoutAuthenticationToken(t *testing.T) {
	env, err := BuildProviderEnv(geminiOpencodeProvider(), mapGetenv(nil))
	if err != nil {
		t.Fatalf("BuildProviderEnv() error = %v", err)
	}
	if len(env) != 0 {
		t.Fatalf("BuildProviderEnv() = %v, want empty environment", env)
	}
}

func TestGetProviderCliArgs(t *testing.T) {
	provider := geminiOpencodeProvider()
	got := GetProviderCliArgs(provider)
	want := []string{"--model", "google/gemini-4-carbon"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProviderCliArgs() = %v, want %v", got, want)
	}
	got[0] = "changed"
	if provider.CliArgs[0] == "changed" {
		t.Fatal("GetProviderCliArgs() returned shared mutable arguments")
	}
}

func TestCodexPresetKeepsTheKeyOutOfArguments(t *testing.T) {
	provider, err := GetProvider("gpt5-codex", types.AgentCodex)
	if err != nil {
		t.Fatalf("GetProvider() error = %v", err)
	}
	env, err := BuildProviderEnv(provider, func(key string) string {
		if key == "CLI_PROXY_API_KEY" {
			return "secret"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("BuildProviderEnv() error = %v", err)
	}
	if env["CLI_PROXY_API_KEY"] != "secret" {
		t.Fatalf("env = %v, want the key in the variable env_key names", env)
	}
	for _, arg := range provider.CliArgs {
		if strings.Contains(arg, "secret") {
			t.Fatalf("argument %q holds the key", arg)
		}
	}
	if !slices.Contains(provider.CliArgs, `model_providers.proxy.env_key="CLI_PROXY_API_KEY"`) {
		t.Fatalf("CliArgs = %v, want an env_key argument that names the credential variable", provider.CliArgs)
	}
}
