package policy

import (
	"errors"
	"testing"
)

func TestEnforcePolicy(t *testing.T) {
	cases := []struct {
		name    string
		caps    Capabilities
		pol     SandboxPolicy
		wantErr error
	}{
		{"no policy allows anything", Capabilities{}, SandboxPolicy{}, nil},
		{"pinned met", Capabilities{Pinned: true}, SandboxPolicy{RequirePinnedProvision: true}, nil},
		{"pinned unmet", Capabilities{Pinned: false}, SandboxPolicy{RequirePinnedProvision: true}, ErrPinnedProvisionUnmet},
		{"host-tools met", Capabilities{HostToolsUnreachable: true}, SandboxPolicy{RequireHostToolsUnreachable: true}, nil},
		{"host-tools unmet", Capabilities{HostToolsUnreachable: false}, SandboxPolicy{RequireHostToolsUnreachable: true}, ErrHostToolsReachable},
		{"egress met", Capabilities{EgressRestricted: true}, SandboxPolicy{RequireEgressRestricted: true}, nil},
		{"egress unmet", Capabilities{EgressRestricted: false}, SandboxPolicy{RequireEgressRestricted: true}, ErrEgressUnrestricted},
		{"kernel met", Capabilities{KernelIsolated: true}, SandboxPolicy{RequireKernelIsolation: true}, nil},
		{"kernel unmet", Capabilities{KernelIsolated: false}, SandboxPolicy{RequireKernelIsolation: true}, ErrKernelIsolationUnmet},
		{"first unmet wins", Capabilities{}, SandboxPolicy{RequirePinnedProvision: true, RequireEgressRestricted: true}, ErrPinnedProvisionUnmet},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := EnforcePolicy(tc.caps, tc.pol)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("EnforcePolicy() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("EnforcePolicy() = %v, want errors.Is %v", err, tc.wantErr)
			}
		})
	}
}

func TestEvaluateReportsEveryGuarantee(t *testing.T) {
	outcomes := Evaluate(Capabilities{Pinned: true, KernelIsolated: true}, SandboxPolicy{RequireHostToolsUnreachable: true, RequireEgressRestricted: true})
	want := []struct {
		guarantee Guarantee
		required  bool
		provided  bool
		unmet     error
	}{
		{GuaranteePinnedProvision, false, true, nil},
		{GuaranteeHostToolsUnreachable, true, false, ErrHostToolsReachable},
		{GuaranteeEgressRestricted, true, false, ErrEgressUnrestricted},
		{GuaranteeKernelIsolation, false, true, nil},
	}
	if len(outcomes) != len(want) {
		t.Fatalf("Evaluate() returned %d outcomes, want %d", len(outcomes), len(want))
	}
	for index, outcome := range outcomes {
		w := want[index]
		if outcome.Guarantee != w.guarantee || outcome.Required != w.required || outcome.Provided != w.provided {
			t.Errorf("outcome %d = %+v, want %+v", index, outcome, w)
		}
		if w.unmet == nil && (outcome.Err != nil || !outcome.Met()) {
			t.Errorf("%s: Err = %v, Met = %v, want a met guarantee", outcome.Guarantee, outcome.Err, outcome.Met())
		}
		if w.unmet != nil && (!errors.Is(outcome.Err, w.unmet) || outcome.Met()) {
			t.Errorf("%s: Err = %v, want errors.Is %v", outcome.Guarantee, outcome.Err, w.unmet)
		}
	}
}
