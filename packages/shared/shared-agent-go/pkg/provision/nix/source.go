package nix

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// NixSourceKind identifies how a nix environment is sourced before it is
// resolved to a content-pinned closure.
type NixSourceKind string

const (
	// NixSourcePackages resolves a rev-pinned nixpkgs package set.
	NixSourcePackages NixSourceKind = "packages"
	// NixSourceProjectFlake resolves a devShell from the project's own flake.
	NixSourceProjectFlake NixSourceKind = "project-flake"
)

// NixSource describes the input a realizer resolves to a ClosureDescriptor.
//
// NixSourcePackages uses Rev (pinned nixpkgs rev) + Packages; NixSourceProjectFlake
// uses FlakeRef (the project root) + Shell (the devShell attribute).
type NixSource struct {
	Kind     NixSourceKind
	Rev      string   // pinned nixpkgs rev (packages source)
	Packages []string // package set (packages source)
	FlakeRef string   // <projectRoot> (project-flake source)
	Shell    string   // devShell name (project-flake source)
}

// ClosureDescriptor is the resolved, mount-ready result a realizer produces.
//
// Requisites carries the transitive closure (the host-resolve in the CLI nix
// provisioner populates it via `nix path-info -r`). Isolators bind ONLY these
// store paths read-only, never the whole world-readable /nix/store — that bound
// set is what makes RequireHostToolsUnreachable accurate.
type ClosureDescriptor struct {
	StorePaths  []string          // top-level realized paths (devShell/agent outputs)
	Requisites  []string          // transitive closure store paths (`nix path-info -r`)
	PathEntries []string          // PATH dirs to prepend (the pinned tools)
	Env         map[string]string // extra env (devShell vars)
}

// packageNamePattern bounds a nix package attribute name.
var packageNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_+\-.]+$`)

// ValidatePackageName reports whether name is a syntactically valid package
// attribute path.
func ValidatePackageName(name string) bool {
	return packageNamePattern.MatchString(name)
}

// ValidateSource returns an error if the source is not structurally resolvable.
// It is pure: it never calls host nix. The committed-lock / concrete-rev pin is
// enforced at resolve time by the host-resolve provisioner.
func ValidateSource(s NixSource) error {
	switch s.Kind {
	case NixSourcePackages:
		if len(s.Packages) == 0 {
			return fmt.Errorf("nix source %q requires at least one package", s.Kind)
		}
		for _, pkg := range ExpandPackageSets(s.Packages) {
			if !ValidatePackageName(pkg) {
				return fmt.Errorf("invalid package name %q", pkg)
			}
		}
		return nil
	case NixSourceProjectFlake:
		if s.FlakeRef == "" {
			return fmt.Errorf("nix source %q requires a flake reference", s.Kind)
		}
		return nil
	default:
		return fmt.Errorf("unknown nix source kind %q", s.Kind)
	}
}

// ComputeEnvID derives a stable 16-character content hash of a source. It is the
// cache / GC-root key the host-resolve provisioner reuses, so it must be
// independent of package argument order.
func ComputeEnvID(s NixSource) string {
	var b strings.Builder
	b.WriteString(string(s.Kind))
	switch s.Kind {
	case NixSourcePackages:
		b.WriteByte('\n')
		b.WriteString(s.Rev)
		for _, pkg := range normalizePackages(s.Packages) {
			b.WriteByte('\n')
			b.WriteString(pkg)
		}
	case NixSourceProjectFlake:
		b.WriteByte('\n')
		b.WriteString(s.FlakeRef)
		b.WriteByte('\n')
		b.WriteString(s.Shell)
	}
	hash := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(hash[:])[:16]
}

// normalizePackages expands sets, dedupes, and sorts so ComputeEnvID is stable
// regardless of the order packages were requested in.
func normalizePackages(packages []string) []string {
	expanded := ExpandPackageSets(packages)
	seen := make(map[string]bool, len(expanded))
	out := make([]string, 0, len(expanded))
	for _, pkg := range expanded {
		if !seen[pkg] {
			seen[pkg] = true
			out = append(out, pkg)
		}
	}
	sort.Strings(out)
	return out
}
