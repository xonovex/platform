package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAssessA3_DemotesWhenRequiredSignalDegrades(t *testing.T) {
	signals := []Signal{
		{Kind: "oversight.control", Control: "governance-verdict", Outcome: "degraded"},
		{Kind: "oversight.control", Control: "provenance", Outcome: "healthy"},
	}
	assessment := AssessA3(signals, []string{"governance-verdict", "provenance"})
	if !assessment.Detected || assessment.EffectiveAutonomy != "A2" {
		t.Fatalf("assessment = %#v", assessment)
	}
	if len(assessment.FailureCodes) != 1 || assessment.FailureCodes[0] != "oversight-governance-verdict-degraded" {
		t.Fatalf("failure codes = %v", assessment.FailureCodes)
	}
}

func TestSignal_HasNoSensitiveContentSurface(t *testing.T) {
	payload, err := json.Marshal(Signal{
		CorrelationID: "correlation-1", RunReference: "agentrun:test/run",
		Kind: "governance.verdict", Control: "governance-verdict", Outcome: "allow",
	})
	if err != nil {
		t.Fatalf("marshal signal: %v", err)
	}
	serialized := strings.ToLower(string(payload))
	for _, forbidden := range []string{"prompt", "secret", "token", "content"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("signal exposes forbidden field %q: %s", forbidden, payload)
		}
	}
}
