package nix

import "fmt"

// Pin represents a nixpkgs channel pin
type Pin struct {
	Name string
	Ref  string
}

// Pins maps friendly names to their channel refs
var Pins = map[string]Pin{
	"nixos-24.11":      {Name: "nixos-24.11", Ref: "nixos-24.11"},
	"nixos-unstable":   {Name: "nixos-unstable", Ref: "nixos-unstable"},
	"nixpkgs-unstable": {Name: "nixpkgs-unstable", Ref: "nixpkgs-unstable"},
}

// DefaultPin is the default nixpkgs pin
const DefaultPin = "nixos-unstable"

// DefaultPackages are the default packages for agent environments
var DefaultPackages = []string{
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

// ExpandPackageSets expands any named package sets in the input list,
// returning a deduplicated slice of individual package names.
// Unknown names are preserved as individual package names.
func ExpandPackageSets(packages []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, pkg := range packages {
		if set, ok := PackageSets[pkg]; ok {
			for _, p := range set {
				if !seen[p] {
					seen[p] = true
					result = append(result, p)
				}
			}
		} else {
			if !seen[pkg] {
				seen[pkg] = true
				result = append(result, pkg)
			}
		}
	}
	return result
}

// ValidatePin returns an error if the pin name is not recognised
func ValidatePin(pin string) error {
	if pin == "" {
		return nil
	}
	if _, ok := Pins[pin]; !ok {
		return fmt.Errorf("unknown nixpkgs pin %q; known pins: nixos-24.11, nixos-unstable, nixpkgs-unstable", pin)
	}
	return nil
}
