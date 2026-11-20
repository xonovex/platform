// Package nix is the provision=nix leaf: it resolves a NixSource to a
// content-pinned closure on the host, GC-roots it, and contributes read-only
// binds of ONLY the closure's requisites (never the whole /nix/store, never the
// nix daemon socket). The shared-module nix engine is imported under the
// sharednix alias since this leaf package is itself named nix.
package nix

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	sharednix "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision/nix"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	provshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/provision"
)

// Host is the port the nix provisioner consumes: the host nix CLI, the
// directory that holds its GC-roots, and the closure files it inspects.
type Host interface {
	// Capture runs name in dir and returns its trimmed standard output.
	Capture(name string, args []string, dir string) (string, error)
	UserHomeDir() (string, error)
	MkdirAll(path string, perm fs.FileMode) error
	Stat(path string) (fs.FileInfo, error)
}

// Provisioner resolves a NixSource to a content-pinned closure on the host,
// GC-roots it, and contributes read-only binds of ONLY the closure's requisites.
type Provisioner struct {
	host Host
}

// New creates a nix provisioner backed by the nix CLI of host.
func New(host Host) *Provisioner {
	return &Provisioner{host: host}
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

	closure, err := p.ResolveClosure(src)
	if err != nil {
		return provision.Contribution{}, fmt.Errorf("resolve nix closure: %w", err)
	}
	// GC-root the full closure BEFORE handing it off, so a concurrent
	// nix-collect-garbage cannot evict the tools mid-run.
	if err := p.registerGCRoot(src, closure); err != nil {
		return provision.Contribution{}, fmt.Errorf("register gc-root: %w", err)
	}

	return provision.Contribution{
		RoBindPaths: closure.Requisites,
		PathEntries: closure.PathEntries,
		Env:         closure.Env,
	}, nil
}

// CheckAgent requires the agent executable in one of the PATH entries of the
// closure, because a nix sandbox binds no other tool.
func (p *Provisioner) CheckAgent(c provision.Contribution, agent *types.AgentConfig) (bool, error) {
	for _, directory := range c.PathEntries {
		info, err := p.host.Stat(filepath.Join(directory, agent.Binary))
		if err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			return true, nil
		}
	}
	return false, fmt.Errorf("nix provision did not supply agent executable %q; include package %q or select a flake shell that provides it", agent.Binary, agent.NixPackage)
}

// gcRootDir is the per-source GC-root directory, keyed by the content hash so the
// same source reuses one root set across runs.
func (p *Provisioner) gcRootDir(src sharednix.NixSource) (string, error) {
	home, err := p.host.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "agent-nix", "gcroots", sharednix.ComputeEnvID(src)), nil
}

// registerGCRoot roots the full dev closure. Rooting each top-level store path
// keeps its entire runtime closure (the requisites) reachable from a GC-root, so
// `nix-collect-garbage -d` cannot reap the tools while a sandbox holds them.
func (p *Provisioner) registerGCRoot(src sharednix.NixSource, closure sharednix.ClosureDescriptor) error {
	roots := closure.StorePaths
	if len(roots) == 0 {
		roots = closure.Requisites
	}
	if len(roots) == 0 {
		return fmt.Errorf("closure has no store paths to root")
	}

	dir, err := p.gcRootDir(src)
	if err != nil {
		return err
	}
	if err := p.host.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, root := range roots {
		link := filepath.Join(dir, filepath.Base(root))
		// An indirect root registers <link> under /nix/var/nix/gcroots/auto, making
		// <link> (and the closure reachable from <root>) a permanent GC-root.
		if _, err := p.host.Capture("nix-store", []string{"--realise", root, "--add-root", link, "--indirect"}, ""); err != nil {
			return err
		}
	}
	return nil
}
