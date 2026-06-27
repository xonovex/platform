package provision

// ProvisionMethod is the tool-provisioning axis: how the agent's tools reach its
// PATH, independent of how the process is isolated. Provisioners are resolved by
// an injected registry and each declares its own guarantees (e.g. Pinned), so
// this set is open for extension.
type ProvisionMethod string

const (
	// ProvisionNone provides no tools; the agent relies on host/base-image PATH
	// (subject to host passthrough).
	ProvisionNone ProvisionMethod = "none"
	// ProvisionNix resolves a flake.lock/rev-pinned closure on the host and mounts
	// it read-only into the sandbox. Pinned provisioning (declares Pinned()=true).
	ProvisionNix ProvisionMethod = "nix"
	// ProvisionCommand runs a single init-command list before the agent.
	ProvisionCommand ProvisionMethod = "command"
)

// Contribution is the single neutral data-coupling handoff across the shared
// module boundary: what a Provisioner hands an Isolator, which the Isolator
// applies via its own mechanism (bwrap binds / docker -v). Pure data, no host
// calls, no method-specific fields.
type Contribution struct {
	// RoBindPaths are host paths to mount read-only (e.g. a resolved closure's
	// requisites).
	RoBindPaths []string
	// PathEntries are PATH directories to PREPEND (the pinned tools).
	PathEntries []string
	// Env is extra environment to set (devShell vars).
	Env map[string]string
	// InitCommands run once at init before the agent (the `command` provisioner).
	InitCommands []string
}
