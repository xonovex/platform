package resolver

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestApplyHarnessDefaults_NoHarness(t *testing.T) {
	run := &agentv1alpha1.AgentRun{
		Spec: agentv1alpha1.AgentRunSpec{},
	}

	d := ApplyHarnessDefaults(run, nil)

	if d.Image != DefaultImage {
		t.Errorf("Image = %q, want %q", d.Image, DefaultImage)
	}
	if d.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", d.Timeout, DefaultTimeout)
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
