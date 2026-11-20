package plugins

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	netshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system/systemtest"
)

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry(systemtest.New(), testLog)

	for _, m := range []isolation.IsolationMethod{isolation.IsolationNone, isolation.IsolationBwrap, isolation.IsolationDocker} {
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

	// Only isolation=none shares the host network.
	for m, confines := range map[isolation.IsolationMethod]bool{isolation.IsolationNone: false, isolation.IsolationBwrap: true, isolation.IsolationDocker: true} {
		iso, _ := reg.Isolator(m)
		if iso.ConfinesNetwork() != confines {
			t.Errorf("isolator %q ConfinesNetwork() = %v, want %v", m, !confines, confines)
		}
	}

	// A real cell selects, and the pinned guarantee is satisfied by the nix plugin's
	// own capability — the policy engine never names nix.
	if _, _, err := sandbox.Select(reg,
		sandbox.Request{Isolation: isolation.IsolationBwrap, Provision: provision.ProvisionNix, Network: netshared.ModeNone},
		policy.SandboxPolicy{RequirePinnedProvision: true}); err != nil {
		t.Errorf("Select(bwrap×nix, require-pinned) = %v, want nil", err)
	}
}
