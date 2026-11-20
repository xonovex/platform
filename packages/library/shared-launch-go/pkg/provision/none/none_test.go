package none

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

func TestNone_ContributesNothingUnpinned(t *testing.T) {
	p := New()
	if p.Pinned() {
		t.Error("none provisioner must report Pinned()=false")
	}
	c, err := p.Contribute(provshared.Input{InitCommands: []string{"ignored"}})
	if err != nil {
		t.Fatalf("Contribute err = %v", err)
	}
	if len(c.InitCommands) != 0 || len(c.RoBindPaths) != 0 || len(c.PathEntries) != 0 || len(c.Env) != 0 {
		t.Errorf("none provisioner must contribute nothing, got %+v", c)
	}
}

func TestNone_LeavesTheAgentToTheIsolator(t *testing.T) {
	handled, err := New().CheckAgent(provision.Contribution{}, &types.AgentConfig{Binary: "claude"})
	if handled || err != nil {
		t.Errorf("CheckAgent() = (%v, %v), want (false, nil)", handled, err)
	}
}
