package record

import (
	"io/fs"
	"strings"
	"testing/fstest"
)

// memFiles is the Files port over an in-memory tree. Paths are absolute; the
// tree stores them without the leading slash, as fs.FS requires.
type memFiles struct{ tree fstest.MapFS }

func newMemFiles() *memFiles { return &memFiles{tree: fstest.MapFS{}} }

func key(path string) string { return strings.TrimPrefix(path, "/") }

// AddFile places a file and returns the tree, for chaining.
func (m *memFiles) AddFile(path, data string, perm fs.FileMode) *memFiles {
	m.tree[key(path)] = &fstest.MapFile{Data: []byte(data), Mode: perm}
	return m
}

// AddDir places a directory and returns the tree, for chaining.
func (m *memFiles) AddDir(path string) *memFiles {
	m.tree[key(path)] = &fstest.MapFile{Mode: fs.ModeDir | 0o755}
	return m
}

func (m *memFiles) ReadFile(path string) ([]byte, error) { return m.tree.ReadFile(key(path)) }

func (m *memFiles) ReadDir(path string) ([]fs.DirEntry, error) { return m.tree.ReadDir(key(path)) }

func (m *memFiles) Stat(path string) (fs.FileInfo, error) { return m.tree.Stat(key(path)) }

func (m *memFiles) MkdirAll(path string, _ fs.FileMode) error {
	m.AddDir(path)
	return nil
}

func (m *memFiles) WriteFile(path string, data []byte, perm fs.FileMode) error {
	m.tree[key(path)] = &fstest.MapFile{Data: data, Mode: perm}
	return nil
}
