package runconfig

import (
	"io/fs"
	"path/filepath"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
)

// Host is the machine the configuration is read through: the files, the
// facts that name the directories of the platform, and the terminal of the
// trust gate.
type Host interface {
	ReadFile(path string) ([]byte, error)
	Stat(path string) (fs.FileInfo, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Rename(oldPath, newPath string) error
	Getenv(key string) string
	Getwd() (string, error)
	UserHomeDir() (string, error)
	GOOS() string
	trust.Prompter
}

// Files adapts the host to the file ports of shared-config-go: the layers and
// shards are read through it, and the approval ledger is written through it.
type Files struct{ Host Host }

// ReadFile returns the content of a file.
func (f Files) ReadFile(path string) ([]byte, error) { return f.Host.ReadFile(path) }

// Stat describes a path.
func (f Files) Stat(path string) (discover.Info, error) {
	info, err := f.Host.Stat(path)
	if err != nil {
		return discover.Info{}, err
	}
	return discover.Info{Mode: info.Mode()}, nil
}

// ReadDir names the entries of a directory.
func (f Files) ReadDir(path string) ([]string, error) {
	entries, err := f.Host.ReadDir(path)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

// WriteFile replaces a file through a rename, so a reader sees the old file
// or the new one and never a part of it. The file and its directory are
// private.
func (f Files) WriteFile(path string, data []byte) error {
	if err := f.Host.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	staged := path + ".tmp"
	if err := f.Host.WriteFile(staged, data, 0o600); err != nil {
		return err
	}
	return f.Host.Rename(staged, path)
}

// Places names the directories of the layers from the facts of the host.
type Places struct{ Host Host }

func (p Places) facts() discover.Host {
	// A host without a home directory still has the variables that replace
	// it; the directory that needs the home reports discover.ErrNoHome.
	home, _ := p.Host.UserHomeDir()
	return discover.Host{GOOS: p.Host.GOOS(), Getenv: p.Host.Getenv, Home: home}
}

// ManagedDir is the directory of the managed file.
func (p Places) ManagedDir() string { return p.facts().ManagedDir() }

// UserDir is the directory of the user file.
func (p Places) UserDir() (string, error) { return p.facts().UserDir() }

// StateDir is the state directory of the user.
func (p Places) StateDir() (string, error) { return p.facts().StateDir() }

// Getwd is the working directory.
func (p Places) Getwd() (string, error) { return p.Host.Getwd() }

// Ledger is the approval ledger of the user in the state directory.
func Ledger(host Host) trust.Ledger {
	return trust.Ledger{Store: Files{Host: host}, StateDir: Places{Host: host}.StateDir}
}
