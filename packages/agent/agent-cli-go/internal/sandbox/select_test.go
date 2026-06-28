package sandbox

import (
	"errors"
	"testing"

	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	netshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/network/shared"
	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
)

// fakeIsolator / fakeProvisioner let the registry + Select + policy be unit-tested
// without importing the concrete plugin packages.
type fakeIsolator struct {
	available bool
	hidesHost bool
	kernelIso bool
}

func (f fakeIsolator) Available() (bool, error)                                     { return f.available, nil }
func (f fakeIsolator) Run(isoshared.RunConfig, provision.Contribution) (int, error) { return 0, nil }
func (f fakeIsolator) Command(isoshared.RunConfig, provision.Contribution) []string { return nil }
func (f fakeIsolator) TerminalCommand(isoshared.RunConfig, provision.Contribution) ([]string, []string) {
	return nil, nil
}
func (f fakeIsolator) HidesHost(_ bool, _ string) bool { return f.hidesHost }
func (f fakeIsolator) KernelIsolated(_ string) bool    { return f.kernelIso }

type fakeProvisioner struct{ pinned bool }

func (f fakeProvisioner) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (f fakeProvisioner) Pinned() bool { return f.pinned }

func testRegistry(iso fakeIsolator, prov fakeProvisioner) *Registry {
	return NewRegistry().
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Isolator { return iso }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return prov })
}

func bwrapNixReq() Request {
	return Request{Isolation: isolation.IsolationBwrap, Provision: provision.ProvisionNix, Network: netshared.ModeNone}
}

func TestSelect_ResolvesRegisteredPlugins(t *testing.T) {
	reg := testRegistry(fakeIsolator{available: true, hidesHost: true}, fakeProvisioner{pinned: true})
	iso, prov, err := Select(reg, bwrapNixReq(), policy.SandboxPolicy{})
	if err != nil || iso == nil || prov == nil {
		t.Fatalf("Select = (%v, %v, %v), want non-nil plugins", iso, prov, err)
	}
}

func TestSelect_UnregisteredFailsClosed(t *testing.T) {
	reg := testRegistry(fakeIsolator{}, fakeProvisioner{})
	if _, _, err := Select(reg, Request{Isolation: "firejail", Provision: provision.ProvisionNix}, policy.SandboxPolicy{}); !errors.Is(err, ErrNoIsolator) {
		t.Errorf("unregistered isolation err = %v, want ErrNoIsolator", err)
	}
	if _, _, err := Select(reg, Request{Isolation: isolation.IsolationBwrap, Provision: "apt"}, policy.SandboxPolicy{}); !errors.Is(err, ErrNoProvisioner) {
		t.Errorf("unregistered provision err = %v, want ErrNoProvisioner", err)
	}
}

func TestSelect_PolicyFromPluginCapabilities(t *testing.T) {
	cases := []struct {
		name    string
		iso     fakeIsolator
		prov    fakeProvisioner
		net     netshared.Mode
		pol     policy.SandboxPolicy
		wantErr error
	}{
		{"pinned unmet", fakeIsolator{hidesHost: true}, fakeProvisioner{pinned: false}, netshared.ModeNone, policy.SandboxPolicy{RequirePinnedProvision: true}, policy.ErrPinnedProvisionUnmet},
		{"host-tools unmet", fakeIsolator{hidesHost: false}, fakeProvisioner{pinned: true}, netshared.ModeNone, policy.SandboxPolicy{RequireHostToolsUnreachable: true}, policy.ErrHostToolsReachable},
		{"egress unmet (host net)", fakeIsolator{}, fakeProvisioner{}, netshared.ModeHost, policy.SandboxPolicy{RequireEgressRestricted: true}, policy.ErrEgressUnrestricted},
		{"kernel unmet", fakeIsolator{kernelIso: false}, fakeProvisioner{}, netshared.ModeNone, policy.SandboxPolicy{RequireKernelIsolation: true}, policy.ErrKernelIsolationUnmet},
		{"all met", fakeIsolator{hidesHost: true, kernelIso: true}, fakeProvisioner{pinned: true}, netshared.ModeNone, policy.SandboxPolicy{RequirePinnedProvision: true, RequireHostToolsUnreachable: true, RequireEgressRestricted: true, RequireKernelIsolation: true}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := testRegistry(tc.iso, tc.prov)
			req := bwrapNixReq()
			req.Network = tc.net
			_, _, err := Select(reg, req, tc.pol)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("Select = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Select err = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestAvailableIsolations(t *testing.T) {
	reg := NewRegistry().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Isolator { return fakeIsolator{available: true} }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return fakeIsolator{available: false} })
	avail := AvailableIsolations(reg)
	if len(avail) != 1 || avail[0] != isolation.IsolationNone {
		t.Errorf("AvailableIsolations = %v, want [none]", avail)
	}
}
