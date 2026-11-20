// Package runspec holds the portable plan of one agent run: what runs, under
// which confinement, and with which guarantees. It holds no checkout, terminal
// or overlay, and no part of the command line or the configuration files that
// produced it, so any executor can carry a Spec out.
package runspec

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/policy"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
)

// Spec is a run with every choice resolved and every input validated.
type Spec struct {
	Agent    *types.AgentConfig
	Provider *types.ModelProvider
	// IsolationName and ProvisionName are the chosen axes; the network axis is
	// Run.Network.
	IsolationName isolation.IsolationMethod
	ProvisionName provision.ProvisionMethod
	// Provision is the input the provisioner of ProvisionName contributes from.
	Provision provshared.Input
	// Allow is the egress allowlist of a managed network run, empty for any other.
	Allow egress.Allowlist
	// Credentials are the credential bindings an executor resolves at run time.
	Credentials []credshared.Binding
	// Capabilities are the guarantees the selected sandbox establishes, and Policy
	// the guarantees the run demands of it.
	Capabilities policy.Capabilities
	Policy       policy.SandboxPolicy
	// Run is the confinement input: the agent Process, the directories, the
	// binds, the image and the runtime. An executor fills in what only exists at
	// run time: the resolved provider environment and secrets, and the egress
	// route of a managed run.
	Run isoshared.RunConfig
}
