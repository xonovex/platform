package plugins

import (
	"testing"

	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

func TestDefaultRegistry(t *testing.T) {
	reg := DefaultRegistry()

	for _, m := range []types.IsolationMethod{types.IsolationNone, types.IsolationBwrap, types.IsolationDocker} {
		if _, err := reg.Isolator(m); err != nil {
			t.Errorf("isolator %q not registered: %v", m, err)
		}
	}
	for _, m := range []types.ProvisionMethod{types.ProvisionNone, types.ProvisionNix, types.ProvisionCommand} {
		if _, err := reg.Provisioner(m); err != nil {
			t.Errorf("provisioner %q not registered: %v", m, err)
		}
	}

	// The nix provisioner declares itself pinned; none does not.
	nixp, _ := reg.Provisioner(types.ProvisionNix)
	if !nixp.Pinned() {
		t.Error("nix provisioner must report Pinned()=true")
	}
	nonep, _ := reg.Provisioner(types.ProvisionNone)
	if nonep.Pinned() {
		t.Error("none provisioner must report Pinned()=false")
	}

	// A real cell selects, and the pinned guarantee is satisfied by the nix plugin's
	// own capability — the policy engine never names nix.
	if _, _, err := sandbox.Select(reg,
		sandbox.Request{Isolation: types.IsolationBwrap, Provision: types.ProvisionNix, Network: types.NetworkNone},
		types.SandboxPolicy{RequirePinnedProvision: true}); err != nil {
		t.Errorf("Select(bwrap×nix, require-pinned) = %v, want nil", err)
	}
}
