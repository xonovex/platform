package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

const (
	governanceDecisionAPIVersion    = "governance.xonovex.com/v1alpha1"
	governanceOperationAnnotation   = "governance.xonovex.com/operation"
	governanceCorrelationAnnotation = "governance.xonovex.com/correlation-id"
)

type GovernanceDecisionRequest struct {
	APIVersion    string            `json:"apiVersion"`
	CorrelationID string            `json:"correlationId"`
	Subject       GovernanceSubject `json:"subject"`
	Policy        GovernancePolicy  `json:"policy"`
	Operation     json.RawMessage   `json:"operation"`
}

type GovernanceSubject struct {
	Reference string `json:"reference"`
	Revision  string `json:"revision,omitempty"`
}

type GovernancePolicy struct {
	Version     string `json:"version"`
	Enforcement string `json:"enforcement"`
}

type GovernanceDecisionResponse struct {
	APIVersion        string `json:"apiVersion"`
	CorrelationID     string `json:"correlationId"`
	SubjectReference  string `json:"subjectReference"`
	Decision          string `json:"decision"`
	FailureCode       string `json:"failureCode,omitempty"`
	PolicyVersion     string `json:"policyVersion"`
	EvidenceReference string `json:"evidenceReference,omitempty"`
}

type GovernanceDecisionClient interface {
	Decide(context.Context, GovernanceDecisionRequest) (GovernanceDecisionResponse, error)
}

type httpGovernanceDecisionClient struct {
	endpoint string
	client   *http.Client
}

func NewHTTPGovernanceDecisionClient(baseURL string) GovernanceDecisionClient {
	return &httpGovernanceDecisionClient{
		endpoint: strings.TrimRight(baseURL, "/") + "/v1/decisions",
		client:   &http.Client{Timeout: 2 * time.Second},
	}
}

func (c *httpGovernanceDecisionClient) Decide(ctx context.Context, request GovernanceDecisionRequest) (GovernanceDecisionResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return GovernanceDecisionResponse{}, fmt.Errorf("encode governance decision request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return GovernanceDecisionResponse{}, fmt.Errorf("create governance decision request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return GovernanceDecisionResponse{}, fmt.Errorf("call governance decision service: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4_096))
		return GovernanceDecisionResponse{}, fmt.Errorf("governance decision service returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var verdict GovernanceDecisionResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1_048_576)).Decode(&verdict); err != nil {
		return GovernanceDecisionResponse{}, fmt.Errorf("decode governance decision response: %w", err)
	}
	if verdict.APIVersion != governanceDecisionAPIVersion || verdict.CorrelationID != request.CorrelationID {
		return GovernanceDecisionResponse{}, fmt.Errorf("governance decision response contract mismatch")
	}
	if verdict.Decision != "allow" && verdict.Decision != "deny" && verdict.Decision != "observe" {
		return GovernanceDecisionResponse{}, fmt.Errorf("governance decision response has unknown decision %q", verdict.Decision)
	}
	return verdict, nil
}

func (w *AgentRunWebhook) enforceGovernance(ctx context.Context, run *agentv1alpha1.AgentRun, policy *agentv1alpha1.AgentPolicy) (admission.Warnings, error) {
	required := policy.Spec.Enforced.RequireGovernanceVerdict
	operation, governed := run.Annotations[governanceOperationAnnotation]
	if !required && !governed {
		return nil, nil
	}

	failure := func(code string, cause error) (admission.Warnings, error) {
		message := code
		if cause != nil {
			message = fmt.Sprintf("%s: %v", code, cause)
		}
		if required {
			return nil, fmt.Errorf("governance decision denied: %s", message)
		}
		return admission.Warnings{fmt.Sprintf("governance decision observed: %s", message)}, nil
	}

	if !governed || !json.Valid([]byte(operation)) {
		return failure("governance-operation-invalid", nil)
	}
	correlationID := run.Annotations[governanceCorrelationAnnotation]
	if correlationID == "" {
		return failure("governance-correlation-id-required", nil)
	}
	policyVersion := policy.Spec.Enforced.GovernancePolicyVersion
	if policyVersion == "" {
		return failure("governance-policy-version-required", nil)
	}
	if w.DecisionClient == nil {
		return failure("governance-decision-unavailable", nil)
	}

	enforcement := "advisory"
	if required {
		enforcement = "mandatory"
	}
	verdict, err := w.DecisionClient.Decide(ctx, GovernanceDecisionRequest{
		APIVersion:    governanceDecisionAPIVersion,
		CorrelationID: correlationID,
		Subject: GovernanceSubject{
			Reference: fmt.Sprintf("agentrun:%s/%s", run.Namespace, run.Name),
			Revision:  fmt.Sprintf("generation:%d", run.Generation),
		},
		Policy:    GovernancePolicy{Version: policyVersion, Enforcement: enforcement},
		Operation: json.RawMessage(operation),
	})
	if err != nil {
		return failure("governance-decision-unavailable", err)
	}
	if verdict.Decision == "allow" {
		return nil, nil
	}
	code := verdict.FailureCode
	if code == "" {
		code = "governance-decision-without-failure-code"
	}
	return failure(code, nil)
}
