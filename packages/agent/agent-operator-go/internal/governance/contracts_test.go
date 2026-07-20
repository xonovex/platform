package governance

import "testing"

func TestContentAddressedEvidenceReference(t *testing.T) {
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tests := []struct {
		name      string
		reference string
		want      bool
	}{
		{name: "provider reference and lowercase digest", reference: "pvc://evidence/verdicts.jsonl#sha256:" + digest, want: true},
		{name: "missing provider reference", reference: "#sha256:" + digest},
		{name: "missing algorithm", reference: "pvc://evidence/verdicts.jsonl#" + digest},
		{name: "uppercase digest", reference: "pvc://evidence/verdicts.jsonl#sha256:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},
		{name: "short digest", reference: "pvc://evidence/verdicts.jsonl#sha256:aaaa"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ContentAddressedEvidenceReference(test.reference); got != test.want {
				t.Fatalf("ContentAddressedEvidenceReference(%q) = %v, want %v", test.reference, got, test.want)
			}
		})
	}
}
