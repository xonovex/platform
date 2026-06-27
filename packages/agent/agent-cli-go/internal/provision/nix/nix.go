// Package nix is the provision=nix leaf: it resolves a NixSource to a
// content-pinned closure on the host, GC-roots it, and contributes read-only
// binds of ONLY the closure's requisites (never the whole /nix/store, never the
// nix daemon socket). The shared-module nix engine is imported under the
// sharednix alias since this leaf package is itself named nix.
package nix

import (
	"fmt"
	"os"
	"path/filepath"

	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision/nix"
)

// defaultFlakeShell is the devShell attribute used when none is requested.
const defaultFlakeShell = "default"

// Options is the nix variant's immutable configuration: the resolved source.
type Options struct {
	Source sharednix.NixSource
}

// SourceFromFlags builds a NixSource from the CLI's nix flags. The user-facing
// source values are "packages" and "flake"; the latter maps to the
// NixSourceProjectFlake kind, defaulting the flake ref to repoDir then workDir.
func SourceFromFlags(kind, rev string, packages []string, shell, flakeRef, repoDir, workDir string) (sharednix.NixSource, error) {
	switch kind {
	case "", "packages":
		return sharednix.NixSource{Kind: sharednix.NixSourcePackages, Rev: rev, Packages: packages}, nil
	case "flake":
		ref := flakeRef
		if ref == "" {
			ref = repoDir
		}
		if ref == "" {
			ref = workDir
		}
		if shell == "" {
			shell = defaultFlakeShell
		}
		return sharednix.NixSource{Kind: sharednix.NixSourceProjectFlake, FlakeRef: ref, Shell: shell}, nil
	default:
		return sharednix.NixSource{}, fmt.Errorf("unknown nix source %q; valid: packages, flake", kind)
	}
}

// agentNixDir is the base directory for agent-nix runtime data (GC-roots).
func agentNixDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "agent-nix")
}

// resolveFunc resolves a source to a closure; rootFunc registers a GC-root over
// it. Both are injected so the Contribution mapping is testable without host nix.
type resolveFunc func(sharednix.NixSource) (sharednix.ClosureDescriptor, error)
type rootFunc func(sharednix.NixSource, sharednix.ClosureDescriptor) error

// Provisioner resolves a NixSource to a content-pinned closure on the host,
// GC-roots it, and contributes read-only binds of ONLY the closure's requisites.
type Provisioner struct {
	resolve resolveFunc
	root    rootFunc
}

// New creates a nix provisioner backed by the host nix CLI.
func New() *Provisioner {
	return &Provisioner{resolve: ResolveClosure, root: registerGCRoot}
}

// Pinned reports true: a nix closure resolves from a flake.lock/rev-pinned source.
func (p *Provisioner) Pinned() bool { return true }

// Contribute resolves the closure and returns a mount-only Contribution: the
// requisites read-only, the closure's PATH entries, and its env. There is no
// daemon socket and no /nix/store bind — that bind discipline is what makes
// RequireHostToolsUnreachable accurate.
func (p *Provisioner) Contribute(in provshared.Input) (provision.Contribution, error) {
	src := in.NixSource
	if err := sharednix.ValidateSource(src); err != nil {
		return provision.Contribution{}, err
	}

	closure, err := p.resolve(src)
	if err != nil {
		return provision.Contribution{}, fmt.Errorf("resolve nix closure: %w", err)
	}
	// GC-root the full closure BEFORE handing it off, so a concurrent
	// nix-collect-garbage cannot evict the tools mid-run.
	if err := p.root(src, closure); err != nil {
		return provision.Contribution{}, fmt.Errorf("register gc-root: %w", err)
	}

	return provision.Contribution{
		RoBindPaths: closure.Requisites,
		PathEntries: closure.PathEntries,
		Env:         closure.Env,
	}, nil
}

// gcRootDir is the per-source GC-root directory, keyed by the content hash so the
// same source reuses one root set across runs.
func gcRootDir(src sharednix.NixSource) string {
	return filepath.Join(agentNixDir(), "gcroots", sharednix.ComputeEnvID(src))
}

// registerGCRoot roots the full dev closure. Rooting each top-level store path
// keeps its entire runtime closure (the requisites) reachable from a GC-root, so
// `nix-collect-garbage -d` cannot reap the tools while a sandbox holds them.
func registerGCRoot(src sharednix.NixSource, closure sharednix.ClosureDescriptor) error {
	roots := closure.StorePaths
	if len(roots) == 0 {
		roots = closure.Requisites
	}
	if len(roots) == 0 {
		return fmt.Errorf("closure has no store paths to root")
	}

	dir := gcRootDir(src)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, p := range roots {
		link := filepath.Join(dir, filepath.Base(p))
		// An indirect root registers <link> under /nix/var/nix/gcroots/auto, making
		// <link> (and the closure reachable from <p>) a permanent GC-root.
		if _, err := runCmd("nix-store", "--realise", p, "--add-root", link, "--indirect"); err != nil {
			return err
		}
	}
	return nil
}
