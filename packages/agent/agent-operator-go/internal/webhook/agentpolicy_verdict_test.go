package webhook

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

// TestEnforcePolicy_VerdictGoldens freezes the admission accept/deny outcome and
// the exact denial-reason substring for representative AgentPolicy violations. It
// is the verdict snapshot that locks the operator's (intentionally NOT shared)
// admission engine against silent behavior drift.
func TestEnforcePolicy_VerdictGoldens(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }
	strPtr := func(s string) *string { return &s }
	dur := func(s string) *metav1.Duration {
		d, err := time.ParseDuration(s)
		if err != nil {
			t.Fatalf("bad duration %q: %v", s, err)
		}
		return &metav1.Duration{Duration: d}
	}

	cases := []struct {
		name       string
		enforced   agentv1alpha1.AgentPolicyEnforced
		run        agentv1alpha1.AgentRunSpec
		wantReason string // "" means accept
	}{
		{
			name:       "compliant-accepts",
			enforced:   agentv1alpha1.AgentPolicyEnforced{RequireSecurityContext: true, RequireNetworkPolicy: true},
			run:        agentv1alpha1.AgentRunSpec{},
			wantReason: "",
		},
		{
			name:       "wrong-runtimeclass",
			enforced:   agentv1alpha1.AgentPolicyEnforced{RuntimeClassName: strPtr("kata")},
			run:        agentv1alpha1.AgentRunSpec{RuntimeClassName: strPtr("runc")},
			wantReason: "policy requires runtimeClassName",
		},
		{
			name:       "runtimeclass-not-in-allowlist",
			enforced:   agentv1alpha1.AgentPolicyEnforced{AllowedRuntimeClassNames: []string{"kata", "gvisor"}},
			run:        agentv1alpha1.AgentRunSpec{RuntimeClassName: strPtr("runc")},
			wantReason: "runtimeClassName must be one of",
		},
		{
			name:       "priv-esc-denied",
			enforced:   agentv1alpha1.AgentPolicyEnforced{RequireSecurityContext: true},
			run:        agentv1alpha1.AgentRunSpec{SecurityContext: &corev1.SecurityContext{AllowPrivilegeEscalation: boolPtr(true)}},
			wantReason: "prohibits AllowPrivilegeEscalation=true",
		},
		{
			name:       "run-as-root-denied",
			enforced:   agentv1alpha1.AgentPolicyEnforced{RequireSecurityContext: true},
			run:        agentv1alpha1.AgentRunSpec{SecurityContext: &corev1.SecurityContext{RunAsNonRoot: boolPtr(false)}},
			wantReason: "requires RunAsNonRoot=true",
		},
		{
			name:       "netpol-disabled-denied",
			enforced:   agentv1alpha1.AgentPolicyEnforced{RequireNetworkPolicy: true},
			run:        agentv1alpha1.AgentRunSpec{NetworkPolicy: &agentv1alpha1.AgentNetworkPolicy{Disabled: true}},
			wantReason: "requires NetworkPolicy to be enabled",
		},
		{
			name:       "timeout-exceeded-denied",
			enforced:   agentv1alpha1.AgentPolicyEnforced{MaxTimeout: dur("1h")},
			run:        agentv1alpha1.AgentRunSpec{Timeout: dur("2h")},
			wantReason: "exceeds policy maximum",
		},
		{
			name:       "image-not-allowed-denied",
			enforced:   agentv1alpha1.AgentPolicyEnforced{AllowedImages: []string{"ghcr.io/xonovex/"}},
			run:        agentv1alpha1.AgentRunSpec{Image: "docker.io/evil/img:latest"},
			wantReason: "is not in the allowed images list",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := &agentv1alpha1.AgentRun{
				ObjectMeta: metav1.ObjectMeta{Name: "r", Namespace: "default"},
				Spec:       tc.run,
			}
			policy := &agentv1alpha1.AgentPolicy{Spec: agentv1alpha1.AgentPolicySpec{Enforced: tc.enforced}}

			err := enforcePolicy(run, policy)
			if tc.wantReason == "" {
				if err != nil {
					t.Fatalf("verdict = deny(%q), want accept", err.Error())
				}
				return
			}
			if err == nil {
				t.Fatalf("verdict = accept, want deny containing %q", tc.wantReason)
			}
			if !strings.Contains(err.Error(), tc.wantReason) {
				t.Errorf("denial reason = %q, want substring %q", err.Error(), tc.wantReason)
			}
		})
	}
}
