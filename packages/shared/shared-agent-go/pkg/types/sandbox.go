package types

// SandboxMethod represents the sandbox execution method. It fuses isolation,
// provisioning, and network into one enum; the three-axis model below
// (IsolationMethod × ProvisioningMethod × NetworkMethod) supersedes it. Kept
// during the transition while callers migrate.
type SandboxMethod string

const (
	SandboxNone     SandboxMethod = "none"
	SandboxBwrap    SandboxMethod = "bwrap"
	SandboxDocker   SandboxMethod = "docker"
	SandboxCompose  SandboxMethod = "compose"
	SandboxNix      SandboxMethod = "nix"
	SandboxNixFlake SandboxMethod = "nixflake"
)

// IsolationMethod is the process-isolation axis: how the agent process is
// confined from the host. It does not determine where tools come from
// (ProvisioningMethod) or whether egress is constrained (NetworkMethod).
type IsolationMethod string

const (
	// IsolationNone runs the agent directly on the host, no namespace boundary.
	IsolationNone IsolationMethod = "none"
	// IsolationBwrap confines the agent with bubblewrap namespaces. Attack-surface
	// reduction, not a kernel trust boundary.
	IsolationBwrap IsolationMethod = "bwrap"
	// IsolationDocker confines the agent in a container. Default runc is
	// attack-surface reduction; a sandboxed runtime (runsc/gVisor) is a kernel
	// boundary (see KernelIsolated).
	IsolationDocker IsolationMethod = "docker"
)

// ProvisioningMethod is the tool-provisioning axis: how the agent's tools reach
// its PATH, independent of how the process is isolated.
type ProvisioningMethod string

const (
	// ProvisionNone provides no tools; the agent relies on host/base-image PATH
	// (subject to HostPassthrough).
	ProvisionNone ProvisioningMethod = "none"
	// ProvisionNix resolves a flake.lock/rev-pinned closure on the host and mounts
	// it read-only into the sandbox. The only pinned provisioning (see
	// ProvisioningIsPinned).
	ProvisionNix ProvisioningMethod = "nix"
	// ProvisionCommand runs a single init-command list before the agent.
	ProvisionCommand ProvisioningMethod = "command"
)

// NetworkMethod is the network-egress axis, replacing the old boolean Network.
//
//	host  = share host net, unrestricted egress — today's de-facto behavior, now
//	        an EXPLICIT opt-in; does NOT satisfy RequireEgressRestricted.
//	none  = no network (bwrap --unshare-net / docker --network none); satisfies
//	        RequireEgressRestricted.
//	proxy = egress ONLY via a host-side allowlist HTTP(S) proxy; link-local,
//	        metadata (169.254.169.254), RFC1918, and loopback denied; satisfies
//	        RequireEgressRestricted. Recommended default for untrusted code that
//	        still needs the model API.
type NetworkMethod string

const (
	NetworkHost  NetworkMethod = "host"
	NetworkNone  NetworkMethod = "none"
	NetworkProxy NetworkMethod = "proxy"
)

// Contribution is what a Provisioner hands an Isolator; the Isolator applies it
// via its own mechanism (bwrap binds / docker -v). Pure data, no host calls.
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

// DefaultEgressAllowlist seeds NetworkProxy: provider API endpoints plus common
// package registries and git forges. EgressAllowlist (CLI `--egress-allow`,
// repeatable) EXTENDS, not replaces, this set.
var DefaultEgressAllowlist = []string{
	// Provider API endpoints
	"api.anthropic.com",
	"api.z.ai",
	// Package registries
	"registry.npmjs.org",
	"pypi.org",
	"files.pythonhosted.org",
	"crates.io",
	"static.crates.io",
	"proxy.golang.org",
	"sum.golang.org",
	// Git forges
	"github.com",
	"codeload.github.com",
	"raw.githubusercontent.com",
	"objects.githubusercontent.com",
	"gitlab.com",
}

// SandboxPolicy expresses the isolation guarantees the caller demands of the
// selected sandbox, independent of which axes were requested. Each guarantee is
// independently requestable and is enforced fail-closed (see EnforcePolicy):
// the engine refuses to run when a requested guarantee cannot be established,
// never silently degrades.
type SandboxPolicy struct {
	// RequirePinnedProvisioning mandates provisioning from a pinned source
	// (nix closure or a pinned image), enforced at resolve via
	// --frozen / --no-write-lock-file against a committed lock (fail closed).
	RequirePinnedProvisioning bool
	// RequireHostToolsUnreachable mandates that host tools are off PATH AND not
	// bind-reachable. Conditioned on closure-only store binds, no host-$HOME bind,
	// and (docker) a pinned image.
	RequireHostToolsUnreachable bool
	// RequireEgressRestricted mandates Network ∈ {none, proxy}; Network=host does
	// not satisfy it.
	RequireEgressRestricted bool
	// RequireKernelIsolation mandates a kernel boundary: docker --runtime
	// runsc/gVisor, or a pod with a sandboxed runtimeClass (gVisor/Kata/kata-cc).
	// NOT satisfied by bwrap or default runc.
	RequireKernelIsolation bool
}

// SandboxConfig holds sandbox configuration
type SandboxConfig struct {
	AgentID             string
	Method              SandboxMethod
	Isolation           IsolationMethod
	Provisioning        ProvisioningMethod
	HostPassthrough     bool
	Policy              SandboxPolicy
	Agent               *AgentConfig
	HomeDir             string
	Image               string
	ComposeFile         string
	Service             string
	WorkDir             string
	RepoDir             string
	Network             NetworkMethod
	EgressAllowlist     []string
	BindPaths           []string
	RoBindPaths         []string
	CustomEnv           []string
	Provider            *ModelProvider
	AgentArgs           []string
	SandboxInitCommands []string
	Verbose             bool
	Debug               bool
	DryRun              bool
}

// SandboxExecutor defines the interface for sandbox implementations
type SandboxExecutor interface {
	IsAvailable() (bool, error)
	Execute(config *SandboxConfig) (int, error)
	GetCommand(config *SandboxConfig) []string
}
