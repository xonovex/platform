package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	governancecontracts "github.com/xonovex/platform/packages/agent/agent-operator-go/internal/governance"
)

const (
	governanceDecisionAPIVersion               = "governance.xonovex.com/v1alpha1"
	governanceEvaluatorVersion                 = "governance-evaluator/1"
	governanceOperationAnnotation              = "governance.xonovex.com/operation"
	governanceCorrelationAnnotation            = "governance.xonovex.com/correlation-id"
	governanceDecisionAnnotation               = "governance.xonovex.com/decision"
	governanceSubjectRevisionAnnotation        = "governance.xonovex.com/subject-revision"
	governancePolicyVersionAnnotation          = "governance.xonovex.com/policy-version"
	governanceEvaluatorVersionAnnotation       = "governance.xonovex.com/evaluator-version"
	governanceOperationDigestAnnotation        = "governance.xonovex.com/operation-digest"
	governanceDecisionEvidenceAnnotation       = "governance.xonovex.com/decision-evidence-reference"
	governanceEnforcementEvidenceAnnotation    = "governance.xonovex.com/enforcement-evidence-reference"
	governanceProtectedTargetsDigestAnnotation = "governance.xonovex.com/protected-targets-digest"
)

type GovernanceDecisionRequest struct {
	APIVersion                string            `json:"apiVersion"`
	CorrelationID             string            `json:"correlationId"`
	Subject                   GovernanceSubject `json:"subject"`
	Policy                    GovernancePolicy  `json:"policy"`
	ProtectedTargetReferences []string          `json:"protectedTargetReferences,omitempty"`
	Operation                 json.RawMessage   `json:"operation"`
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
	APIVersion             string `json:"apiVersion"`
	CorrelationID          string `json:"correlationId"`
	SubjectReference       string `json:"subjectReference"`
	SubjectRevision        string `json:"subjectRevision,omitempty"`
	Decision               string `json:"decision"`
	FailureCode            string `json:"failureCode,omitempty"`
	PolicyVersion          string `json:"policyVersion"`
	EvaluatorVersion       string `json:"evaluatorVersion"`
	OperationDigest        string `json:"operationDigest"`
	ProtectedTargetsDigest string `json:"protectedTargetsDigest"`
	EvidenceReference      string `json:"evidenceReference,omitempty"`
}

type GovernanceEnforcementRequest struct {
	APIVersion                string `json:"apiVersion"`
	CorrelationID             string `json:"correlationId"`
	SubjectReference          string `json:"subjectReference"`
	SubjectRevision           string `json:"subjectRevision,omitempty"`
	Outcome                   string `json:"outcome"`
	FailureCode               string `json:"failureCode,omitempty"`
	PolicyVersion             string `json:"policyVersion"`
	EvaluatorVersion          string `json:"evaluatorVersion"`
	OperationDigest           string `json:"operationDigest"`
	ProtectedTargetsDigest    string `json:"protectedTargetsDigest"`
	DecisionEvidenceReference string `json:"decisionEvidenceReference"`
	EnforcementPoint          string `json:"enforcementPoint"`
}

type GovernanceEnforcementResponse struct {
	APIVersion        string `json:"apiVersion"`
	CorrelationID     string `json:"correlationId"`
	Status            string `json:"status"`
	EvidenceReference string `json:"evidenceReference"`
}

type GovernanceDecisionClient interface {
	Decide(context.Context, GovernanceDecisionRequest) (GovernanceDecisionResponse, error)
	RecordEnforcement(context.Context, GovernanceEnforcementRequest) (GovernanceEnforcementResponse, error)
}

type httpGovernanceDecisionClient struct {
	decisionEndpoint    string
	enforcementEndpoint string
	client              *http.Client
}

func decodeGovernanceResponse(input io.Reader, target any) error {
	const maximumBytes = 1_048_576
	payload, err := io.ReadAll(io.LimitReader(input, maximumBytes+1))
	if err != nil {
		return err
	}
	if len(payload) > maximumBytes {
		return fmt.Errorf("governance response exceeds %d bytes", maximumBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func NewHTTPGovernanceDecisionClient(baseURL string) GovernanceDecisionClient {
	return &httpGovernanceDecisionClient{
		decisionEndpoint:    strings.TrimRight(baseURL, "/") + "/v1/decisions",
		enforcementEndpoint: strings.TrimRight(baseURL, "/") + "/v1/enforcements",
		client:              &http.Client{Timeout: 2 * time.Second},
	}
}

func (c *httpGovernanceDecisionClient) Decide(ctx context.Context, request GovernanceDecisionRequest) (GovernanceDecisionResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return GovernanceDecisionResponse{}, fmt.Errorf("encode governance decision request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.decisionEndpoint, bytes.NewReader(payload))
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
	if err := decodeGovernanceResponse(response.Body, &verdict); err != nil {
		return GovernanceDecisionResponse{}, fmt.Errorf("decode governance decision response: %w", err)
	}
	if err := validateGovernanceVerdict(request, verdict); err != nil {
		return GovernanceDecisionResponse{}, err
	}
	return verdict, nil
}

func (c *httpGovernanceDecisionClient) RecordEnforcement(ctx context.Context, request GovernanceEnforcementRequest) (GovernanceEnforcementResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return GovernanceEnforcementResponse{}, fmt.Errorf("encode governance enforcement evidence: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.enforcementEndpoint, bytes.NewReader(payload))
	if err != nil {
		return GovernanceEnforcementResponse{}, fmt.Errorf("create governance enforcement request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(httpRequest)
	if err != nil {
		return GovernanceEnforcementResponse{}, fmt.Errorf("record governance enforcement evidence: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4_096))
		return GovernanceEnforcementResponse{}, fmt.Errorf("governance enforcement service returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var receipt GovernanceEnforcementResponse
	if err := decodeGovernanceResponse(response.Body, &receipt); err != nil {
		return GovernanceEnforcementResponse{}, fmt.Errorf("decode governance enforcement response: %w", err)
	}
	if receipt.APIVersion != request.APIVersion || receipt.CorrelationID != request.CorrelationID || receipt.Status != "recorded" || receipt.EvidenceReference == "" {
		return GovernanceEnforcementResponse{}, fmt.Errorf("governance enforcement response contract mismatch")
	}
	if !governancecontracts.ContentAddressedEvidenceReference(receipt.EvidenceReference) {
		return GovernanceEnforcementResponse{}, fmt.Errorf("governance enforcement response evidence is not content-addressed")
	}
	return receipt, nil
}

func digestGovernanceOperation(operation json.RawMessage) (string, error) {
	var decoded any
	if err := json.Unmarshal(operation, &decoded); err != nil {
		return "", fmt.Errorf("decode governance operation: %w", err)
	}
	return governancecontracts.DigestJSON(decoded)
}

func validateGovernanceVerdict(request GovernanceDecisionRequest, verdict GovernanceDecisionResponse) error {
	digest, err := digestGovernanceOperation(request.Operation)
	if err != nil {
		return err
	}
	protectedTargets := append([]string(nil), request.ProtectedTargetReferences...)
	slices.Sort(protectedTargets)
	protectedTargetsDigest, err := governancecontracts.DigestJSON(protectedTargets)
	if err != nil {
		return err
	}
	if verdict.APIVersion != request.APIVersion ||
		verdict.CorrelationID != request.CorrelationID ||
		verdict.SubjectReference != request.Subject.Reference ||
		verdict.SubjectRevision != request.Subject.Revision ||
		verdict.PolicyVersion != request.Policy.Version ||
		verdict.EvaluatorVersion != governanceEvaluatorVersion ||
		verdict.OperationDigest != digest ||
		verdict.ProtectedTargetsDigest != protectedTargetsDigest ||
		!governancecontracts.ContentAddressedEvidenceReference(verdict.EvidenceReference) {
		return fmt.Errorf("governance decision response contract mismatch")
	}
	if verdict.Decision != "allow" && verdict.Decision != "deny" && verdict.Decision != "observe" {
		return fmt.Errorf("governance decision response has unknown decision %q", verdict.Decision)
	}
	if verdict.Decision != "allow" && verdict.FailureCode == "" {
		return fmt.Errorf("governance decision response is missing failureCode")
	}
	return nil
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
	subjectRevision, err := governancecontracts.AgentRunSubjectRevision(run)
	if err != nil {
		return failure("governance-subject-revision-unavailable", err)
	}
	subjectReference := fmt.Sprintf("agentrun:%s/%s", run.Namespace, run.Name)
	operationDigest, err := digestGovernanceOperation(json.RawMessage(operation))
	if err != nil {
		return failure("governance-operation-invalid", err)
	}
	correlationDigest, err := governancecontracts.DigestJSON(map[string]string{
		"subjectReference": subjectReference,
		"subjectRevision":  subjectRevision,
		"operationDigest":  operationDigest,
	})
	if err != nil {
		return failure("governance-correlation-id-unavailable", err)
	}
	correlationID := "agentrun:" + strings.TrimPrefix(correlationDigest, "sha256:")
	if run.Annotations == nil {
		run.Annotations = map[string]string{}
	}
	run.Annotations[governanceCorrelationAnnotation] = correlationID
	decisionRequest := GovernanceDecisionRequest{
		APIVersion:    governanceDecisionAPIVersion,
		CorrelationID: correlationID,
		Subject: GovernanceSubject{
			Reference: subjectReference,
			Revision:  subjectRevision,
		},
		Policy:                    GovernancePolicy{Version: policyVersion, Enforcement: enforcement},
		ProtectedTargetReferences: protectedTargetReferences(run),
		Operation:                 json.RawMessage(operation),
	}
	verdict, err := w.DecisionClient.Decide(ctx, decisionRequest)
	if err != nil {
		return failure("governance-decision-unavailable", err)
	}
	if err := validateGovernanceVerdict(decisionRequest, verdict); err != nil {
		return failure("governance-decision-invalid", err)
	}
	enforcementEvidence, err := w.DecisionClient.RecordEnforcement(ctx, GovernanceEnforcementRequest{
		APIVersion:                verdict.APIVersion,
		CorrelationID:             verdict.CorrelationID,
		SubjectReference:          verdict.SubjectReference,
		SubjectRevision:           verdict.SubjectRevision,
		Outcome:                   verdict.Decision,
		FailureCode:               verdict.FailureCode,
		PolicyVersion:             verdict.PolicyVersion,
		EvaluatorVersion:          verdict.EvaluatorVersion,
		OperationDigest:           verdict.OperationDigest,
		ProtectedTargetsDigest:    verdict.ProtectedTargetsDigest,
		DecisionEvidenceReference: verdict.EvidenceReference,
		EnforcementPoint:          "kubernetes:AgentRun:admission",
	})
	if err != nil || enforcementEvidence.EvidenceReference == "" {
		return failure("governance-enforcement-evidence-unavailable", err)
	}
	run.Annotations[governanceDecisionAnnotation] = verdict.Decision
	run.Annotations[governanceSubjectRevisionAnnotation] = verdict.SubjectRevision
	run.Annotations[governancePolicyVersionAnnotation] = verdict.PolicyVersion
	run.Annotations[governanceEvaluatorVersionAnnotation] = verdict.EvaluatorVersion
	run.Annotations[governanceOperationDigestAnnotation] = verdict.OperationDigest
	run.Annotations[governanceDecisionEvidenceAnnotation] = verdict.EvidenceReference
	run.Annotations[governanceEnforcementEvidenceAnnotation] = enforcementEvidence.EvidenceReference
	run.Annotations[governanceProtectedTargetsDigestAnnotation] = verdict.ProtectedTargetsDigest
	if verdict.Decision == "allow" {
		return nil, nil
	}
	code := verdict.FailureCode
	if code == "" {
		code = "governance-decision-without-failure-code"
	}
	return failure(code, nil)
}

func protectedTargetReferences(run *agentv1alpha1.AgentRun) []string {
	if run.Spec.Autonomy == nil {
		return []string{}
	}
	return append([]string(nil), run.Spec.Autonomy.ProtectedTargets...)
}
