package sandbox

import (
	"errors"
	"testing"

	shared "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// fakeIsolator / fakeProvisioner let the registry + Select + policy be unit-tested
// without importing the concrete plugin packages.
type fakeIsolator struct {
	available bool
	hidesHost bool
	kernelIso bool
}

func (f fakeIsolator) Available() (bool, error)                                  { return f.available, nil }
func (f fakeIsolator) Run(*types.SandboxConfig, types.Contribution) (int, error) { return 0, nil }
func (f fakeIsolator) Command(*types.SandboxConfig, types.Contribution) []string { return nil }
func (f fakeIsolator) HidesHost(_ bool, _ string) bool                           { return f.hidesHost }
func (f fakeIsolator) KernelIsolated(_ string) bool                              { return f.kernelIso }

type fakeProvisioner struct{ pinned bool }

func (f fakeProvisioner) Contribute(*types.SandboxConfig) (types.Contribution, error) {
	return types.Contribution{}, nil
}
func (f fakeProvisioner) Pinned() bool { return f.pinned }

func testRegistry(iso fakeIsolator, prov fakeProvisioner) *Registry {
	return NewRegistry().
		RegisterIsolator(types.IsolationBwrap, func() Isolator { return iso }).
		RegisterProvisioner(types.ProvisionNix, func() Provisioner { return prov })
}

func bwrapNixReq() Request {
	return Request{Isolation: types.IsolationBwrap, Provision: types.ProvisionNix, Network: types.NetworkNone}
}

func TestSelect_ResolvesRegisteredPlugins(t *testing.T) {
	reg := testRegistry(fakeIsolator{available: true, hidesHost: true}, fakeProvisioner{pinned: true})
	iso, prov, err := Select(reg, bwrapNixReq(), types.SandboxPolicy{})
	if err != nil || iso == nil || prov == nil {
		t.Fatalf("Select = (%v, %v, %v), want non-nil plugins", iso, prov, err)
	}
}

func TestSelect_UnregisteredFailsClosed(t *testing.T) {
	reg := testRegistry(fakeIsolator{}, fakeProvisioner{})
	if _, _, err := Select(reg, Request{Isolation: "firejail", Provision: types.ProvisionNix}, types.SandboxPolicy{}); !errors.Is(err, ErrNoIsolator) {
		t.Errorf("unregistered isolation err = %v, want ErrNoIsolator", err)
	}
	if _, _, err := Select(reg, Request{Isolation: types.IsolationBwrap, Provision: "apt"}, types.SandboxPolicy{}); !errors.Is(err, ErrNoProvisioner) {
		t.Errorf("unregistered provision err = %v, want ErrNoProvisioner", err)
	}
}

func TestSelect_PolicyFromPluginCapabilities(t *testing.T) {
	cases := []struct {
		name    string
		iso     fakeIsolator
		prov    fakeProvisioner
		net     types.NetworkMethod
		pol     types.SandboxPolicy
		wantErr error
	}{
		{"pinned unmet", fakeIsolator{hidesHost: true}, fakeProvisioner{pinned: false}, types.NetworkNone, types.SandboxPolicy{RequirePinnedProvision: true}, shared.ErrPinnedProvisionUnmet},
		{"host-tools unmet", fakeIsolator{hidesHost: false}, fakeProvisioner{pinned: true}, types.NetworkNone, types.SandboxPolicy{RequireHostToolsUnreachable: true}, shared.ErrHostToolsReachable},
		{"egress unmet (host net)", fakeIsolator{}, fakeProvisioner{}, types.NetworkHost, types.SandboxPolicy{RequireEgressRestricted: true}, shared.ErrEgressUnrestricted},
		{"kernel unmet", fakeIsolator{kernelIso: false}, fakeProvisioner{}, types.NetworkNone, types.SandboxPolicy{RequireKernelIsolation: true}, shared.ErrKernelIsolationUnmet},
		{"all met", fakeIsolator{hidesHost: true, kernelIso: true}, fakeProvisioner{pinned: true}, types.NetworkNone, types.SandboxPolicy{RequirePinnedProvision: true, RequireHostToolsUnreachable: true, RequireEgressRestricted: true, RequireKernelIsolation: true}, nil},
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
		RegisterIsolator(types.IsolationNone, func() Isolator { return fakeIsolator{available: true} }).
		RegisterIsolator(types.IsolationDocker, func() Isolator { return fakeIsolator{available: false} })
	avail := AvailableIsolations(reg)
	if len(avail) != 1 || avail[0] != types.IsolationNone {
		t.Errorf("AvailableIsolations = %v, want [none]", avail)
	}
}

func TestBuiltinProvisioners(t *testing.T) {
	if NewNoneProvisioner().Pinned() || NewCommandProvisioner().Pinned() {
		t.Error("none/command provisioners must report Pinned()=false")
	}
	if c, _ := NewNoneProvisioner().Contribute(&types.SandboxConfig{}); len(c.InitCommands) != 0 {
		t.Error("none provisioner must contribute nothing")
	}
	c, _ := NewCommandProvisioner().Contribute(&types.SandboxConfig{SandboxInitCommands: []string{"echo hi"}})
	if len(c.InitCommands) != 1 || c.InitCommands[0] != "echo hi" {
		t.Errorf("command provisioner InitCommands = %v, want [echo hi]", c.InitCommands)
	}
}
