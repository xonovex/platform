package isolation

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"

// IsolationMethod is the process-isolation axis: how the agent process is
// confined from the host. It does not determine where tools come from
// (provision.ProvisionMethod) or whether egress is constrained
// (network.NetworkMethod). Isolators are resolved by an injected registry, so
// this set is open for extension.
type IsolationMethod string

const (
	// IsolationNone runs the agent directly on the host, no namespace boundary.
	IsolationNone IsolationMethod = "none"
	// IsolationBwrap confines the agent with bubblewrap namespaces. Attack-surface
	// reduction, not a kernel trust boundary.
	IsolationBwrap IsolationMethod = "bwrap"
	// IsolationDocker confines the agent in a container. Default runc is
	// attack-surface reduction; a sandboxed runtime (runsc/gVisor) is a kernel
	// boundary (each isolator declares its own KernelIsolated capability).
	IsolationDocker IsolationMethod = "docker"
)

// DefaultContainerImage is the default container image for running agents.
const DefaultContainerImage = "docker.io/library/node:26.3.0-trixie-slim@sha256:95a34da32a840bd9b3b09a5b773591c16923e350174b1c50e1200c75bf15eaa9"

// UserConfigPaths returns the smallest home-relative configuration set needed
// by the selected agent. General credential stores and broad configuration
// directories require an explicit read-only bind from the caller.
func UserConfigPaths(agentType types.AgentType) []string {
	switch agentType {
	case types.AgentClaude:
		return []string{".claude", ".claude.json"}
	case types.AgentOpencode:
		return []string{".config/opencode"}
	default:
		return nil
	}
}
