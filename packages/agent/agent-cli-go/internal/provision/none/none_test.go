package none

import (
	"testing"

	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
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
