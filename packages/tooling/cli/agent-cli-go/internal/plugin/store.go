// Package plugin is the plugin store of agent-cli: a content-addressed directory
// of installed plugins, a lock file that records each plugin with its source and
// digest, and a registry of marketplaces. It reaches the machine through the
// Files port and the sources through the origin axis, so a unit test runs on
// in-memory doubles.
package plugin

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
)

// Files is the port the store reads and writes through.
type Files interface {
	origin.Files
	Stat(path string) (fs.FileInfo, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Rename(oldPath, newPath string) error
	RemoveAll(path string) error
}

const (
	lockFile         = "lock.json"
	registryFile     = "marketplaces.json"
	contentDirectory = "content"
	stagingPrefix    = ".staging-"
)

// DefaultDir returns the store directory: agent-cli/plugins below the XDG data
// directory when it is absolute, else below ~/.local/share.
func DefaultDir(getenv func(string) string, home string) string {
	data := getenv("XDG_DATA_HOME")
	if !path.IsAbs(data) {
		data = path.Join(home, ".local", "share")
	}
	return path.Join(data, "agent-cli", "plugins")
}

// Store is the plugin directory on disk.
type Store struct {
	files Files
	dir   string
}

// NewStore returns the store at dir, which must be absolute.
func NewStore(files Files, dir string) (Store, error) {
	if !path.IsAbs(dir) || path.Clean(dir) != dir {
		return Store{}, fmt.Errorf("the plugin store %q is not a clean absolute path", dir)
	}
	return Store{files: files, dir: dir}, nil
}

// Dir returns the directory of the store.
func (s Store) Dir() string { return s.dir }

// Lock reads the lock file. A store without a lock file has an empty lock.
func (s Store) Lock() (plugin.Lock, error) {
	data, err := s.files.ReadFile(path.Join(s.dir, lockFile))
	if errors.Is(err, fs.ErrNotExist) {
		return plugin.EmptyLock(), nil
	}
	if err != nil {
		return plugin.Lock{}, err
	}
	return plugin.DecodeLock(data)
}

// SaveLock replaces the lock file.
func (s Store) SaveLock(lock plugin.Lock) error {
	data, err := plugin.EncodeLock(lock)
	if err != nil {
		return err
	}
	return s.write(lockFile, data)
}

// Registry reads the marketplace registry. A store without one has an empty one.
func (s Store) Registry() (plugin.Registry, error) {
	data, err := s.files.ReadFile(path.Join(s.dir, registryFile))
	if errors.Is(err, fs.ErrNotExist) {
		return plugin.EmptyRegistry(), nil
	}
	if err != nil {
		return plugin.Registry{}, err
	}
	return plugin.DecodeRegistry(data)
}

// SaveRegistry replaces the marketplace registry.
func (s Store) SaveRegistry(registry plugin.Registry) error {
	data, err := plugin.EncodeRegistry(registry)
	if err != nil {
		return err
	}
	return s.write(registryFile, data)
}

// write replaces a file of the store through a rename, so a reader sees the old
// file or the new one and never a part of it.
func (s Store) write(name string, data []byte) error {
	if err := s.files.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	staged := path.Join(s.dir, name+".tmp")
	if err := s.files.WriteFile(staged, data, 0o600); err != nil {
		return err
	}
	return s.files.Rename(staged, path.Join(s.dir, name))
}

func (s Store) contentDir(digest string) string {
	return path.Join(s.dir, contentDirectory, strings.TrimPrefix(digest, artifact.DigestPrefix))
}

// Put writes the tree below the directory of its digest. A tree that is already
// there is kept. The files go into a staging directory first and move into place
// in one rename.
func (s Store) Put(tree artifact.Tree) error {
	digest := tree.Digest()
	target := s.contentDir(digest)
	if _, err := s.files.Stat(target); err == nil {
		return nil
	}
	staging := path.Join(path.Dir(target), stagingPrefix+path.Base(target))
	if err := s.files.RemoveAll(staging); err != nil {
		return err
	}
	if err := s.files.MkdirAll(staging, 0o700); err != nil {
		return err
	}
	for _, file := range tree {
		full := path.Join(staging, file.Path)
		if err := s.files.MkdirAll(path.Dir(full), 0o700); err != nil {
			return err
		}
		perm := fs.FileMode(0o644)
		if file.Executable {
			perm = 0o755
		}
		if err := s.files.WriteFile(full, file.Content, perm); err != nil {
			return err
		}
	}
	return s.files.Rename(staging, target)
}

// Get reads the tree stored for digest and fails closed when the files on disk no
// longer have that digest.
func (s Store) Get(digest string) (artifact.Tree, error) {
	tree, err := origin.ReadTree(s.files, s.contentDir(digest))
	if err != nil {
		return nil, fmt.Errorf("read the stored plugin %s: %w", digest, err)
	}
	if actual := tree.Digest(); actual != digest {
		return nil, fmt.Errorf("the stored plugin %s changed on disk: its files have the digest %s; remove the plugin and install it again", digest, actual)
	}
	return tree, nil
}

// Delete removes the tree stored for digest.
func (s Store) Delete(digest string) error { return s.files.RemoveAll(s.contentDir(digest)) }
