package nix

import "fmt"

// DefaultPin is the default nixpkgs pin
const DefaultPin = "nixos-unstable"

// ExpandPackageSets expands any named package sets in the input list,
// returning a deduplicated slice of individual package names.
// Unknown names are preserved as individual package names.
func ExpandPackageSets(packages []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, pkg := range packages {
		if set, ok := packageSet(pkg); ok {
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

func packageSet(name string) ([]string, bool) {
	switch name {
	case "nodejs":
		return []string{"nodejs_24", "python312", "gnumake", "gcc", "gnused", "gawk", "binutils"}, true
	case "python":
		return []string{"python312", "python312Packages.pip"}, true
	case "go":
		return []string{"go"}, true
	case "rust":
		return []string{"rustc", "cargo"}, true
	case "kubernetes":
		return []string{"kubectl", "kubernetes-helm", "k9s"}, true
	case "terraform":
		return []string{"terraform", "terragrunt"}, true
	case "docker":
		return []string{"docker-client"}, true
	case "aws":
		return []string{"awscli2"}, true
	case "gcp":
		return []string{"google-cloud-sdk"}, true
	default:
		return nil, false
	}
}

// ValidatePin returns an error if the pin name is not recognised
func ValidatePin(pin string) error {
	if pin == "" {
		return nil
	}
	if pin != "nixos-24.11" && pin != "nixos-unstable" && pin != "nixpkgs-unstable" {
		return fmt.Errorf("unknown nixpkgs pin %q; known pins: nixos-24.11, nixos-unstable, nixpkgs-unstable", pin)
	}
	return nil
}
