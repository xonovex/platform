package config

// Layer names one source of configuration. A layer is a file or the command line;
// the merge rule of a key decides how the layers combine.
type Layer string

const (
	// LayerManaged is the system-wide file an administrator controls. Its values
	// cannot be relaxed by any other layer.
	LayerManaged Layer = "managed"
	// LayerUser is the file of the user: the default configuration file, or the
	// one --config names.
	LayerUser Layer = "user"
	// LayerProject is the file that sits in the project the agent works on. A
	// project can be hostile, so it may set only keys that cannot widen access,
	// and the provider content it supplies takes effect only after the user
	// approves its digest.
	LayerProject Layer = "project"
	// LayerFlags is the command line of this run.
	LayerFlags Layer = "flags"
)

// Key names one configuration setting.
type Key string

// The configuration keys. A key is the name of the matching setting in every
// configuration file.
const (
	KeyAgent     Key = "agent"
	KeyProvider  Key = "provider"
	KeyModels    Key = "models"
	KeyIsolation Key = "isolation"
	KeyProvision Key = "provision"
	KeyNetwork   Key = "network"
	KeyHomeDir   Key = "homeDir"

	KeyBindPaths   Key = "bindPaths"
	KeyRoBindPaths Key = "roBindPaths"
	KeyCustomEnv   Key = "customEnv"
	KeyAllowHosts  Key = "allowHosts"
	KeyCredentials Key = "credentials"

	KeyRequirePinnedProvision      Key = "requirePinnedProvision"
	KeyRequireHostToolsUnreachable Key = "requireHostToolsUnreachable"
	KeyRequireEgressRestricted     Key = "requireEgressRestricted"
	KeyRequireKernelIsolation      Key = "requireKernelIsolation"
)

// Rule is how the layers that set one key combine.
type Rule string

const (
	// RuleReplace gives the key one value. The managed layer wins, then the
	// command line, then the project, then the user. A command line value that
	// differs from a managed one is an error, not a silent override.
	RuleReplace Rule = "replace"
	// RuleAppend gives the key the values of every layer, the managed ones last so
	// that a later entry of the same name still comes from the administrator.
	RuleAppend Rule = "append"
	// RuleRequire gives the key true when any layer sets it. A guarantee that one
	// layer demands stays demanded.
	RuleRequire Rule = "require"
)

// Spec declares one key as data: the rule that merges it, the layers that may set
// it, and how a configuration file holds it.
type Spec struct {
	Key  Key
	Rule Rule
	// Layers are the layers that may set the key. The command line may always set it.
	Layers []Layer
	Doc    string
}

var (
	// everyFile lets the managed, user and project files set a key that only
	// narrows what a run may do or that selects among fixed choices.
	everyFile = []Layer{LayerManaged, LayerUser, LayerProject}
	// trustedFiles keeps a project file from setting a key that widens access, such
	// as a bind mount, a credential, a host the run may reach, an environment
	// variable, the home directory or the confinement itself.
	trustedFiles = []Layer{LayerManaged, LayerUser}
)

// Specs lists every key in display order.
var Specs = []Spec{
	{KeyAgent, RuleReplace, everyFile, "Agent to run."},
	{KeyProvider, RuleReplace, everyFile, "Model provider to use, by name."},
	{KeyModels, RuleAppend, everyFile, `Model slot overrides of the provider, as "slot=id". A later layer wins per slot, the managed layer last. A project value takes effect only after the user approves its digest.`},
	{KeyIsolation, RuleReplace, trustedFiles, "Isolation axis."},
	{KeyProvision, RuleReplace, trustedFiles, "Provision axis."},
	{KeyNetwork, RuleReplace, trustedFiles, "Network egress axis (host, none, managed)."},
	{KeyHomeDir, RuleReplace, trustedFiles, "Home directory of the sandbox."},
	{KeyBindPaths, RuleAppend, trustedFiles, "Read-write bind mounts."},
	{KeyRoBindPaths, RuleAppend, trustedFiles, "Read-only bind mounts."},
	{KeyCustomEnv, RuleAppend, trustedFiles, "Environment variables, as KEY=VALUE."},
	{KeyAllowHosts, RuleAppend, trustedFiles, `Hosts a managed network run may reach, as "host", "host:port" or "*.domain[:port]".`},
	{KeyCredentials, RuleAppend, trustedFiles, `Credential bindings, as "NAME=scheme:locator". A file holds where a credential is, never its value.`},
	{KeyRequirePinnedProvision, RuleRequire, everyFile, "Require provisioning from a pinned source."},
	{KeyRequireHostToolsUnreachable, RuleRequire, everyFile, "Require host tools to be unreachable from the sandbox."},
	{KeyRequireEgressRestricted, RuleRequire, everyFile, "Require network egress to be disabled or enforced by a supported transport."},
	{KeyRequireKernelIsolation, RuleRequire, everyFile, "Require a kernel-isolating runtime."},
}

// specOf returns the spec of key.
func specOf(key Key) (Spec, bool) {
	for _, spec := range Specs {
		if spec.Key == key {
			return spec, true
		}
	}
	return Spec{}, false
}

// allows reports whether layer may set the spec's key.
func (s Spec) allows(layer Layer) bool {
	if layer == LayerFlags {
		return true
	}
	for _, allowed := range s.Layers {
		if allowed == layer {
			return true
		}
	}
	return false
}
