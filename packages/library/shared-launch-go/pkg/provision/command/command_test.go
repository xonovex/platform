package command

import (
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

func TestCommand_ContributesInitCommandsUnpinned(t *testing.T) {
	p := New()
	if p.Pinned() {
		t.Error("command provisioner must report Pinned()=false")
	}
	c, err := p.Contribute(provshared.Input{InitCommands: []string{"echo hi"}})
	if err != nil {
		t.Fatalf("Contribute err = %v", err)
	}
	if len(c.InitCommands) != 1 || c.InitCommands[0] != "echo hi" {
		t.Errorf("command provisioner InitCommands = %v, want [echo hi]", c.InitCommands)
	}
}

func TestCommand_TrustsTheInitCommandsWithTheAgent(t *testing.T) {
	handled, err := New().CheckAgent(provision.Contribution{}, &types.AgentConfig{Binary: "claude"})
	if !handled || err != nil {
		t.Errorf("CheckAgent() = (%v, %v), want (true, nil)", handled, err)
	}
}
