package webhook

import (
	"context"
	"errors"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/agent/agent-operator-go/test/testutil"
)

type stubGovernanceDecisionClient struct {
	response GovernanceDecisionResponse
	err      error
	request  GovernanceDecisionRequest
}

func (s *stubGovernanceDecisionClient) Decide(_ context.Context, request GovernanceDecisionRequest) (GovernanceDecisionResponse, error) {
	s.request = request
	return s.response, s.err
}

func governedPolicy(required bool) *agentv1alpha1.AgentPolicy {
	return &agentv1alpha1.AgentPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "governance", Namespace: "test-ns"},
		Spec: agentv1alpha1.AgentPolicySpec{Enforced: agentv1alpha1.AgentPolicyEnforced{
			RequireGovernanceVerdict: required,
			GovernancePolicyVersion:  "governance-policy/1",
			AllowedRuntimeClassNames: []string{"kata"},
		}},
	}
}

func governedRun(operation string) *agentv1alpha1.AgentRun {
	run := baseRun()
	run.Name = "run"
	run.Annotations = map[string]string{
		governanceOperationAnnotation:   operation,
		governanceCorrelationAnnotation: "correlation-1",
	}
	return run
}

func TestAgentRunWebhook_GovernanceVerdicts(t *testing.T) {
	tests := []struct {
		name        string
		operation   string
		decision    string
		failureCode string
		wantErr     bool
	}{
		{name: "independent release", operation: `{"kind":"independence","input":{"decider":"reviewer","author":"author"}}`, decision: "allow"},
		{name: "self-approved release", operation: `{"kind":"independence","input":{"decider":"author","author":"author"}}`, decision: "deny", failureCode: "release-independence-failed", wantErr: true},
		{name: "in-scope emergency access", operation: `{"kind":"emergency-access","input":{"scope":"production"}}`, decision: "allow"},
		{name: "expired emergency access", operation: `{"kind":"emergency-access","input":{"scope":"production"}}`, decision: "deny", failureCode: "emergency-access-expired", wantErr: true},
		{name: "mechanical executor", operation: `{"kind":"development","input":{"workShape":"mechanical"}}`, decision: "allow"},
		{name: "unbounded adaptive executor", operation: `{"kind":"development","input":{"workShape":"adaptive"}}`, decision: "deny", failureCode: "unbounded-agent", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := governedPolicy(true)
			decisionClient := &stubGovernanceDecisionClient{response: GovernanceDecisionResponse{
				APIVersion:    governanceDecisionAPIVersion,
				CorrelationID: "correlation-1",
				Decision:      test.decision,
				FailureCode:   test.failureCode,
			}}
			webhook := &AgentRunWebhook{
				Client:         fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(policy).Build(),
				DecisionClient: decisionClient,
			}

			_, err := webhook.ValidateCreate(context.Background(), governedRun(test.operation))

			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateCreate() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.failureCode != "" && !strings.Contains(err.Error(), test.failureCode) {
				t.Fatalf("ValidateCreate() error = %v, want failure code %q", err, test.failureCode)
			}
			if decisionClient.request.Policy.Enforcement != "mandatory" {
				t.Fatalf("enforcement = %q, want mandatory", decisionClient.request.Policy.Enforcement)
			}
		})
	}
}

func TestAgentRunWebhook_GovernanceDecisionOutage(t *testing.T) {
	tests := []struct {
		name         string
		required     bool
		wantErr      bool
		wantWarnings int
	}{
		{name: "mandatory denies", required: true, wantErr: true},
		{name: "advisory observes", required: false, wantWarnings: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := governedPolicy(test.required)
			webhook := &AgentRunWebhook{
				Client: fake.NewClientBuilder().WithScheme(testutil.NewScheme()).WithObjects(policy).Build(),
				DecisionClient: &stubGovernanceDecisionClient{
					err: errors.New("connection refused"),
				},
			}

			warnings, err := webhook.ValidateCreate(context.Background(), governedRun(`{"kind":"independence","input":{}}`))

			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateCreate() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(warnings) != test.wantWarnings {
				t.Fatalf("warnings = %v, want %d", warnings, test.wantWarnings)
			}
		})
	}
}
