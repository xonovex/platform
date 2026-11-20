package webhook

import (
	agentv1alpha1 "github.com/xonovex/platform/packages/sandbox/operator/agent-operator-go/api/v1alpha1"
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

func newRunWebhook(store *testutil.MemoryStore) *AgentRunWebhook {
	return NewAgentRunWebhook(store, store, testCatalog{})
}

func runWebhookWith(objects ...any) *AgentRunWebhook {
	return newRunWebhook(testutil.NewMemoryStore(objects...))
}

func workspaceWebhookWith(objects ...any) *AgentWorkspaceWebhook {
	return NewAgentWorkspaceWebhook(testutil.NewMemoryStore(objects...), testCatalog{})
}
