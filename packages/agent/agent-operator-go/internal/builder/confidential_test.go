package builder

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	agentv1alpha1 "github.com/xonovex/platform/packages/agent/agent-operator-go/api/v1alpha1"
)

func TestTEERuntimeClassName_AMDSEVSNP(t *testing.T) {
	cc := &agentv1alpha1.ConfidentialComputingSpec{TEE: agentv1alpha1.TEETypeAMDSEVSNP}
	got := TEERuntimeClassName(cc)
	if got == nil || *got != "kata-cc" {
		t.Errorf("TEERuntimeClassName(amd-sev-snp) = %v, want kata-cc", got)
	}
}

func TestTEERuntimeClassName_IntelTDX(t *testing.T) {
	cc := &agentv1alpha1.ConfidentialComputingSpec{TEE: agentv1alpha1.TEETypeIntelTDX}
	got := TEERuntimeClassName(cc)
	if got == nil || *got != "kata-tdx" {
		t.Errorf("TEERuntimeClassName(intel-tdx) = %v, want kata-tdx", got)
	}
}

func TestTEERuntimeClassName_Override(t *testing.T) {
	override := "my-custom-runtime"
	cc := &agentv1alpha1.ConfidentialComputingSpec{
		TEE:                     agentv1alpha1.TEETypeAMDSEVSNP,
		OverrideRuntimeClassName: &override,
	}
	got := TEERuntimeClassName(cc)
	if got == nil || *got != "my-custom-runtime" {
		t.Errorf("TEERuntimeClassName(override) = %v, want my-custom-runtime", got)
	}
}

func TestTEERuntimeClassName_Nil(t *testing.T) {
	got := TEERuntimeClassName(nil)
	if got != nil {
		t.Errorf("TEERuntimeClassName(nil) = %v, want nil", got)
	}
}

func TestTEENodeAffinity_AMDSEVSNP(t *testing.T) {
	cc := &agentv1alpha1.ConfidentialComputingSpec{TEE: agentv1alpha1.TEETypeAMDSEVSNP}
	got := TEENodeAffinity(cc)
	if got == nil {
		t.Fatal("TEENodeAffinity(amd-sev-snp) = nil, want non-nil")
	}
	terms := got.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	if len(terms) != 1 {
		t.Fatalf("expected 1 NodeSelectorTerm, got %d", len(terms))
	}
	expr := terms[0].MatchExpressions
	if len(expr) != 1 {
		t.Fatalf("expected 1 MatchExpression, got %d", len(expr))
	}
	if expr[0].Key != aksConfidentialNodeLabel {
		t.Errorf("expected key %q, got %q", aksConfidentialNodeLabel, expr[0].Key)
	}
	if expr[0].Operator != corev1.NodeSelectorOpIn {
		t.Errorf("expected operator In, got %v", expr[0].Operator)
	}
	if len(expr[0].Values) != 1 || expr[0].Values[0] != "true" {
		t.Errorf("expected values [true], got %v", expr[0].Values)
	}
}

func TestTEENodeAffinity_DisableNodeAffinity(t *testing.T) {
	cc := &agentv1alpha1.ConfidentialComputingSpec{
		TEE:                 agentv1alpha1.TEETypeAMDSEVSNP,
		DisableNodeAffinity: true,
	}
	got := TEENodeAffinity(cc)
	if got != nil {
		t.Errorf("TEENodeAffinity(disabled) = %v, want nil", got)
	}
}

func TestTEENodeAffinity_Nil(t *testing.T) {
	got := TEENodeAffinity(nil)
	if got != nil {
		t.Errorf("TEENodeAffinity(nil) = %v, want nil", got)
	}
}
