package webhook

import (
	"context"
	"errors"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

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

func TestEnforcePolicy_RejectsResourceLimitAboveMaximum(t *testing.T) {
	run := baseRun()
	run.Spec.Resources.Limits = corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("2"),
	}
	maximum := corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("500m"),
	}
	policy := basePolicy()
	policy.Spec.Enforced.MaxResources = &maximum

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for resource limit above policy maximum")
	}
}

func TestEnforcePolicy_RejectsMissingResourceLimit(t *testing.T) {
	run := baseRun()
	maximum := corev1.ResourceList{
		corev1.ResourceMemory: resource.MustParse("1Gi"),
	}
	policy := basePolicy()
	policy.Spec.Enforced.MaxResources = &maximum

	err := enforcePolicy(run, policy)
	if err == nil {
		t.Error("enforcePolicy() expected error for missing bounded resource limit")
	}
}

func TestEnforcePolicy_AllowsResourcesWithinMaximum(t *testing.T) {
	run := baseRun()
	run.Spec.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("250m"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("500m"),
		},
	}
	maximum := corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("1"),
	}
	policy := basePolicy()
	policy.Spec.Enforced.MaxResources = &maximum

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

func TestEnforcePolicy_RejectsEmptyImageWhenAllowlistIsConfigured(t *testing.T) {
	run := baseRun()
	run.Spec.Image = ""
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err == nil {
		t.Error("enforcePolicy() expected error for an image that cannot be checked against the allowlist")
	}
}

func TestEnforcePolicy_RejectsNilTimeoutWhenMaximumIsConfigured(t *testing.T) {
	run := baseRun()
	run.Spec.Timeout = nil
	policy := basePolicy()

	if err := enforcePolicy(run, policy); err == nil {
		t.Error("enforcePolicy() expected error for a timeout that cannot be checked against the maximum")
	}
}

func TestAgentRunWebhook_Default_AppliesNamespacePolicyDefaults(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}
	policy := &agentv1alpha1.AgentPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "namespace-policy", Namespace: "test-ns"},
		Spec: agentv1alpha1.AgentPolicySpec{Defaults: agentv1alpha1.AgentPolicyDefaults{
			Image:            "ghcr.io/xonovex/agent@sha256:abc",
			Timeout:          &metav1.Duration{Duration: 20 * time.Minute},
			RuntimeClassName: strPtr("kata"),
		}},
	}
	w := &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(policy).Build()}
	run := &agentv1alpha1.AgentRun{ObjectMeta: metav1.ObjectMeta{Namespace: "test-ns"}}

	if err := w.Default(context.Background(), run); err != nil {
		t.Fatalf("Default() error = %v", err)
	}

	if run.Spec.Image != policy.Spec.Defaults.Image {
		t.Errorf("Image = %q, want %q", run.Spec.Image, policy.Spec.Defaults.Image)
	}
	if run.Spec.Timeout == nil || run.Spec.Timeout.Duration != 20*time.Minute {
		t.Errorf("Timeout = %v, want 20m", run.Spec.Timeout)
	}
	if run.Spec.RuntimeClassName == nil || *run.Spec.RuntimeClassName != "kata" {
		t.Errorf("RuntimeClassName = %v, want kata", run.Spec.RuntimeClassName)
	}
}

func TestAgentRunWebhook_Validate_RejectsMultipleNamespacePolicies(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}
	policies := []client.Object{
		&agentv1alpha1.AgentPolicy{ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "test-ns"}},
		&agentv1alpha1.AgentPolicy{ObjectMeta: metav1.ObjectMeta{Name: "second", Namespace: "test-ns"}},
	}
	w := &AgentRunWebhook{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(policies...).Build()}

	_, err := w.validate(context.Background(), baseRun())
	if err == nil {
		t.Error("validate() expected error for ambiguous namespace policy authority")
	}
}

func TestAgentRunWebhook_Validate_FailsClosedWhenPolicyLookupFails(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := agentv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}
	baseClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	failingClient := interceptor.NewClient(baseClient, interceptor.Funcs{
		List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
			return errors.New("policy API unavailable")
		},
	})
	w := &AgentRunWebhook{Client: failingClient}

	_, err := w.validate(context.Background(), baseRun())
	if err == nil {
		t.Error("validate() expected error when policy lookup is unavailable")
	}
}
