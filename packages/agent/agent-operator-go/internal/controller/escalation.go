package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// EscalationRouter delivers a bounded request to an accountable recipient.
// The returned provider-native reference is retained as observed evidence.
type EscalationRouter interface {
	Raise(context.Context, *agentv1alpha1.AgentRun, *agentv1alpha1.AgentEscalationStatus) (string, error)
}

type httpEscalationRouter struct {
	endpoint        string
	token           string
	responseBaseURL string
	client          *http.Client
}

type escalationDeliveryResponse struct {
	DeliveryReference string `json:"deliveryReference"`
}

func NewHTTPEscalationRouter(endpoint, token, responseBaseURL string) EscalationRouter {
	return &httpEscalationRouter{
		endpoint:        strings.TrimRight(endpoint, "/"),
		token:           token,
		responseBaseURL: strings.TrimRight(responseBaseURL, "/"),
		client:          &http.Client{Timeout: 5 * time.Second},
	}
}

func (router *httpEscalationRouter) Raise(ctx context.Context, run *agentv1alpha1.AgentRun, escalation *agentv1alpha1.AgentEscalationStatus) (string, error) {
	payload, err := json.Marshal(map[string]any{
		"correlationId":   runCorrelationID(run),
		"runReference":    fmt.Sprintf("agentrun:%s/%s", run.Namespace, run.Name),
		"subjectRevision": escalation.SubjectRevision,
		"recipient":       escalation.Recipient,
		"requestedAt":     escalation.RequestedAt,
		"expiresAt":       escalation.ExpiresAt,
		"safeDefault":     escalation.SafeDefault,
		"responseUrl":     fmt.Sprintf("%s/v1/escalations/%s/%s", router.responseBaseURL, run.Namespace, run.Name),
	})
	if err != nil {
		return "", fmt.Errorf("encode escalation delivery: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, router.endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("create escalation delivery request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+router.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", runCorrelationID(run)+":escalation:"+escalation.SubjectRevision)
	response, err := router.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("deliver escalation: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4_096))
		return "", fmt.Errorf("escalation router returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	var receipt escalationDeliveryResponse
	if err := decodeStrictJSON(response.Body, 1_048_576, &receipt); err != nil {
		return "", fmt.Errorf("decode escalation delivery receipt: %w", err)
	}
	if strings.TrimSpace(receipt.DeliveryReference) == "" {
		return "", fmt.Errorf("escalation delivery receipt has no deliveryReference")
	}
	return receipt.DeliveryReference, nil
}
