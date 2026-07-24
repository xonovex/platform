package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
)

const (
	defaultAddress         = ":8080"
	defaultBodyLimit       = 2 << 20
	defaultReplayWindow    = 5 * time.Minute
	defaultWorkerInterval  = time.Second
	defaultRequestTimeout  = 15 * time.Second
	defaultShutdownTimeout = 20 * time.Second
)

type File struct {
	Address             string            `json:"address"`
	AdminTokenEnv       string            `json:"admin_token_env"`
	DatabaseURLEnv      string            `json:"database_url_env"`
	ApplyEnabled        bool              `json:"apply_enabled"`
	BodyLimitBytes      int64             `json:"body_limit_bytes"`
	ReplayWindowSeconds int64             `json:"replay_window_seconds"`
	WorkerIntervalMS    int64             `json:"worker_interval_ms"`
	RequestTimeoutMS    int64             `json:"request_timeout_ms"`
	ShutdownTimeoutMS   int64             `json:"shutdown_timeout_ms"`
	MaxEffectAttempts   int               `json:"max_effect_attempts"`
	Tenants             map[string]Tenant `json:"tenants"`
}

type Tenant struct {
	Provider           domain.Provider     `json:"provider"`
	BaseURL            string              `json:"base_url"`
	TokenEnv           string              `json:"token_env"`
	WebhookSecretEnv   string              `json:"webhook_secret_env"`
	WebhookSigningEnv  string              `json:"webhook_signing_env"`
	AllowLegacyWebhook bool                `json:"allow_legacy_webhook"`
	AllowInsecureHTTP  bool                `json:"allow_insecure_http"`
	Repositories       []string            `json:"repositories"`
	Effects            []domain.EffectKind `json:"effects"`
	BotLogins          []string            `json:"bot_logins"`
}

type Runtime struct {
	Address           string
	AdminToken        string
	DatabaseURL       string
	ApplyEnabled      bool
	BodyLimitBytes    int64
	ReplayWindow      time.Duration
	WorkerInterval    time.Duration
	RequestTimeout    time.Duration
	ShutdownTimeout   time.Duration
	MaxEffectAttempts int
	Tenants           map[string]RuntimeTenant
}

type RuntimeTenant struct {
	Provider           domain.Provider
	BaseURL            *url.URL
	Token              string
	WebhookSecret      string
	WebhookSigning     string
	AllowLegacyWebhook bool
	Repositories       []string
	Effects            []domain.EffectKind
	BotLogins          []string
}

func Load(path string) (Runtime, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Runtime{}, fmt.Errorf("read dispatcher config: %w", err)
	}
	var file File
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return Runtime{}, fmt.Errorf("decode dispatcher config: %w", err)
	}
	return file.Resolve(os.LookupEnv)
}

func (file File) Resolve(lookup func(string) (string, bool)) (Runtime, error) {
	if file.Address == "" {
		file.Address = defaultAddress
	}
	if file.AdminTokenEnv == "" {
		file.AdminTokenEnv = "DISPATCHER_ADMIN_TOKEN"
	}
	if file.DatabaseURLEnv == "" {
		file.DatabaseURLEnv = "DISPATCHER_DATABASE_URL"
	}
	if file.BodyLimitBytes == 0 {
		file.BodyLimitBytes = defaultBodyLimit
	}
	if file.ReplayWindowSeconds == 0 {
		file.ReplayWindowSeconds = int64(defaultReplayWindow.Seconds())
	}
	if file.WorkerIntervalMS == 0 {
		file.WorkerIntervalMS = defaultWorkerInterval.Milliseconds()
	}
	if file.RequestTimeoutMS == 0 {
		file.RequestTimeoutMS = defaultRequestTimeout.Milliseconds()
	}
	if file.ShutdownTimeoutMS == 0 {
		file.ShutdownTimeoutMS = defaultShutdownTimeout.Milliseconds()
	}
	if file.MaxEffectAttempts == 0 {
		file.MaxEffectAttempts = 8
	}
	if file.BodyLimitBytes < 1024 || file.ReplayWindowSeconds < 1 || file.WorkerIntervalMS < 10 ||
		file.RequestTimeoutMS < 100 || file.ShutdownTimeoutMS < 100 || file.MaxEffectAttempts < 1 {
		return Runtime{}, errors.New("dispatcher limits and durations must be positive and within safe minimums")
	}
	adminToken, err := requiredEnvironment(lookup, file.AdminTokenEnv)
	if err != nil {
		return Runtime{}, err
	}
	databaseURL, err := requiredEnvironment(lookup, file.DatabaseURLEnv)
	if err != nil {
		return Runtime{}, err
	}
	if len(file.Tenants) == 0 {
		return Runtime{}, errors.New("at least one tenant is required")
	}

	runtimeTenants := make(map[string]RuntimeTenant, len(file.Tenants))
	for id, tenant := range file.Tenants {
		resolved, err := tenant.resolve(lookup)
		if err != nil {
			return Runtime{}, fmt.Errorf("tenant %q: %w", id, err)
		}
		runtimeTenants[id] = resolved
	}

	return Runtime{
		Address:           file.Address,
		AdminToken:        adminToken,
		DatabaseURL:       databaseURL,
		ApplyEnabled:      file.ApplyEnabled,
		BodyLimitBytes:    file.BodyLimitBytes,
		ReplayWindow:      time.Duration(file.ReplayWindowSeconds) * time.Second,
		WorkerInterval:    time.Duration(file.WorkerIntervalMS) * time.Millisecond,
		RequestTimeout:    time.Duration(file.RequestTimeoutMS) * time.Millisecond,
		ShutdownTimeout:   time.Duration(file.ShutdownTimeoutMS) * time.Millisecond,
		MaxEffectAttempts: file.MaxEffectAttempts,
		Tenants:           runtimeTenants,
	}, nil
}

func (tenant Tenant) resolve(lookup func(string) (string, bool)) (RuntimeTenant, error) {
	if tenant.Provider != domain.ProviderGitHub && tenant.Provider != domain.ProviderGitLab {
		return RuntimeTenant{}, fmt.Errorf("unsupported provider %q", tenant.Provider)
	}
	parsedURL, err := url.Parse(tenant.BaseURL)
	if err != nil || parsedURL.Host == "" || parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return RuntimeTenant{}, errors.New("base_url must be an absolute provider API URL without credentials, query, or fragment")
	}
	if parsedURL.Scheme != "https" && (!tenant.AllowInsecureHTTP || parsedURL.Scheme != "http") {
		return RuntimeTenant{}, errors.New("base_url must use HTTPS unless allow_insecure_http is explicitly enabled")
	}
	token, err := requiredEnvironment(lookup, tenant.TokenEnv)
	if err != nil {
		return RuntimeTenant{}, fmt.Errorf("provider token: %w", err)
	}
	webhookSecret := optionalEnvironment(lookup, tenant.WebhookSecretEnv)
	webhookSigning := optionalEnvironment(lookup, tenant.WebhookSigningEnv)
	if tenant.Provider == domain.ProviderGitHub && webhookSecret == "" {
		return RuntimeTenant{}, errors.New("GitHub requires webhook_secret_env")
	}
	if tenant.Provider == domain.ProviderGitLab && webhookSigning == "" && (!tenant.AllowLegacyWebhook || webhookSecret == "") {
		return RuntimeTenant{}, errors.New("GitLab requires webhook_signing_env or an explicitly allowed legacy webhook secret")
	}
	if len(tenant.Repositories) == 0 || len(tenant.Effects) == 0 {
		return RuntimeTenant{}, errors.New("repositories and effects allowlists must not be empty")
	}
	for _, effect := range tenant.Effects {
		if !slices.Contains(domain.EffectKinds, effect) {
			return RuntimeTenant{}, fmt.Errorf("unsupported effect allowlist entry %q", effect)
		}
	}
	return RuntimeTenant{
		Provider:           tenant.Provider,
		BaseURL:            parsedURL,
		Token:              token,
		WebhookSecret:      webhookSecret,
		WebhookSigning:     webhookSigning,
		AllowLegacyWebhook: tenant.AllowLegacyWebhook,
		Repositories:       uniqueNonEmpty(tenant.Repositories),
		Effects:            slices.Clone(tenant.Effects),
		BotLogins:          uniqueNonEmpty(tenant.BotLogins),
	}, nil
}

func (tenant RuntimeTenant) Allows(repository string, effect domain.EffectKind) bool {
	return slices.Contains(tenant.Repositories, repository) && slices.Contains(tenant.Effects, effect)
}

func (tenant RuntimeTenant) IsBot(login string) bool {
	return login != "" && slices.Contains(tenant.BotLogins, login)
}

func requiredEnvironment(lookup func(string) (string, bool), name string) (string, error) {
	if name == "" {
		return "", errors.New("environment variable name is required")
	}
	value, ok := lookup(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("environment variable %q is required", name)
	}
	return value, nil
}

func optionalEnvironment(lookup func(string) (string, bool), name string) string {
	if name == "" {
		return ""
	}
	value, _ := lookup(name)
	return value
}

func uniqueNonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result
}
