package sandbox

import (
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

// fakeConfinement / fakeProvisioner let the registry + Select + policy be
// unit-tested without importing the concrete plugin packages. The selection
// tests register a Confinement alone, which proves that selection needs no
// Launcher.
type fakeConfinement struct {
	hidesHost bool
	kernelIso bool
	// sharesNetwork makes the double unable to confine the network; the zero
	// value confines it.
	sharesNetwork bool
	// noTransport makes the double declare no egress transport; the zero value
	// declares the unix socket.
	noTransport bool
	// noBinds makes the double refuse host binds; the zero value accepts them.
	noBinds bool
}

func (f fakeConfinement) PinnedProvision(_ provision.ProvisionMethod, provisionerPinned bool, _ string) bool {
	return provisionerPinned
}
func (f fakeConfinement) HidesHost(_ bool, _ string) bool { return f.hidesHost }
func (f fakeConfinement) KernelIsolated(_ string) bool    { return f.kernelIso }
func (f fakeConfinement) EgressTransport() network.EgressTransport {
	if f.noTransport {
		return network.TransportNone
	}
	return network.TransportUnixSocket
}
func (f fakeConfinement) ConfinesNetwork() bool  { return !f.sharesNetwork }
func (f fakeConfinement) Bounded() bool          { return true }
func (f fakeConfinement) AcceptsHostBinds() bool { return !f.noBinds }

// fakeIsolator adds a Launcher to the confinement, for the availability probe.
type fakeIsolator struct {
	fakeConfinement
	available bool
}

func (f fakeIsolator) Available() (bool, error)                                     { return f.available, nil }
func (f fakeIsolator) Run(isoshared.RunConfig, provision.Contribution) (int, error) { return 0, nil }
func (f fakeIsolator) Command(isoshared.RunConfig, provision.Contribution) ([]string, error) {
	return nil, nil
}
func (f fakeIsolator) TerminalCommand(isoshared.RunConfig, provision.Contribution) ([]string, []string, error) {
	return nil, nil, nil
}
func (f fakeIsolator) CheckAgent(isoshared.RunConfig) error     { return nil }
func (f fakeIsolator) StopCommand(isoshared.RunConfig) []string { return nil }

type fakeProvisioner struct{ pinned bool }

func (f fakeProvisioner) Contribute(provshared.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (f fakeProvisioner) Pinned() bool { return f.pinned }
func (f fakeProvisioner) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}

func testRegistry(iso fakeConfinement, prov fakeProvisioner) *Registry[isoshared.Confinement] {
	return NewRegistry[isoshared.Confinement]().
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Confinement { return iso }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return prov })
}

func bwrapNixReq() Request {
	return Request{Isolation: isolation.IsolationBwrap, Provision: provision.ProvisionNix, Network: network.NetworkNone}
}

func TestSelect_ResolvesRegisteredPlugins(t *testing.T) {
	reg := testRegistry(fakeConfinement{hidesHost: true}, fakeProvisioner{pinned: true})
	choice, err := Select(reg, bwrapNixReq(), policy.SandboxPolicy{})
	if err != nil || choice.Isolator == nil || choice.Provisioner == nil {
		t.Fatalf("Select = (%v, %v), want non-nil plugins", choice, err)
	}
}

func TestEvaluate_ReportsGuaranteesWithoutDemandingThem(t *testing.T) {
	reg := testRegistry(fakeConfinement{hidesHost: true}, fakeProvisioner{pinned: true})
	choice, err := Evaluate(reg, bwrapNixReq())
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !choice.Capabilities.Pinned || !choice.Capabilities.HostToolsUnreachable || choice.Capabilities.KernelIsolated {
		t.Errorf("Capabilities = %+v, want pinned and host tools unreachable, no kernel boundary", choice.Capabilities)
	}
	if _, err := Select(reg, bwrapNixReq(), policy.SandboxPolicy{RequireKernelIsolation: true}); !errors.Is(err, policy.ErrKernelIsolationUnmet) {
		t.Errorf("Select(require kernel isolation) error = %v, want %v", err, policy.ErrKernelIsolationUnmet)
	}
	if _, err := Evaluate(reg, Request{Isolation: "firejail", Provision: provision.ProvisionNix}); !errors.Is(err, ErrNoIsolator) {
		t.Errorf("Evaluate(unregistered) error = %v, want %v", err, ErrNoIsolator)
	}
}

func TestSelect_UnregisteredFailsClosed(t *testing.T) {
	reg := testRegistry(fakeConfinement{}, fakeProvisioner{})
	if _, err := Select(reg, Request{Isolation: "firejail", Provision: provision.ProvisionNix}, policy.SandboxPolicy{}); !errors.Is(err, ErrNoIsolator) {
		t.Errorf("unregistered isolation err = %v, want ErrNoIsolator", err)
	}
	if _, err := Select(reg, Request{Isolation: isolation.IsolationBwrap, Provision: "apt"}, policy.SandboxPolicy{}); !errors.Is(err, ErrNoProvisioner) {
		t.Errorf("unregistered provision err = %v, want ErrNoProvisioner", err)
	}
}

func TestSelect_PolicyFromPluginCapabilities(t *testing.T) {
	cases := []struct {
		name    string
		iso     fakeConfinement
		prov    fakeProvisioner
		net     network.NetworkMethod
		pol     policy.SandboxPolicy
		wantErr error
	}{
		{"pinned unmet", fakeConfinement{hidesHost: true}, fakeProvisioner{pinned: false}, network.NetworkNone, policy.SandboxPolicy{RequirePinnedProvision: true}, policy.ErrPinnedProvisionUnmet},
		{"host-tools unmet", fakeConfinement{hidesHost: false}, fakeProvisioner{pinned: true}, network.NetworkNone, policy.SandboxPolicy{RequireHostToolsUnreachable: true}, policy.ErrHostToolsReachable},
		{"egress unmet (host net)", fakeConfinement{}, fakeProvisioner{}, network.NetworkHost, policy.SandboxPolicy{RequireEgressRestricted: true}, policy.ErrEgressUnrestricted},
		{"kernel unmet", fakeConfinement{kernelIso: false}, fakeProvisioner{}, network.NetworkNone, policy.SandboxPolicy{RequireKernelIsolation: true}, policy.ErrKernelIsolationUnmet},
		{"all met", fakeConfinement{hidesHost: true, kernelIso: true}, fakeProvisioner{pinned: true}, network.NetworkNone, policy.SandboxPolicy{RequirePinnedProvision: true, RequireHostToolsUnreachable: true, RequireEgressRestricted: true, RequireKernelIsolation: true}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reg := testRegistry(tc.iso, tc.prov)
			req := bwrapNixReq()
			req.Network = tc.net
			_, err := Select(reg, req, tc.pol)
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

func TestSelect_RejectsCustomBindsWhenHostToolsMustBeUnreachable(t *testing.T) {
	reg := testRegistry(fakeConfinement{hidesHost: true}, fakeProvisioner{pinned: true})
	req := bwrapNixReq()
	req.HasCustomBinds = true

	_, err := Select(reg, req, policy.SandboxPolicy{RequireHostToolsUnreachable: true})

	if !errors.Is(err, policy.ErrHostToolsReachable) {
		t.Fatalf("Select() error = %v, want %v", err, policy.ErrHostToolsReachable)
	}
}

func TestEvaluate_RejectsBindPathsForAnIsolationWithoutHostBinds(t *testing.T) {
	reg := testRegistry(fakeConfinement{noBinds: true}, fakeProvisioner{})
	req := bwrapNixReq()

	if _, err := Evaluate(reg, req); err != nil {
		t.Fatalf("Evaluate(no bind paths) error = %v, want nil", err)
	}
	req.HasCustomBinds = true
	if _, err := Evaluate(reg, req); err == nil || !strings.Contains(err.Error(), "cannot bind a host path") {
		t.Fatalf("Evaluate(bind paths) error = %v, want the refused bind", err)
	}
}

func TestSelect_RejectsConfinedNetworkWithoutANetworkNamespace(t *testing.T) {
	reg := testRegistry(fakeConfinement{sharesNetwork: true}, fakeProvisioner{})

	if _, err := Select(reg, bwrapNixReq(), policy.SandboxPolicy{}); err == nil {
		t.Fatal("Select(network none, shared network) error = nil, want a confinement error")
	}

	req := bwrapNixReq()
	req.Network = network.NetworkHost
	if _, err := Select(reg, req, policy.SandboxPolicy{}); err != nil {
		t.Fatalf("Select(network host, shared network) error = %v, want nil", err)
	}
}

func TestSelect_ManagedEgressNeedsATransport(t *testing.T) {
	req := bwrapNixReq()
	req.Network = network.NetworkManaged
	noTransport := testRegistry(fakeConfinement{noTransport: true}, fakeProvisioner{})
	if _, err := Select(noTransport, req, policy.SandboxPolicy{}); !errors.Is(err, network.ErrManagedEgressUnenforceable) {
		t.Errorf("an isolator without a transport must refuse managed whatever the policy: %v", err)
	}
	withTransport := testRegistry(fakeConfinement{}, fakeProvisioner{})
	if _, err := Select(withTransport, req, policy.SandboxPolicy{RequireEgressRestricted: true}); err != nil {
		t.Errorf("managed over a transport must satisfy RequireEgressRestricted: %v", err)
	}
	sharing := testRegistry(fakeConfinement{sharesNetwork: true}, fakeProvisioner{})
	if _, err := Select(sharing, req, policy.SandboxPolicy{}); err == nil {
		t.Error("an isolator that cannot confine the network must refuse managed")
	}
}

func TestAvailableIsolations(t *testing.T) {
	reg := NewRegistry[isoshared.Isolator]().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Isolator { return fakeIsolator{available: true} }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return fakeIsolator{available: false} })
	avail := AvailableIsolations(reg)
	if len(avail) != 1 || avail[0] != isolation.IsolationNone {
		t.Errorf("AvailableIsolations = %v, want [none]", avail)
	}
}
