package webhook

import (
	"slices"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	launchiso "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	launchprov "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/launch"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/internal/validator"
	"github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/test/testutil"
)

// testCatalog registers the claude and opencode harnesses, the git and jj
// workspaces and the nix toolchain.
type testCatalog struct{}

func (testCatalog) HasHarness(agent agentv1alpha1.AgentType) bool {
	return agent == agentv1alpha1.AgentTypeClaude || agent == agentv1alpha1.AgentTypeOpencode
}

func (testCatalog) HasWorkspace(wsType agentv1alpha1.WorkspaceType) bool {
	return wsType == agentv1alpha1.WorkspaceTypeGit || wsType == agentv1alpha1.WorkspaceTypeJujutsu
}

func (testCatalog) HasToolchain(tcType agentv1alpha1.ToolchainType) bool {
	return tcType == agentv1alpha1.ToolchainTypeNix
}

// podDouble is a confinement with the guarantees of a hardened pod: a
// digest-pinned image is pinned, and a runtime class the policy approves is a
// kernel boundary.
type podDouble struct{ approved []string }

func (podDouble) PinnedProvision(_ provision.ProvisionMethod, _ bool, image string) bool {
	return validator.ValidatePinnedImageReference(image) == nil
}
func (podDouble) HidesHost(bool, string) bool { return true }
func (p podDouble) KernelIsolated(runtime string) bool {
	return strings.TrimSpace(runtime) != "" && slices.Contains(p.approved, runtime)
}
func (podDouble) ConfinesNetwork() bool                    { return true }
func (podDouble) EgressTransport() network.EgressTransport { return network.TransportNone }
func (podDouble) Bounded() bool                            { return true }
func (podDouble) RequiresImage() bool                      { return true }
func (podDouble) AcceptsHostBinds() bool                   { return false }

// imageDouble is a provisioner that leaves every tool to the image.
type imageDouble struct{}

func (imageDouble) Contribute(launchprov.Input) (provision.Contribution, error) {
	return provision.Contribution{}, nil
}
func (imageDouble) Pinned() bool { return false }
func (imageDouble) CheckAgent(provision.Contribution, *types.AgentConfig) (bool, error) {
	return false, nil
}

// testPlanner plans with a pod sandbox and the shipped harness facts.
type testPlanner struct{}

func (testPlanner) Sandbox(runtimeClasses []string) launch.Sandbox {
	confinement := podDouble{approved: slices.Clone(runtimeClasses)}
	registry := sandbox.NewRegistry[launchiso.Confinement]().
		RegisterIsolator("pod", func() launchiso.Confinement { return confinement }).
		RegisterProvisioner(provision.ProvisionNone, func() launchprov.Provisioner { return imageDouble{} })
	return launch.Sandbox{Registry: registry, Isolation: "pod", Provision: provision.ProvisionNone}
}

func (testPlanner) Harness(agent agentv1alpha1.AgentType) (*types.AgentConfig, error) {
	return agents.GetAgent(types.AgentType(agent))
}

func newRunWebhook(store *testutil.MemoryStore) *AgentRunWebhook {
	return NewAgentRunWebhook(store, store, testCatalog{}, testPlanner{}, testutil.NewPresets())
}

func runWebhookWith(objects ...any) *AgentRunWebhook {
	return newRunWebhook(testutil.NewMemoryStore(objects...))
}

func workspaceWebhookWith(objects ...any) *AgentWorkspaceWebhook {
	return NewAgentWorkspaceWebhook(testutil.NewMemoryStore(objects...), testCatalog{})
}
