package sandbox

import (
	"errors"
	"testing"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestProvisioningIsPinned(t *testing.T) {
	cases := map[types.ProvisioningMethod]bool{
		types.ProvisionNix:     true,
		types.ProvisionNone:    false,
		types.ProvisionCommand: false,
	}
	for prov, want := range cases {
		if got := ProvisioningIsPinned(prov); got != want {
			t.Errorf("ProvisioningIsPinned(%q)=%t, want %t", prov, got, want)
		}
	}
}

func TestIsolationHidesHost(t *testing.T) {
	cases := []struct {
		iso         types.IsolationMethod
		passthrough bool
		want        bool
	}{
		{types.IsolationNone, false, false},
		{types.IsolationNone, true, false},
		{types.IsolationBwrap, false, true},
		{types.IsolationBwrap, true, false},
		{types.IsolationDocker, false, true},
		{types.IsolationDocker, true, true},
	}
	for _, tc := range cases {
		if got := IsolationHidesHost(tc.iso, tc.passthrough); got != tc.want {
			t.Errorf("IsolationHidesHost(%q, passthrough=%t)=%t, want %t", tc.iso, tc.passthrough, got, tc.want)
		}
	}
}

func TestEgressIsRestricted(t *testing.T) {
	cases := map[types.NetworkMethod]bool{
		types.NetworkHost:  false,
		types.NetworkNone:  true,
		types.NetworkProxy: true,
	}
	for net, want := range cases {
		if got := EgressIsRestricted(net); got != want {
			t.Errorf("EgressIsRestricted(%q)=%t, want %t", net, got, want)
		}
	}
}

func TestKernelIsolated(t *testing.T) {
	cases := []struct {
		iso     types.IsolationMethod
		runtime string
		want    bool
	}{
		{types.IsolationDocker, "runsc", true},
		{types.IsolationDocker, "gvisor", true},
		{types.IsolationDocker, "", false},
		{types.IsolationDocker, "runc", false},
		{types.IsolationBwrap, "runsc", false},
		{types.IsolationNone, "runsc", false},
	}
	for _, tc := range cases {
		if got := KernelIsolated(tc.iso, tc.runtime); got != tc.want {
			t.Errorf("KernelIsolated(%q, %q)=%t, want %t", tc.iso, tc.runtime, got, tc.want)
		}
	}
}

func TestValidAxes(t *testing.T) {
	if err := ValidAxes(types.IsolationDocker, types.ProvisionNix, types.NetworkProxy); err != nil {
		t.Errorf("ValidAxes(valid) = %v, want nil", err)
	}
	cases := []struct {
		name string
		iso  types.IsolationMethod
		prov types.ProvisioningMethod
		net  types.NetworkMethod
	}{
		{"bad isolation", types.IsolationMethod("vm"), types.ProvisionNix, types.NetworkProxy},
		{"bad provisioning", types.IsolationBwrap, types.ProvisioningMethod("apt"), types.NetworkProxy},
		{"bad network", types.IsolationBwrap, types.ProvisionNix, types.NetworkMethod("vpn")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidAxes(tc.iso, tc.prov, tc.net)
			if !errors.Is(err, ErrInvalidAxes) {
				t.Errorf("ValidAxes(%q,%q,%q) err=%v, want ErrInvalidAxes", tc.iso, tc.prov, tc.net, err)
			}
		})
	}
}

func TestEnforcePolicy(t *testing.T) {
	cases := []struct {
		name        string
		iso         types.IsolationMethod
		prov        types.ProvisioningMethod
		net         types.NetworkMethod
		passthrough bool
		runtime     string
		pol         types.SandboxPolicy
		wantErr     error // nil = expect success
	}{
		{
			name: "no policy passes any valid cell",
			iso:  types.IsolationNone, prov: types.ProvisionNone, net: types.NetworkHost,
			pol: types.SandboxPolicy{}, wantErr: nil,
		},
		// RequirePinnedProvisioning
		{
			name: "pinned met by nix",
			iso:  types.IsolationNone, prov: types.ProvisionNix, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequirePinnedProvisioning: true}, wantErr: nil,
		},
		{
			name: "pinned unmet by none",
			iso:  types.IsolationBwrap, prov: types.ProvisionNone, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequirePinnedProvisioning: true}, wantErr: ErrPinnedProvisioningUnmet,
		},
		{
			name: "pinned unmet by command",
			iso:  types.IsolationBwrap, prov: types.ProvisionCommand, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequirePinnedProvisioning: true}, wantErr: ErrPinnedProvisioningUnmet,
		},
		// RequireHostToolsUnreachable
		{
			name: "unreachable met by bwrap no-passthrough",
			iso:  types.IsolationBwrap, prov: types.ProvisionNix, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequireHostToolsUnreachable: true}, wantErr: nil,
		},
		{
			name: "unreachable met by docker",
			iso:  types.IsolationDocker, prov: types.ProvisionNix, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequireHostToolsUnreachable: true}, wantErr: nil,
		},
		{
			name: "unreachable unmet by none",
			iso:  types.IsolationNone, prov: types.ProvisionNix, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequireHostToolsUnreachable: true}, wantErr: ErrHostToolsReachable,
		},
		{
			name: "unreachable unmet by bwrap+passthrough",
			iso:  types.IsolationBwrap, prov: types.ProvisionNix, net: types.NetworkHost, passthrough: true,
			pol: types.SandboxPolicy{RequireHostToolsUnreachable: true}, wantErr: ErrHostToolsReachable,
		},
		// RequireEgressRestricted
		{
			name: "egress met by none",
			iso:  types.IsolationBwrap, prov: types.ProvisionNix, net: types.NetworkNone,
			pol: types.SandboxPolicy{RequireEgressRestricted: true}, wantErr: nil,
		},
		{
			name: "egress met by proxy",
			iso:  types.IsolationBwrap, prov: types.ProvisionNix, net: types.NetworkProxy,
			pol: types.SandboxPolicy{RequireEgressRestricted: true}, wantErr: nil,
		},
		{
			name: "egress unmet by host",
			iso:  types.IsolationBwrap, prov: types.ProvisionNix, net: types.NetworkHost,
			pol: types.SandboxPolicy{RequireEgressRestricted: true}, wantErr: ErrEgressUnrestricted,
		},
		// RequireKernelIsolation
		{
			name: "kernel met by docker+runsc",
			iso:  types.IsolationDocker, prov: types.ProvisionNix, net: types.NetworkHost, runtime: "runsc",
			pol: types.SandboxPolicy{RequireKernelIsolation: true}, wantErr: nil,
		},
		{
			name: "kernel unmet by docker default runc",
			iso:  types.IsolationDocker, prov: types.ProvisionNix, net: types.NetworkHost, runtime: "",
			pol: types.SandboxPolicy{RequireKernelIsolation: true}, wantErr: ErrKernelIsolationUnmet,
		},
		{
			name: "kernel unmet by bwrap",
			iso:  types.IsolationBwrap, prov: types.ProvisionNix, net: types.NetworkHost, runtime: "runsc",
			pol: types.SandboxPolicy{RequireKernelIsolation: true}, wantErr: ErrKernelIsolationUnmet,
		},
		// Invalid axes are rejected before any guarantee check.
		{
			name: "invalid isolation rejected",
			iso:  types.IsolationMethod("vm"), prov: types.ProvisionNix, net: types.NetworkHost,
			pol: types.SandboxPolicy{}, wantErr: ErrInvalidAxes,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := EnforcePolicy(tc.iso, tc.prov, tc.net, tc.passthrough, tc.runtime, tc.pol)
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
