package providers

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestGetProvider(t *testing.T) {
	provider, err := GetProvider("gemini", types.AgentClaude)
	if err != nil {
		t.Fatalf("GetProvider() error = %v", err)
	}
	if provider != geminiProvider {
		t.Fatalf("GetProvider() = %+v, want Claude Gemini provider", provider)
	}
}

func TestGetProviderRejectsUnknownCombination(t *testing.T) {
	provider, err := GetProvider("glm", types.AgentOpencode)
	if err == nil || provider != nil {
		t.Fatalf("GetProvider() = (%+v, %v), want nil provider and error", provider, err)
	}
}

func TestGetProviderNamesReturnsSortedProviderNames(t *testing.T) {
	got := GetProviderNames(types.AgentClaude)
	want := []string{"gemini", "gemini-claude", "glm", "gpt5-codex"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProviderNames() = %v, want %v", got, want)
	}
}

func TestBuildProviderEnvInjectsAuthenticationTokenIntoCopy(t *testing.T) {
	t.Setenv("CLI_PROXY_API_KEY", "secret")

	env, err := BuildProviderEnv(geminiProvider)
	if err != nil {
		t.Fatalf("BuildProviderEnv() error = %v", err)
	}
	if env["ANTHROPIC_AUTH_TOKEN"] != "secret" {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN = %q, want secret", env["ANTHROPIC_AUTH_TOKEN"])
	}
	env["ANTHROPIC_BASE_URL"] = "changed"
	if geminiProvider.Environment["ANTHROPIC_BASE_URL"] == "changed" {
		t.Fatal("BuildProviderEnv() mutated provider environment")
	}
}

func TestBuildProviderEnvRejectsMissingAuthenticationToken(t *testing.T) {
	t.Setenv("ZAI_AUTH_TOKEN", "")

	_, err := BuildProviderEnv(glmProvider)
	if err == nil || !strings.Contains(err.Error(), "ZAI_AUTH_TOKEN") {
		t.Fatalf("BuildProviderEnv() error = %v, want missing-token error", err)
	}
}

func TestBuildProviderEnvAllowsProviderWithoutAuthenticationToken(t *testing.T) {
	env, err := BuildProviderEnv(geminiOpencodeProvider)
	if err != nil {
		t.Fatalf("BuildProviderEnv() error = %v", err)
	}
	if len(env) != 0 {
		t.Fatalf("BuildProviderEnv() = %v, want empty environment", env)
	}
}

func TestGetProviderCliArgs(t *testing.T) {
	got := GetProviderCliArgs(geminiOpencodeProvider)
	want := []string{"--model", "google/gemini-2.5-pro"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProviderCliArgs() = %v, want %v", got, want)
	}
}
