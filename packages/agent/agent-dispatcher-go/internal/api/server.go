package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/config"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/domain"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/metrics"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/provider"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/store"
	"github.com/xonovex/platform/packages/agent/agent-dispatcher-go/internal/webhook"
)

type AdapterLookup func(domain.Provider, string) (provider.Adapter, bool)

type Server struct {
	config  config.Runtime
	store   store.Store
	adapter AdapterLookup
	metrics *metrics.Registry
	logger  *slog.Logger
	now     func() time.Time
}

func New(
	runtimeConfig config.Runtime,
	database store.Store,
	adapter AdapterLookup,
	registry *metrics.Registry,
	logger *slog.Logger,
) *Server {
	return &Server{
		config:  runtimeConfig,
		store:   database,
		adapter: adapter,
		metrics: registry,
		logger:  logger,
		now:     time.Now,
	}
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhooks/github/{tenant}", server.githubWebhook)
	mux.HandleFunc("POST /webhooks/gitlab/{tenant}", server.gitlabWebhook)
	mux.HandleFunc("POST /v1/effects", server.authorize(server.enqueueEffect))
	mux.HandleFunc("GET /v1/effects/{id}", server.authorize(server.getEffect))
	mux.HandleFunc("POST /v1/dead-letters/{id}/requeue", server.authorize(server.requeueDeadLetter))
	mux.HandleFunc("GET /health/live", server.live)
	mux.HandleFunc("GET /health/ready", server.ready)
	mux.HandleFunc("GET /metrics", server.prometheus)
	return server.securityHeaders(server.recoverPanic(mux))
}

func (server *Server) githubWebhook(writer http.ResponseWriter, request *http.Request) {
	tenantID := request.PathValue("tenant")
	tenant, ok := server.config.Tenants[tenantID]
	if !ok || tenant.Provider != domain.ProviderGitHub {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusNotFound, "unknown webhook tenant")
		return
	}
	body, ok := server.readBody(writer, request)
	if !ok {
		return
	}
	if err := webhook.VerifyGitHub(tenant.WebhookSecret, body, request.Header.Get("X-Hub-Signature-256")); err != nil {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusUnauthorized, "invalid webhook signature")
		return
	}
	deliveryID := request.Header.Get("X-GitHub-Delivery")
	eventName := request.Header.Get("X-GitHub-Event")
	if deliveryID == "" || eventName == "" {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusBadRequest, "delivery and event headers are required")
		return
	}
	event, err := webhook.NormalizeGitHub(tenantID, deliveryID, eventName, body, tenant.IsBot)
	if err != nil {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusBadRequest, err.Error())
		return
	}
	server.ingest(request.Context(), writer, domain.Delivery{
		Provider:      domain.ProviderGitHub,
		Tenant:        tenantID,
		DeliveryID:    deliveryID,
		Event:         eventName,
		Action:        event.Action,
		PayloadDigest: domain.Digest(body),
		Payload:       body,
		Headers: map[string]string{
			"x-github-delivery": deliveryID,
			"x-github-event":    eventName,
			"x-github-hook-id":  request.Header.Get("X-GitHub-Hook-ID"),
		},
		State:      deliveryState(event, tenant),
		ReceivedAt: server.now().UTC(),
	}, event)
}

func (server *Server) gitlabWebhook(writer http.ResponseWriter, request *http.Request) {
	tenantID := request.PathValue("tenant")
	tenant, ok := server.config.Tenants[tenantID]
	if !ok || tenant.Provider != domain.ProviderGitLab {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusNotFound, "unknown webhook tenant")
		return
	}
	body, ok := server.readBody(writer, request)
	if !ok {
		return
	}
	signature := request.Header.Get("webhook-signature")
	var verificationError error
	if signature != "" {
		verificationError = webhook.VerifyGitLabStandard(
			tenant.WebhookSigning,
			request.Header.Get("webhook-id"),
			request.Header.Get("webhook-timestamp"),
			body,
			signature,
			server.now(),
			server.config.ReplayWindow,
		)
	} else if tenant.AllowLegacyWebhook {
		verificationError = webhook.VerifyGitLabLegacy(
			tenant.WebhookSecret,
			request.Header.Get("X-Gitlab-Token"),
		)
	} else {
		verificationError = webhook.ErrMissingSignature
	}
	if verificationError != nil {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusUnauthorized, "invalid webhook signature")
		return
	}
	eventName := request.Header.Get("X-Gitlab-Event")
	eventUUID := request.Header.Get("X-Gitlab-Event-UUID")
	deliveryID := firstNonEmpty(
		request.Header.Get("webhook-id"),
		request.Header.Get("Idempotency-Key"),
		eventUUID+":"+domain.Digest(body),
	)
	if eventName == "" {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusBadRequest, "event header is required")
		return
	}
	event, err := webhook.NormalizeGitLab(tenantID, deliveryID, eventUUID, eventName, body, tenant.IsBot)
	if err != nil {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusBadRequest, err.Error())
		return
	}
	server.ingest(request.Context(), writer, domain.Delivery{
		Provider:      domain.ProviderGitLab,
		Tenant:        tenantID,
		DeliveryID:    deliveryID,
		EventUUID:     eventUUID,
		Event:         eventName,
		Action:        event.Action,
		PayloadDigest: domain.Digest(body),
		Payload:       body,
		Headers: map[string]string{
			"webhook-id":            request.Header.Get("webhook-id"),
			"webhook-timestamp":     request.Header.Get("webhook-timestamp"),
			"x-gitlab-event":        eventName,
			"x-gitlab-event-uuid":   eventUUID,
			"x-gitlab-webhook-uuid": request.Header.Get("X-Gitlab-Webhook-UUID"),
		},
		State:      deliveryState(event, tenant),
		ReceivedAt: server.now().UTC(),
	}, event)
}

func (server *Server) ingest(
	ctx context.Context,
	writer http.ResponseWriter,
	delivery domain.Delivery,
	event domain.WorkflowEvent,
) {
	tenant := server.config.Tenants[delivery.Tenant]
	if !contains(tenant.Repositories, event.Repository) {
		event.Suppressed = true
		delivery.State = domain.DeliveryIgnored
	}
	inserted, err := server.store.Ingest(ctx, delivery, event)
	if err != nil {
		server.logger.Error("persist webhook delivery", "error", err)
		writeProblem(writer, http.StatusServiceUnavailable, "delivery persistence failed")
		return
	}
	if inserted {
		server.metrics.DeliveriesAccepted.Add(1)
	} else {
		server.metrics.DeliveriesDuplicate.Add(1)
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{
		"accepted":   true,
		"duplicate":  !inserted,
		"suppressed": event.Suppressed,
	})
}

func (server *Server) enqueueEffect(writer http.ResponseWriter, request *http.Request) {
	var effect domain.Effect
	if !decodeJSON(writer, request, server.config.BodyLimitBytes, &effect) {
		return
	}
	if effect.ID == "" {
		id, err := domain.NewID("eff")
		if err != nil {
			writeProblem(writer, http.StatusInternalServerError, "effect id generation failed")
			return
		}
		effect.ID = id
	}
	if effect.CorrelationID == "" {
		effect.CorrelationID = effect.ID
	}
	effect.State = domain.EffectQueued
	effect.Attempts = 0
	effect.CreatedAt = server.now().UTC()
	effect.NextAttemptAt = effect.CreatedAt
	if err := effect.Validate(); err != nil {
		writeProblem(writer, http.StatusBadRequest, err.Error())
		return
	}
	tenant, ok := server.config.Tenants[effect.Tenant]
	if !ok || tenant.Provider != effect.Provider {
		writeProblem(writer, http.StatusForbidden, "effect tenant and provider are not authorized")
		return
	}
	if !tenant.Allows(effect.Target.Repository, effect.Kind) {
		writeProblem(writer, http.StatusForbidden, "effect kind or repository is not allowlisted")
		return
	}
	if effect.Mode == domain.EffectModeApply && !server.config.ApplyEnabled {
		writeProblem(writer, http.StatusConflict, "provider apply mode is disabled")
		return
	}
	if _, ok := server.adapter(effect.Provider, effect.Tenant); !ok {
		writeProblem(writer, http.StatusServiceUnavailable, "provider adapter is unavailable")
		return
	}
	stored, inserted, err := server.store.EnqueueEffect(request.Context(), effect)
	if errors.Is(err, store.ErrIdempotencyConflict) {
		writeProblem(writer, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		server.logger.Error("enqueue effect", "error", err)
		writeProblem(writer, http.StatusServiceUnavailable, "effect persistence failed")
		return
	}
	if inserted {
		server.metrics.EffectsQueued.Add(1)
	}
	writeJSON(writer, http.StatusAccepted, map[string]any{
		"effect":    stored,
		"duplicate": !inserted,
	})
}

func (server *Server) getEffect(writer http.ResponseWriter, request *http.Request) {
	effect, result, err := server.store.GetEffect(request.Context(), request.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeProblem(writer, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeProblem(writer, http.StatusServiceUnavailable, "effect lookup failed")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"effect": effect, "result": result})
}

func (server *Server) requeueDeadLetter(writer http.ResponseWriter, request *http.Request) {
	if err := server.store.RequeueDeadLetter(request.Context(), request.PathValue("id")); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeProblem(writer, http.StatusNotFound, err.Error())
			return
		}
		writeProblem(writer, http.StatusServiceUnavailable, "dead-letter requeue failed")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]bool{"requeued": true})
}

func (server *Server) live(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "live"})
}

func (server *Server) ready(writer http.ResponseWriter, request *http.Request) {
	if err := server.store.Ping(request.Context()); err != nil {
		writeProblem(writer, http.StatusServiceUnavailable, "database is unavailable")
		return
	}
	counts, err := server.store.Counts(request.Context())
	if err != nil {
		writeProblem(writer, http.StatusServiceUnavailable, "dispatcher state is unavailable")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ready", "counts": counts})
}

func (server *Server) prometheus(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if err := server.metrics.WritePrometheus(writer); err != nil {
		server.logger.Error("write metrics", "error", err)
	}
}

func (server *Server) authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authorization := request.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			writeProblem(writer, http.StatusUnauthorized, "authentication required")
			return
		}
		provided := strings.TrimPrefix(authorization, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(server.config.AdminToken)) != 1 {
			writeProblem(writer, http.StatusUnauthorized, "authentication required")
			return
		}
		next(writer, request)
	}
}

func (server *Server) readBody(writer http.ResponseWriter, request *http.Request) ([]byte, bool) {
	request.Body = http.MaxBytesReader(writer, request.Body, server.config.BodyLimitBytes)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		server.metrics.DeliveriesRejected.Add(1)
		writeProblem(writer, http.StatusBadRequest, "request body is invalid or too large")
		return nil, false
	}
	return body, true
}

func (server *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store")
		writer.Header().Set("Content-Security-Policy", "default-src 'none'")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(writer, request)
	})
}

func (server *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				server.logger.Error("request panic", "recovered", recovered)
				writeProblem(writer, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, limit int64, target any) bool {
	request.Body = http.MaxBytesReader(writer, request.Body, limit)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeProblem(writer, http.StatusBadRequest, "invalid JSON request: "+err.Error())
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeProblem(writer, http.StatusBadRequest, "request body must contain one JSON value")
		return false
	}
	return true
}

func deliveryState(event domain.WorkflowEvent, tenant config.RuntimeTenant) domain.DeliveryState {
	if event.Suppressed || !contains(tenant.Repositories, event.Repository) {
		return domain.DeliveryIgnored
	}
	return domain.DeliveryNormalized
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		slog.Error("encode HTTP response", "error", err)
	}
}

func writeProblem(writer http.ResponseWriter, status int, detail string) {
	writeJSON(writer, status, map[string]any{
		"type":   "about:blank",
		"title":  http.StatusText(status),
		"status": status,
		"detail": detail,
	})
}
