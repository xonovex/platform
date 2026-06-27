package sandbox

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestEgressIsRestricted(t *testing.T) {
	cases := map[types.NetworkMethod]bool{
		types.NetworkHost:  false,
		types.NetworkNone:  true,
		types.NetworkProxy: true,
	}
	for net, want := range cases {
		if got := EgressIsRestricted(net); got != want {
			t.Errorf("EgressIsRestricted(%q) = %t, want %t", net, got, want)
		}
	}
}

func TestEnforcePolicy(t *testing.T) {
	cases := []struct {
		name    string
		caps    Capabilities
		pol     types.SandboxPolicy
		wantErr error
	}{
		{"no policy allows anything", Capabilities{}, types.SandboxPolicy{}, nil},
		{"pinned met", Capabilities{Pinned: true}, types.SandboxPolicy{RequirePinnedProvision: true}, nil},
		{"pinned unmet", Capabilities{Pinned: false}, types.SandboxPolicy{RequirePinnedProvision: true}, ErrPinnedProvisionUnmet},
		{"host-tools met", Capabilities{HostToolsUnreachable: true}, types.SandboxPolicy{RequireHostToolsUnreachable: true}, nil},
		{"host-tools unmet", Capabilities{HostToolsUnreachable: false}, types.SandboxPolicy{RequireHostToolsUnreachable: true}, ErrHostToolsReachable},
		{"egress met", Capabilities{EgressRestricted: true}, types.SandboxPolicy{RequireEgressRestricted: true}, nil},
		{"egress unmet", Capabilities{EgressRestricted: false}, types.SandboxPolicy{RequireEgressRestricted: true}, ErrEgressUnrestricted},
		{"kernel met", Capabilities{KernelIsolated: true}, types.SandboxPolicy{RequireKernelIsolation: true}, nil},
		{"kernel unmet", Capabilities{KernelIsolated: false}, types.SandboxPolicy{RequireKernelIsolation: true}, ErrKernelIsolationUnmet},
		{"first unmet wins", Capabilities{}, types.SandboxPolicy{RequirePinnedProvision: true, RequireEgressRestricted: true}, ErrPinnedProvisionUnmet},
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
