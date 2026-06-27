package nixprov

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/nix"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// defaultFlakeShell is the devShell attribute used when none is requested.
const defaultFlakeShell = "default"

// agentNixDir is the base directory for agent-nix runtime data (GC-roots).
func agentNixDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "agent-nix")
}

// resolveFunc resolves a source to a closure; rootFunc registers a GC-root over
// it. Both are injected so the Contribution mapping is testable without host nix.
type resolveFunc func(nix.NixSource) (nix.ClosureDescriptor, error)
type rootFunc func(nix.NixSource, nix.ClosureDescriptor) error

// Provisioner resolves a NixSource to a content-pinned closure on the host,
// GC-roots it, and contributes read-only binds of ONLY the closure's requisites
// (never the whole /nix/store, never the nix daemon socket).
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
func (p *Provisioner) Contribute(cfg *types.SandboxConfig) (types.Contribution, error) {
	src, err := sourceFromConfig(cfg)
	if err != nil {
		return types.Contribution{}, err
	}
	if err := nix.ValidateSource(src); err != nil {
		return types.Contribution{}, err
	}

	closure, err := p.resolve(src)
	if err != nil {
		return types.Contribution{}, fmt.Errorf("resolve nix closure: %w", err)
	}
	// GC-root the full closure BEFORE handing it off, so a concurrent
	// nix-collect-garbage cannot evict the tools mid-run.
	if err := p.root(src, closure); err != nil {
		return types.Contribution{}, fmt.Errorf("register gc-root: %w", err)
	}

	return types.Contribution{
		RoBindPaths: closure.Requisites,
		PathEntries: closure.PathEntries,
		Env:         closure.Env,
	}, nil
}

// sourceFromConfig builds the NixSource from the sandbox config's nix inputs.
// The user-facing source values are "packages" and "flake"; the latter maps to
// the NixSourceProjectFlake kind.
func sourceFromConfig(cfg *types.SandboxConfig) (nix.NixSource, error) {
	switch cfg.NixSourceKind {
	case "", "packages":
		return nix.NixSource{Kind: nix.NixSourcePackages, Rev: cfg.NixRev, Packages: cfg.NixPackages}, nil
	case "flake":
		ref := cfg.NixFlakeRef
		if ref == "" {
			ref = cfg.RepoDir
		}
		if ref == "" {
			ref = cfg.WorkDir
		}
		shell := cfg.NixShell
		if shell == "" {
			shell = defaultFlakeShell
		}
		return nix.NixSource{Kind: nix.NixSourceProjectFlake, FlakeRef: ref, Shell: shell}, nil
	default:
		return nix.NixSource{}, fmt.Errorf("unknown nix source %q; valid: packages, flake", cfg.NixSourceKind)
	}
}

// gcRootDir is the per-source GC-root directory, keyed by the content hash so the
// same source reuses one root set across runs.
func gcRootDir(src nix.NixSource) string {
	return filepath.Join(agentNixDir(), "gcroots", nix.ComputeEnvID(src))
}

// registerGCRoot roots the full dev closure. Rooting each top-level store path
// keeps its entire runtime closure (the requisites) reachable from a GC-root, so
// `nix-collect-garbage -d` cannot reap the tools while a sandbox holds them.
func registerGCRoot(src nix.NixSource, closure nix.ClosureDescriptor) error {
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
