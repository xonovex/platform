package resolver

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
)

func TestApplyHarnessDefaults_NoHarness(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}

	d := ApplyHarnessDefaults(run, nil)

	if d.Image != "" {
		t.Errorf("Image = %q, want no unsafe built-in image", d.Image)
	}
	if d.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", d.Timeout, DefaultTimeout)
	}
}

func TestApplyHarnessDefaults_AppliesResourcesAndMergesEnvironment(t *testing.T) {
	run := &agentv1alpha1.AgentRun{Spec: agentv1alpha1.AgentRunSpec{Env: []corev1.EnvVar{
		{Name: "OVERRIDE", Value: "run"},
	}}}
	harness := &agentv1alpha1.AgentHarness{Spec: agentv1alpha1.AgentSpec{
		DefaultResources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("250m")},
			Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
		},
		Env: []corev1.EnvVar{
			{Name: "HARNESS_ONLY", Value: "default"},
			{Name: "OVERRIDE", Value: "harness"},
		},
	}}

	ApplyHarnessDefaults(run, harness)

	if run.Spec.Resources.Limits.Cpu().String() != "1" {
		t.Fatalf("CPU limit = %q, want 1", run.Spec.Resources.Limits.Cpu().String())
	}
	if len(run.Spec.Env) != 2 || run.Spec.Env[0].Name != "HARNESS_ONLY" || run.Spec.Env[1].Value != "run" {
		t.Fatalf("Env = %#v, want harness default with run override", run.Spec.Env)
	}

	ApplyHarnessDefaults(run, harness)
	if len(run.Spec.Env) != 2 {
		t.Fatalf("second ApplyHarnessDefaults() duplicated environment: %#v", run.Spec.Env)
	}
}

func TestApplyHarnessDefaults_HarnessOverridesBuiltins(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}
	harnessTimeout := metav1.Duration{Duration: 30 * time.Minute}
	harness := &agentv1alpha1.AgentHarness{
		Spec: agentv1alpha1.AgentSpec{
			DefaultImage:   "custom:latest",
			DefaultTimeout: &harnessTimeout,
		},
	}

	d := ApplyHarnessDefaults(run, harness)

	if d.Image != "custom:latest" {
		t.Errorf("Image = %q, want %q", d.Image, "custom:latest")
	}
	if d.Timeout != 30*time.Minute {
		t.Errorf("Timeout = %v, want %v", d.Timeout, 30*time.Minute)
	}
}

func TestApplyHarnessDefaults_RunOverridesHarness(t *testing.T) {
	runTimeout := metav1.Duration{Duration: 15 * time.Minute}
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			Image:   "run-image:v1",
			Timeout: &runTimeout,
		},
	}
	harnessTimeout := metav1.Duration{Duration: 30 * time.Minute}
	harness := &agentv1alpha1.AgentHarness{
		Spec: agentv1alpha1.AgentSpec{
			DefaultImage:   "harness-image:v1",
			DefaultTimeout: &harnessTimeout,
		},
	}

	d := ApplyHarnessDefaults(run, harness)

	if d.Image != "run-image:v1" {
		t.Errorf("Image = %q, want %q", d.Image, "run-image:v1")
	}
	if d.Timeout != 15*time.Minute {
		t.Errorf("Timeout = %v, want %v", d.Timeout, 15*time.Minute)
	}
}

func TestApplyHarnessDefaults_RuntimeClassName(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}
	kata := "kata"
	harness := &agentv1alpha1.AgentHarness{
		Spec: agentv1alpha1.AgentSpec{
			DefaultRuntimeClassName: &kata,
		},
	}

	ApplyHarnessDefaults(run, harness)

	if run.Spec.RuntimeClassName == nil || *run.Spec.RuntimeClassName != "kata" {
		t.Errorf("RuntimeClassName = %v, want kata", run.Spec.RuntimeClassName)
	}
}

func TestApplyHarnessDefaults_RuntimeClassName_RunTakesPrecedence(t *testing.T) {
	gvisor := "gvisor"
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{
			RuntimeClassName: &gvisor,
		},
	}
	kata := "kata"
	harness := &agentv1alpha1.AgentHarness{
		Spec: agentv1alpha1.AgentSpec{
			DefaultRuntimeClassName: &kata,
		},
	}

	ApplyHarnessDefaults(run, harness)

	if *run.Spec.RuntimeClassName != "gvisor" {
		t.Errorf("RuntimeClassName = %q, want gvisor", *run.Spec.RuntimeClassName)
	}
}

func TestApplyHarnessDefaults_FillsTheSandboxDefaultsARunLeavesEmpty(t *testing.T) {
	ttl := int32(600)
	harness := &agentv1alpha1.AgentHarness{Spec: agentv1alpha1.AgentSpec{
		DefaultSecurityContext:         &corev1.SecurityContext{ReadOnlyRootFilesystem: boolPointer(true)},
		DefaultPodSecurityContext:      &corev1.PodSecurityContext{FSGroup: int64Pointer(1000)},
		DefaultNetworkPolicy:           &agentv1alpha1.AgentNetworkPolicy{Disabled: true},
		DefaultTTLSecondsAfterFinished: &ttl,
	}}
	run := &agentv1alpha1.AgentRun{}

	d := ApplyHarnessDefaults(run, harness)

	if run.Spec.SecurityContext == nil || run.Spec.PodSecurityContext == nil {
		t.Errorf("security contexts = %v, %v, want the harness defaults", run.Spec.SecurityContext, run.Spec.PodSecurityContext)
	}
	if d.NetworkPolicy == nil || !d.NetworkPolicy.Disabled {
		t.Errorf("NetworkPolicy = %v, want the harness default", d.NetworkPolicy)
	}
	if d.TTL == nil || *d.TTL != ttl {
		t.Errorf("TTL = %v, want %d", d.TTL, ttl)
	}
	*harness.Spec.DefaultTTLSecondsAfterFinished = 1
	if *d.TTL != 600 {
		t.Error("the run shares the harness TTL instead of a copy")
	}
}

func boolPointer(value bool) *bool { return &value }

func int64Pointer(value int64) *int64 { return &value }
