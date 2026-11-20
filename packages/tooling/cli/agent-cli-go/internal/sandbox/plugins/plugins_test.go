package plugins

import (
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry(systemtest.New(), testLog)

	for _, m := range []isolation.IsolationMethod{
		isolation.IsolationNone, isolation.IsolationBwrap, isolation.IsolationDocker,
		isolation.IsolationSeatbelt, isolation.IsolationDockerSandboxes, isolation.IsolationOpenShell,
	} {
		if _, err := reg.Isolator(m); err != nil {
			t.Errorf("isolator %q not registered: %v", m, err)
		}
	}
	for _, m := range []provision.ProvisionMethod{provision.ProvisionNone, provision.ProvisionNix, provision.ProvisionCommand} {
		if _, err := reg.Provisioner(m); err != nil {
			t.Errorf("provisioner %q not registered: %v", m, err)
		}
	}

	// The nix provisioner declares itself pinned; none does not.
	nixp, _ := reg.Provisioner(provision.ProvisionNix)
	if !nixp.Pinned() {
		t.Error("nix provisioner must report Pinned()=true")
	}
	nonep, _ := reg.Provisioner(provision.ProvisionNone)
	if nonep.Pinned() {
		t.Error("none provisioner must report Pinned()=false")
	}

	// Each isolator declares whether it confines the network and how a managed run
	// reaches the allowlist; the registry never names one.
	capabilities := map[isolation.IsolationMethod]struct {
		confines  bool
		transport network.EgressTransport
	}{
		isolation.IsolationNone:            {false, network.TransportNone},
		isolation.IsolationBwrap:           {true, network.TransportUnixSocket},
		isolation.IsolationDocker:          {true, network.TransportUnixSocket},
		isolation.IsolationSeatbelt:        {true, network.TransportLoopback},
		isolation.IsolationDockerSandboxes: {false, network.TransportNone},
		isolation.IsolationOpenShell:       {true, network.TransportPolicy},
	}
	for m, want := range capabilities {
		iso, _ := reg.Isolator(m)
		if iso.ConfinesNetwork() != want.confines || iso.EgressTransport() != want.transport {
			t.Errorf("isolator %q confines %v over %q, want %v over %q", m, iso.ConfinesNetwork(), iso.EgressTransport(), want.confines, want.transport)
		}
	}
	for _, m := range []isolation.IsolationMethod{isolation.IsolationNone, isolation.IsolationDockerSandboxes} {
		_, err := sandbox.Select(reg,
			sandbox.Request{Isolation: m, Provision: provision.ProvisionNone, Network: network.NetworkManaged},
			policy.SandboxPolicy{RequireEgressRestricted: true})
		if err == nil {
			t.Errorf("Select(%s, managed, require-egress-restricted) must fail closed", m)
		}
	}

	// A real cell selects, and the pinned guarantee is satisfied by the nix plugin's
	// own capability — the policy engine never names nix.
	if _, err := sandbox.Select(reg,
		sandbox.Request{Isolation: isolation.IsolationBwrap, Provision: provision.ProvisionNix, Network: network.NetworkNone},
		policy.SandboxPolicy{RequirePinnedProvision: true}); err != nil {
		t.Errorf("Select(bwrap×nix, require-pinned) = %v, want nil", err)
	}
}

func TestDefaultCredentials(t *testing.T) {
	host := systemtest.New()
	host.Env["SOURCE"] = "value"
	registry := DefaultCredentials(host)
	if got := strings.Join(registry.Schemes(), ","); got != "env,file,keychain,pass,secret-service" {
		t.Errorf("schemes = %q", got)
	}
	secrets, err := credshared.Resolve(registry, []credshared.Binding{{Name: "TOKEN", Scheme: "env", Locator: "SOURCE"}})
	if err != nil || secrets["TOKEN"].Reveal() != "value" {
		t.Errorf("Resolve = %v, %v", secrets, err)
	}
	if store, err := registry.Store("file"); err != nil || !store.Available() {
		t.Errorf("file store: %v, %v", store, err)
	}
	for _, scheme := range []string{"keychain", "secret-service", "pass"} {
		store, err := registry.Store(scheme)
		if err != nil || store.Available() {
			t.Errorf("%s: store %v, err %v; the tool is not installed here", scheme, store, err)
		}
	}
}
