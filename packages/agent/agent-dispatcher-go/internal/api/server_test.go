package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/config"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/metrics"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store"
)

func TestGitHubWebhookAcceptsDeduplicatesAndSuppresses(t *testing.T) {
	database := &fakeStore{}
	runtime := testRuntime()
	server := New(runtime, database, noAdapter, &metrics.Registry{}, discardLogger())
	body := []byte(`{
		"action":"opened",
		"repository":{"full_name":"owner/repo"},
		"sender":{"login":"dispatcher[bot]"},
		"issue":{"id":7,"number":4}
	}`)
	request := httptest.NewRequest(http.MethodPost, "/webhooks/github/github", bytes.NewReader(body))
	request.Header.Set("X-GitHub-Delivery", "delivery-1")
	request.Header.Set("X-GitHub-Event", "issues")
	request.Header.Set("X-Hub-Signature-256", githubSignature("github-secret", body))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || database.ingested.Provider != domain.ProviderGitHub {
		t.Fatalf("unexpected response=%d delivery=%+v body=%s", response.Code, database.ingested, response.Body.String())
	}
	if database.ingested.State != domain.DeliveryIgnored || !database.event.Suppressed {
		t.Fatalf("bot event was not suppressed: %+v %+v", database.ingested, database.event)
	}

	database.inserted = false
	request = httptest.NewRequest(http.MethodPost, "/webhooks/github/github", bytes.NewReader(body))
	request.Header.Set("X-GitHub-Delivery", "delivery-1")
	request.Header.Set("X-GitHub-Event", "issues")
	request.Header.Set("X-Hub-Signature-256", githubSignature("github-secret", body))
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || !strings.Contains(response.Body.String(), `"duplicate":true`) {
		t.Fatalf("duplicate response: %d %s", response.Code, response.Body.String())
	}
}

func TestWebhookRejectsInvalidRequests(t *testing.T) {
	server := New(testRuntime(), &fakeStore{}, noAdapter, &metrics.Registry{}, discardLogger())
	tests := []*http.Request{
		httptest.NewRequest(http.MethodPost, "/webhooks/github/missing", strings.NewReader(`{}`)),
		httptest.NewRequest(http.MethodPost, "/webhooks/github/github", strings.NewReader(`{}`)),
		httptest.NewRequest(http.MethodPost, "/webhooks/gitlab/gitlab", strings.NewReader(`{}`)),
	}
	for index, request := range tests {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		if response.Code < 400 {
			t.Errorf("case %d: expected rejection, got %d", index, response.Code)
		}
	}
}

func TestGitLabStandardWebhook(t *testing.T) {
	runtime := testRuntime()
	now := time.Unix(1_800_000_000, 0)
	database := &fakeStore{}
	server := New(runtime, database, noAdapter, &metrics.Registry{}, discardLogger())
	server.now = func() time.Time { return now }
	body := []byte(`{
		"object_kind":"issue",
		"project":{"path_with_namespace":"group/project"},
		"user":{"username":"author"},
		"object_attributes":{"id":5,"iid":3,"action":"update"}
	}`)
	messageID := "gitlab-delivery"
	timestamp := "1800000000"
	request := httptest.NewRequest(http.MethodPost, "/webhooks/gitlab/gitlab", bytes.NewReader(body))
	request.Header.Set("webhook-id", messageID)
	request.Header.Set("webhook-timestamp", timestamp)
	request.Header.Set("webhook-signature", gitlabSignature([]byte("gitlab-signing"), messageID, timestamp, body))
	request.Header.Set("X-Gitlab-Event", "Issue Hook")
	request.Header.Set("X-Gitlab-Event-UUID", "event-uuid")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || database.ingested.DeliveryID != messageID ||
		database.event.Repository != "group/project" {
		t.Fatalf("unexpected GitLab ingress: %d %+v %+v %s", response.Code, database.ingested, database.event, response.Body.String())
	}
}

func TestEffectAPIEnqueuesReadsAndRequeues(t *testing.T) {
	runtime := testRuntime()
	database := &fakeStore{}
	adapter := &fakeAdapter{}
	server := New(runtime, database, func(_ domain.Provider, _ string) (provider.Adapter, bool) {
		return adapter, true
	}, &metrics.Registry{}, discardLogger())
	effect := domain.Effect{
		IdempotencyKey: "key-1",
		Provider:       domain.ProviderGitHub,
		Tenant:         "github",
		Kind:           domain.EffectTicketCreate,
		Mode:           domain.EffectModePreview,
		Target:         domain.EffectTarget{Repository: "owner/repo"},
		Payload:        domain.EffectPayload{Title: "Ticket"},
	}
	body, _ := json.Marshal(effect)
	request := httptest.NewRequest(http.MethodPost, "/v1/effects", bytes.NewReader(body))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected authentication failure, got %d", response.Code)
	}

	request = authorizedRequest(http.MethodPost, "/v1/effects", body)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || database.enqueued.ID == "" ||
		database.enqueued.CorrelationID == "" {
		t.Fatalf("unexpected enqueue response=%d effect=%+v body=%s", response.Code, database.enqueued, response.Body.String())
	}
	database.effect = database.enqueued
	database.result = domain.EffectResult{StatusCode: 200}
	request = authorizedRequest(http.MethodGet, "/v1/effects/"+database.effect.ID, nil)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), database.effect.ID) {
		t.Fatalf("unexpected effect lookup: %d %s", response.Code, response.Body.String())
	}
	request = authorizedRequest(http.MethodPost, "/v1/dead-letters/"+database.effect.ID+"/requeue", nil)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || database.requeued != database.effect.ID {
		t.Fatalf("unexpected requeue: %d %s", response.Code, response.Body.String())
	}
}

func TestEffectAPIEnforcesApplyAndAllowlists(t *testing.T) {
	runtime := testRuntime()
	server := New(runtime, &fakeStore{}, func(_ domain.Provider, _ string) (provider.Adapter, bool) {
		return &fakeAdapter{}, true
	}, &metrics.Registry{}, discardLogger())
	base := domain.Effect{
		ID: "effect", CorrelationID: "correlation", IdempotencyKey: "key",
		Provider: domain.ProviderGitHub, Tenant: "github", Kind: domain.EffectTicketCreate,
		Mode: domain.EffectModeApply, Target: domain.EffectTarget{Repository: "owner/repo"},
		Payload: domain.EffectPayload{Title: "Ticket"},
	}
	tests := []struct {
		name   string
		change func(*domain.Effect)
		status int
	}{
		{"apply disabled", func(*domain.Effect) {}, http.StatusConflict},
		{"repository", func(effect *domain.Effect) {
			effect.Mode = domain.EffectModePreview
			effect.Target.Repository = "other/repo"
		}, http.StatusForbidden},
		{"tenant", func(effect *domain.Effect) {
			effect.Mode = domain.EffectModePreview
			effect.Tenant = "missing"
		}, http.StatusForbidden},
		{"invalid", func(effect *domain.Effect) {
			effect.Mode = domain.EffectModePreview
			effect.Payload.Title = ""
		}, http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			effect := base
			test.change(&effect)
			body, _ := json.Marshal(effect)
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, authorizedRequest(http.MethodPost, "/v1/effects", body))
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestHealthAndMetrics(t *testing.T) {
	database := &fakeStore{counts: store.Counts{QueuedEffects: 2}}
	registry := &metrics.Registry{}
	registry.EffectsQueued.Add(3)
	server := New(testRuntime(), database, noAdapter, registry, discardLogger())
	for _, path := range []string{"/health/live", "/health/ready", "/metrics"} {
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(response.Body.String(), "xonovex_dispatcher_effects_queued_total 3") {
		t.Fatalf("unexpected metrics: %s", response.Body.String())
	}
}

type fakeStore struct {
	inserted    bool
	ingestCalls int
	ingested    domain.Delivery
	event       domain.WorkflowEvent
	enqueued    domain.Effect
	effect      domain.Effect
	result      domain.EffectResult
	requeued    string
	counts      store.Counts
	err         error
}

func (database *fakeStore) Close() {}
func (database *fakeStore) Ping(context.Context) error {
	return database.err
}
func (database *fakeStore) Migrate(context.Context) error {
	return database.err
}
func (database *fakeStore) Ingest(_ context.Context, delivery domain.Delivery, event domain.WorkflowEvent) (bool, error) {
	database.ingested = delivery
	database.event = event
	database.ingestCalls++
	if database.err != nil {
		return false, database.err
	}
	if delivery.DeliveryID == "delivery-1" {
		return database.ingestCalls == 1, nil
	}
	if !database.inserted {
		return true, nil
	}
	return database.inserted, nil
}
func (database *fakeStore) EnqueueEffect(_ context.Context, effect domain.Effect) (domain.Effect, bool, error) {
	database.enqueued = effect
	return effect, true, database.err
}
func (database *fakeStore) GetEffect(context.Context, string) (domain.Effect, domain.EffectResult, error) {
	if database.effect.ID == "" {
		return domain.Effect{}, domain.EffectResult{}, store.ErrNotFound
	}
	return database.effect, database.result, database.err
}
func (database *fakeStore) ClaimEffect(context.Context, string, time.Duration) (domain.Effect, error) {
	return database.effect, database.err
}
func (database *fakeStore) RenewEffectLease(context.Context, domain.Effect, string, time.Duration) error {
	return database.err
}
func (database *fakeStore) ValidateContextSequence(context.Context, domain.Effect) error {
	return database.err
}
func (database *fakeStore) CompleteEffect(context.Context, domain.Effect, domain.EffectResult) error {
	return database.err
}
func (database *fakeStore) RetryEffect(context.Context, domain.Effect, string, time.Time) error {
	return database.err
}
func (database *fakeStore) DeadLetterEffect(context.Context, domain.Effect, string) error {
	return database.err
}
func (database *fakeStore) ReleaseExpiredLeases(context.Context) (int64, error) {
	return 0, database.err
}
func (database *fakeStore) RequeueDeadLetter(_ context.Context, id string) error {
	database.requeued = id
	return database.err
}
func (database *fakeStore) Counts(context.Context) (store.Counts, error) {
	return database.counts, database.err
}

type fakeAdapter struct{}

func (*fakeAdapter) Preview(domain.Effect) (domain.RequestPreview, error) {
	return domain.RequestPreview{}, nil
}
func (*fakeAdapter) Apply(context.Context, domain.Effect) (domain.EffectResult, error) {
	return domain.EffectResult{}, nil
}

func testRuntime() config.Runtime {
	githubURL, _ := url.Parse("https://api.github.com")
	gitlabURL, _ := url.Parse("https://gitlab.com")
	return config.Runtime{
		AdminToken:     "admin-token",
		ApplyEnabled:   false,
		BodyLimitBytes: 1 << 20,
		ReplayWindow:   5 * time.Minute,
		Tenants: map[string]config.RuntimeTenant{
			"github": {
				Provider: domain.ProviderGitHub, BaseURL: githubURL, Token: "token",
				WebhookSecret: "github-secret", Repositories: []string{"owner/repo"},
				Effects: []domain.EffectKind{domain.EffectTicketCreate}, BotLogins: []string{"dispatcher[bot]"},
			},
			"gitlab": {
				Provider: domain.ProviderGitLab, BaseURL: gitlabURL, Token: "token",
				WebhookSigning: "whsec_" + base64.StdEncoding.EncodeToString([]byte("gitlab-signing")),
				Repositories:   []string{"group/project"}, Effects: []domain.EffectKind{domain.EffectTicketCreate},
			},
		},
	}
}

func noAdapter(domain.Provider, string) (provider.Adapter, bool) {
	return nil, false
}

func githubSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func gitlabSignature(key []byte, messageID string, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(messageID + "." + timestamp + "." + string(body)))
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func authorizedRequest(method string, path string, body []byte) *http.Request {
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer admin-token")
	return request
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

var _ store.Store = (*fakeStore)(nil)
var _ provider.Adapter = (*fakeAdapter)(nil)
