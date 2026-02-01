package nixenv

// NixpkgsPin represents a nixpkgs channel pin configuration
type NixpkgsPin struct {
	Name string
	Ref  string
}

// Allowed nixpkgs pins - maps friendly names to tarball refs
var NixpkgsPins = map[string]NixpkgsPin{
	"nixos-24.11": {
		Name: "nixos-24.11",
		Ref:  "nixos-24.11",
	},
	"nixos-unstable": {
		Name: "nixos-unstable",
		Ref:  "nixos-unstable",
	},
	"nixpkgs-unstable": {
		Name: "nixpkgs-unstable",
		Ref:  "nixpkgs-unstable",
	},
}

// DefaultNixpkgsPin is the default nixpkgs pin to use
const DefaultNixpkgsPin = "nixos-unstable"

// DefaultBasePackages are the default packages for agent environments
var DefaultBasePackages = []string{
	"nodejs_24",
	"git",
	"ripgrep",
	"fd",
	"fzf",
	"jq",
	"curl",
	"coreutils",
	"bash",
}

// PackageSets are predefined collections of packages for common use cases
var PackageSets = map[string][]string{
	"nodejs": {
		"nodejs_24",
		"python312",
		"gnumake",
		"gcc",
		"gnused",
		"gawk",
		"binutils",
	},
	"python":     {"python312", "python312Packages.pip"},
	"go":         {"go"},
	"rust":       {"rustc", "cargo"},
	"kubernetes": {"kubectl", "kubernetes-helm", "k9s"},
	"terraform":  {"terraform", "terragrunt"},
	"docker":     {"docker-client"},
	"aws":        {"awscli2"},
	"gcp":        {"google-cloud-sdk"},
}

// EnvSpec is the environment specification for building a Nix environment
type EnvSpec struct {
	NixpkgsPin string
	Packages   []string
}

// ResolvedEnv is the result of resolving an environment specification
type ResolvedEnv struct {
	EnvID    string
	SpecPath string
	OutLink  string
	Ready    bool
}

// BuildResult is the result of building a Nix environment
type BuildResult struct {
	Success   bool
	StorePath string
	Error     string
	Duration  int64 // milliseconds
}

// NixSandboxConfig is the extended sandbox config for Nix
type NixSandboxConfig struct {
	Packages   []string
	NixpkgsPin string
	NoDefaults bool
}
