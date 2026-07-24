package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
)

func TestResolveAppliesDefaultsAndSecrets(t *testing.T) {
	file := File{
		Tenants: map[string]Tenant{
			"github-main": {
				Provider:         domain.ProviderGitHub,
				BaseURL:          "https://api.github.com",
				TokenEnv:         "GH_TOKEN",
				WebhookSecretEnv: "GH_WEBHOOK",
				Repositories:     []string{"owner/repo", "owner/repo"},
				Effects:          []domain.EffectKind{domain.EffectTicketCreate},
				BotLogins:        []string{"dispatcher[bot]"},
			},
		},
	}
	environment := map[string]string{
		"DISPATCHER_ADMIN_TOKEN":  "admin",
		"DISPATCHER_DATABASE_URL": "postgres://dispatcher",
		"GH_TOKEN":                "token",
		"GH_WEBHOOK":              "secret",
	}
	runtime, err := file.Resolve(func(name string) (string, bool) {
		value, ok := environment[name]
		return value, ok
	})
	if err != nil {
		t.Fatal(err)
	}
	if runtime.Address != ":8080" || runtime.BodyLimitBytes == 0 || runtime.MaxEffectAttempts != 8 {
		t.Fatalf("defaults not applied: %+v", runtime)
	}
	tenant := runtime.Tenants["github-main"]
	if !tenant.Allows("owner/repo", domain.EffectTicketCreate) || !tenant.IsBot("dispatcher[bot]") {
		t.Fatal("allowlists were not resolved")
	}
	if len(tenant.Repositories) != 1 {
		t.Fatalf("repositories were not deduplicated: %v", tenant.Repositories)
	}
}

func TestLoadReadsStrictJSON(t *testing.T) {
	t.Setenv("ADMIN", "admin")
	t.Setenv("DB", "postgres://dispatcher")
	t.Setenv("TOKEN", "token")
	t.Setenv("SECRET", "secret")
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
		"admin_token_env":"ADMIN",
		"database_url_env":"DB",
		"tenants":{
			"github":{
				"provider":"github",
				"base_url":"https://api.github.com",
				"token_env":"TOKEN",
				"webhook_secret_env":"SECRET",
				"repositories":["owner/repo"],
				"effects":["ticket.create"]
			}
		}
	}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	runtime, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.Tenants["github"].Token != "token" {
		t.Fatal("configuration secrets were not resolved")
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("expected missing configuration failure")
	}
}

func TestResolveGitLabSigningAndLegacyModes(t *testing.T) {
	base := File{
		AdminTokenEnv:  "ADMIN",
		DatabaseURLEnv: "DB",
		Tenants: map[string]Tenant{
			"gitlab": {
				Provider:          domain.ProviderGitLab,
				BaseURL:           "https://gitlab.example.com",
				TokenEnv:          "TOKEN",
				WebhookSigningEnv: "SIGNING",
				Repositories:      []string{"group/project"},
				Effects:           []domain.EffectKind{domain.EffectContextPublish},
			},
		},
	}
	values := map[string]string{"ADMIN": "a", "DB": "d", "TOKEN": "t", "SIGNING": "whsec_abc"}
	if _, err := base.Resolve(func(name string) (string, bool) { value, ok := values[name]; return value, ok }); err != nil {
		t.Fatal(err)
	}
	delete(values, "SIGNING")
	if _, err := base.Resolve(func(name string) (string, bool) { value, ok := values[name]; return value, ok }); err == nil {
		t.Fatal("expected missing signing token failure")
	}
	tenant := base.Tenants["gitlab"]
	tenant.AllowLegacyWebhook = true
	tenant.WebhookSecretEnv = "SECRET"
	base.Tenants["gitlab"] = tenant
	values["SECRET"] = "legacy"
	if _, err := base.Resolve(func(name string) (string, bool) { value, ok := values[name]; return value, ok }); err != nil {
		t.Fatal(err)
	}
}

func TestResolveRejectsUnsafeConfiguration(t *testing.T) {
	tests := []File{
		{},
		{Tenants: map[string]Tenant{}},
		{Tenants: map[string]Tenant{"bad": {Provider: "other"}}},
		{Tenants: map[string]Tenant{"bad": {
			Provider: domain.ProviderGitHub, BaseURL: "http://example.com", TokenEnv: "TOKEN",
			WebhookSecretEnv: "SECRET", Repositories: []string{"a/b"}, Effects: []domain.EffectKind{domain.EffectTicketCreate},
		}}},
	}
	lookup := func(name string) (string, bool) {
		return map[string]string{
			"DISPATCHER_ADMIN_TOKEN": "admin", "DISPATCHER_DATABASE_URL": "db",
			"TOKEN": "token", "SECRET": "secret",
		}[name], true
	}
	for index, file := range tests {
		if _, err := file.Resolve(lookup); err == nil {
			t.Errorf("case %d: expected configuration error", index)
		}
	}
}
