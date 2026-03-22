package nixenv

import "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/nix"

// NixpkgsPin is an alias for the shared Pin type
type NixpkgsPin = nix.Pin

// Shared definitions re-exported for backward compatibility within this package
var NixpkgsPins = nix.Pins

const DefaultNixpkgsPin = nix.DefaultPin

var DefaultBasePackages = nix.DefaultPackages
var PackageSets = nix.PackageSets

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
