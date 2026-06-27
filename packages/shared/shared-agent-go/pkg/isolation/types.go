package isolation

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
const DefaultContainerImage = "node:trixie-slim"

// UserConfigPaths lists home-relative paths that should be bind-mounted
// into sandboxed environments so agents can access user configuration.
var UserConfigPaths = []string{
	".claude",
	".claude.json",
	".gitconfig",
	".gitignore_global",
	".ssh",
	".config",
	".npmrc",
	".npm",
	".npm-global",
	".cargo",
	".rustup",
	".local",
	".cache",
}
