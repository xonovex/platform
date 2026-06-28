package command

import (
	"testing"

	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
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
