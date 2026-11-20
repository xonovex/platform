// Package memfs is the in-memory adapter of the file-system port. It backs
// every unit test, so no test reads or writes a real disk.
package memfs

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
)

// FileSystem keeps files and directories in a map guarded by a mutex.
type FileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
	dirs  map[string]struct{}
}

var _ fs.FileSystem = (*FileSystem)(nil)

// New returns a file system holding the given files, keyed by slash path. The
// parent directories of each file exist.
func New(files map[string]string) *FileSystem {
	fsys := &FileSystem{files: map[string][]byte{}, dirs: map[string]struct{}{".": {}, "/": {}}}
	for name, content := range files {
		fsys.put(name, []byte(content))
	}
	return fsys
}

func clean(p string) string { return path.Clean(strings.TrimPrefix(p, "./")) }

// put stores a file and registers its parent directories. The caller holds the
// write lock or owns the value.
func (m *FileSystem) put(name string, data []byte) {
	name = clean(name)
	m.files[name] = append([]byte(nil), data...)
	m.addDirs(path.Dir(name))
}

func (m *FileSystem) addDirs(dir string) {
	for dir != "." && dir != "/" {
		m.dirs[dir] = struct{}{}
		dir = path.Dir(dir)
	}
}

func notExist(op, p string) error {
	return fmt.Errorf("%s %s: %w", op, p, fs.ErrNotExist)
}

// ReadFile returns a copy of the file content.
func (m *FileSystem) ReadFile(p string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p = clean(p)
	data, ok := m.files[p]
	if !ok {
		if _, isDir := m.dirs[p]; isDir {
			return nil, fmt.Errorf("read %s: is a directory", p)
		}
		return nil, notExist("read", p)
	}
	return append([]byte(nil), data...), nil
}

// WriteFile stores a file and creates its parent directories. A path that is a
// directory is rejected.
func (m *FileSystem) WriteFile(p string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p = clean(p)
	if _, isDir := m.dirs[p]; isDir {
		return fmt.Errorf("write %s: is a directory", p)
	}
	m.put(p, data)
	return nil
}

// ReadDir lists the children of a directory sorted by name.
func (m *FileSystem) ReadDir(p string) ([]fs.Entry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p = clean(p)
	if _, isFile := m.files[p]; isFile {
		return nil, fmt.Errorf("read dir %s: %w", p, fs.ErrNotDir)
	}
	if _, isDir := m.dirs[p]; !isDir {
		return nil, notExist("read dir", p)
	}
	children := map[string]bool{}
	for name := range m.files {
		if child, ok := childOf(p, name); ok {
			if _, seen := children[child]; !seen {
				children[child] = false
			}
		}
	}
	for name := range m.dirs {
		if child, ok := childOf(p, name); ok {
			children[child] = true
		}
	}
	entries := make([]fs.Entry, 0, len(children))
	for name, isDir := range children {
		entries = append(entries, fs.Entry{Name: name, IsDir: isDir})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

// childOf returns the first path element of name below dir. A file below a
// nested directory yields that directory name, which the dirs map marks as a
// directory.
func childOf(dir, name string) (string, bool) {
	rest := name
	if dir != "." {
		prefix := strings.TrimSuffix(dir, "/") + "/"
		if !strings.HasPrefix(name, prefix) {
			return "", false
		}
		rest = strings.TrimPrefix(name, prefix)
	}
	if rest == "" || rest == "." {
		return "", false
	}
	first, _, _ := strings.Cut(rest, "/")
	return first, first != ""
}

// Stat describes a file or directory.
func (m *FileSystem) Stat(p string) (fs.Info, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p = clean(p)
	if data, ok := m.files[p]; ok {
		return fs.Info{Size: int64(len(data))}, nil
	}
	if _, ok := m.dirs[p]; ok {
		return fs.Info{IsDir: true}, nil
	}
	return fs.Info{}, notExist("stat", p)
}

// MkdirAll creates a directory and its parents. A path that is a file is
// rejected.
func (m *FileSystem) MkdirAll(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p = clean(p)
	if _, isFile := m.files[p]; isFile {
		return fmt.Errorf("mkdir %s: %w", p, fs.ErrNotDir)
	}
	m.addDirs(p)
	return nil
}

// Remove deletes a file or an empty directory.
func (m *FileSystem) Remove(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p = clean(p)
	if _, ok := m.files[p]; ok {
		delete(m.files, p)
		return nil
	}
	if _, ok := m.dirs[p]; !ok {
		return notExist("remove", p)
	}
	for name := range m.files {
		if _, below := childOf(p, name); below {
			return fmt.Errorf("remove %s: directory not empty", p)
		}
	}
	for name := range m.dirs {
		if _, below := childOf(p, name); below {
			return fmt.Errorf("remove %s: directory not empty", p)
		}
	}
	delete(m.dirs, p)
	return nil
}
