package webhook

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

func baseRun() *agentv1alpha1.AgentRun {
	return &agentv1alpha1.AgentRun{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"},
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
			RuntimeClassName: strPtr("kata"),
			Image:            "ghcr.io/xonovex/agent:latest",
			Timeout:          &metav1.Duration{Duration: 30 * time.Minute},
		},
	}
}

func basePolicy() *agentv1alpha1.AgentPolicy {
	return &agentv1alpha1.AgentPolicy{
		Spec: agentv1alpha1.AgentPolicySpec{
			Enforced: agentv1alpha1.AgentPolicyEnforced{
				RuntimeClassName:         strPtr("kata"),
				RequireSecurityContext:   true,
				RequireNetworkPolicy:     true,
				MaxTimeout:               &metav1.Duration{Duration: 2 * time.Hour},
				AllowedImages:            []string{"ghcr.io/xonovex/"},
				AllowedRuntimeClassNames: []string{"kata", "gvisor"},
			},
		},
	}
}

func TestEnforcePolicy_AllowsCompliantRun(t *testing.T) {
	run := baseRun()
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil", err)
	}
}

func TestEnforcePolicy_RejectsWrongRuntimeClass(t *testing.T) {
	run := baseRun()
	run.Spec.RuntimeClassName = strPtr("wrong")
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for wrong runtimeClassName")
	}
}

func TestEnforcePolicy_RejectsNilRuntimeClassWhenRequired(t *testing.T) {
	run := baseRun()
	run.Spec.RuntimeClassName = nil
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for nil runtimeClassName when policy requires one")
	}
}

func TestEnforcePolicy_RejectsPrivEsc(t *testing.T) {
	run := baseRun()
	run.Spec.SecurityContext = &corev1.SecurityContext{
		AllowPrivilegeEscalation: boolPtr(true),
	}
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for AllowPrivilegeEscalation=true")
	}
}

func TestEnforcePolicy_RejectsRunAsNonRootFalse(t *testing.T) {
	run := baseRun()
	run.Spec.SecurityContext = &corev1.SecurityContext{
		RunAsNonRoot: boolPtr(false),
	}
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for RunAsNonRoot=false")
	}
}

func TestEnforcePolicy_AllowsSecurityContextWithCompliantValues(t *testing.T) {
	run := baseRun()
	run.Spec.SecurityContext = &corev1.SecurityContext{
		RunAsNonRoot:             boolPtr(true),
		AllowPrivilegeEscalation: boolPtr(false),
	}
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil", err)
	}
}

func TestEnforcePolicy_RejectsDisabledNetworkPolicy(t *testing.T) {
	run := baseRun()
	run.Spec.NetworkPolicy = &agentv1alpha1.AgentNetworkPolicy{Disabled: true}
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for NetworkPolicy.Disabled=true")
	}
}

func TestEnforcePolicy_AllowsEnabledNetworkPolicy(t *testing.T) {
	run := baseRun()
	run.Spec.NetworkPolicy = &agentv1alpha1.AgentNetworkPolicy{Disabled: false}
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil", err)
	}
}

func TestEnforcePolicy_RejectsExceededTimeout(t *testing.T) {
	run := baseRun()
	run.Spec.Timeout = &metav1.Duration{Duration: 5 * time.Hour}
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for timeout exceeding max")
	}
}

func TestEnforcePolicy_AllowsTimeoutWithinMax(t *testing.T) {
	run := baseRun()
	run.Spec.Timeout = &metav1.Duration{Duration: 1 * time.Hour}
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil", err)
	}
}

func TestEnforcePolicy_RejectsDisallowedImage(t *testing.T) {
	run := baseRun()
	run.Spec.Image = "docker.io/evil/image:latest"
	policy := basePolicy()

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for disallowed image")
	}
}

func TestEnforcePolicy_AllowsMatchingImagePrefix(t *testing.T) {
	run := baseRun()
	run.Spec.Image = "ghcr.io/xonovex/custom-agent:v1"
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil", err)
	}
}

func TestEnforcePolicy_RejectsRuntimeClassNotInAllowedList(t *testing.T) {
	run := baseRun()
	run.Spec.RuntimeClassName = strPtr("runc")
	policy := &agentv1alpha1.AgentPolicy{
		Spec: agentv1alpha1.AgentPolicySpec{
			Enforced: agentv1alpha1.AgentPolicyEnforced{
				AllowedRuntimeClassNames: []string{"kata", "gvisor"},
			},
		},
	}

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for runtimeClassName not in allowed list")
	}
}

func TestEnforcePolicy_NoPolicy_AllowsAll(t *testing.T) {
	// When no policy exists, enforcePolicy is never called.
	// This test verifies the webhook's validate() path via direct call
	// with a nil Client (no policy lookup).
	w := &AgentRunWebhook{Client: nil}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Workspace: &agentv1alpha1.WorkspaceSpec{
				Repository: agentv1alpha1.RepositorySpec{
					URL: "https://github.com/example/repo.git",
				},
			},
		},
	}

	_, err := w.validate(context.Background(), run)
	if err != nil {
		t.Errorf("validate() error = %v, want nil (no policy should allow all)", err)
	}
}

func TestEnforcePolicy_EmptyImage_SkipsImageCheck(t *testing.T) {
	run := baseRun()
	run.Spec.Image = ""
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil (empty image should skip check)", err)
	}
}

func TestEnforcePolicy_NilTimeout_SkipsTimeoutCheck(t *testing.T) {
	run := baseRun()
	run.Spec.Timeout = nil
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err != nil {
		t.Errorf("enforcePolicy() error = %v, want nil (nil timeout should skip check)", err)
	}
}
