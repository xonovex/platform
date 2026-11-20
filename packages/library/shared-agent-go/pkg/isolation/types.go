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
	// IsolationSeatbelt confines the agent with a macOS Seatbelt sandbox profile.
	// Attack-surface reduction, not a kernel trust boundary.
	IsolationSeatbelt IsolationMethod = "seatbelt"
	// IsolationDockerSandboxes runs the agent in a Docker Sandboxes sandbox, a
	// virtual machine with its own agent image.
	IsolationDockerSandboxes IsolationMethod = "docker-sandboxes"
	// IsolationOpenShell runs the agent in an OpenShell sandbox on a gateway,
	// under a policy document the CLI hands the gateway.
	IsolationOpenShell IsolationMethod = "openshell"
)

// DefaultContainerImage is the default container image for running agents.
const DefaultContainerImage = "docker.io/library/node:26.3.0-trixie-slim@sha256:95a34da32a840bd9b3b09a5b773591c16923e350174b1c50e1200c75bf15eaa9"
