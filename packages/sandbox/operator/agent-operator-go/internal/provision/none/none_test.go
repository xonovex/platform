package none

import (
	"reflect"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

func TestProvisionerContributesNothing(t *testing.T) {
	contribution, err := Provisioner{}.Contribute(provshared.Input{InitCommands: []string{"make"}})
	if err != nil || !reflect.DeepEqual(contribution, provision.Contribution{}) {
		t.Fatalf("Contribute() = %+v, %v, want an empty contribution", contribution, err)
	}
	if (Provisioner{}).Pinned() {
		t.Error("Pinned() = true, want false: the image carries the pin")
	}
	handled, err := Provisioner{}.CheckAgent(contribution, &types.AgentConfig{Type: types.AgentClaude})
	if handled || err != nil {
		t.Errorf("CheckAgent() = %v, %v, want the image to supply the agent", handled, err)
	}
}
