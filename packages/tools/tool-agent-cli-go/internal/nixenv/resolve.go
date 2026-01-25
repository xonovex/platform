package nixenv

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Package name validation regex
var packageNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_+\-.]+$`)

// GetAgentNixDir returns the base directory for agent-nix data
func GetAgentNixDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "agent-nix")
}

// GetSpecsDir returns the directory for storing generated .nix spec files
func GetSpecsDir() string {
	return filepath.Join(GetAgentNixDir(), "specs")
}

// GetEnvsDir returns the directory for storing nix-build output symlinks
func GetEnvsDir() string {
	return filepath.Join(GetAgentNixDir(), "envs")
}

// GetAgentsDir returns the directory for per-agent runtime data
func GetAgentsDir() string {
	return filepath.Join(GetAgentNixDir(), "agents")
}

// ValidatePackageName validates a single package name
func ValidatePackageName(name string) bool {
	return packageNameRegex.MatchString(name)
}

// ValidateEnvSpec validates an EnvSpec
func ValidateEnvSpec(spec *EnvSpec) error {
	if spec.NixpkgsPin == "" {
		return fmt.Errorf("nixpkgs_pin is required")
	}

	if _, ok := NixpkgsPins[spec.NixpkgsPin]; !ok {
		allowed := make([]string, 0, len(NixpkgsPins))
		for k := range NixpkgsPins {
			allowed = append(allowed, k)
		}
		return fmt.Errorf("invalid nixpkgs_pin %q. Allowed: %s", spec.NixpkgsPin, strings.Join(allowed, ", "))
	}

	if len(spec.Packages) == 0 {
		return fmt.Errorf("packages must be a non-empty array")
	}

	for _, pkg := range spec.Packages {
		if !ValidatePackageName(pkg) {
			return fmt.Errorf("invalid package name %q", pkg)
		}
	}

	return nil
}

// NormalizeEnvSpec normalizes an EnvSpec by sorting and deduplicating packages
func NormalizeEnvSpec(spec *EnvSpec) *EnvSpec {
	// Deduplicate
	seen := make(map[string]bool)
	packages := make([]string, 0, len(spec.Packages))
	for _, pkg := range spec.Packages {
		if !seen[pkg] {
			seen[pkg] = true
			packages = append(packages, pkg)
		}
	}

	// Sort
	sort.Strings(packages)

	pin := spec.NixpkgsPin
	if pin == "" {
		pin = DefaultNixpkgsPin
	}

	return &EnvSpec{
		NixpkgsPin: pin,
		Packages:   packages,
	}
}

// ComputeEnvID computes the environment ID from a normalized EnvSpec
func ComputeEnvID(spec *EnvSpec) string {
	content := spec.NixpkgsPin + "\n" + strings.Join(spec.Packages, "\n")
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16]
}

// isValidOutLink checks if an outLink symlink points to a valid store path
func isValidOutLink(outLink string) bool {
	info, err := os.Lstat(outLink)
	if err != nil {
		return false
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return false
	}

	target, err := filepath.EvalSymlinks(outLink)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(target, "/nix/store/") {
		return false
	}

	if _, err := os.Stat(target); err != nil {
		return false
	}

	return true
}

// ResolveEnv resolves an EnvSpec to paths and checks cache status
func ResolveEnv(spec *EnvSpec) (*ResolvedEnv, error) {
	if err := ValidateEnvSpec(spec); err != nil {
		return nil, err
	}

	normalized := NormalizeEnvSpec(spec)
	envID := ComputeEnvID(normalized)

	specPath := filepath.Join(GetSpecsDir(), envID+".nix")
	outLink := filepath.Join(GetEnvsDir(), envID)
	ready := isValidOutLink(outLink)

	return &ResolvedEnv{
		EnvID:    envID,
		SpecPath: specPath,
		OutLink:  outLink,
		Ready:    ready,
	}, nil
}

// ExpandPackageSets expands set names to package lists
func ExpandPackageSets(sets []string) []string {
	packages := make([]string, 0)
	for _, set := range sets {
		set = strings.TrimSpace(set)
		if setPackages, ok := PackageSets[set]; ok {
			packages = append(packages, setPackages...)
		}
	}
	return packages
}
