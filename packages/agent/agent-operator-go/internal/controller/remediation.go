package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	runtel "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/telemetry"
)

// RemediationRouter is the incident-to-trigger output port. Implementations
// must use the AgentTrigger HTTP surface rather than creating runs directly.
type RemediationRouter interface {
	Raise(context.Context, *agentv1alpha1.AgentRun, runtel.DriftAssessment) error
}

type agentTriggerRemediationRouter struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewAgentTriggerRemediationRouter(endpoint, token string) RemediationRouter {
	return &agentTriggerRemediationRouter{
		endpoint: endpoint,
		token:    token,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (router *agentTriggerRemediationRouter) Raise(ctx context.Context, run *agentv1alpha1.AgentRun, assessment runtel.DriftAssessment) error {
	payload, err := json.Marshal(map[string]any{
		"correlationId": runCorrelationID(run),
		"runReference":  fmt.Sprintf("agentrun:%s/%s", run.Namespace, run.Name),
		"failureCodes":  assessment.FailureCodes,
	})
	if err != nil {
		return fmt.Errorf("encode remediation trigger: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, router.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create remediation trigger request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+router.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", runCorrelationID(run)+":drift-remediation")
	response, err := router.client.Do(request)
	if err != nil {
		return fmt.Errorf("post remediation AgentTrigger: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("remediation AgentTrigger returned %s", strings.TrimSpace(response.Status))
	}
	return nil
}
