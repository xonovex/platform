// Package osfs is the operating-system adapter of the file-system port. It is
// the only package of the library that touches a real disk.
package osfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
)

// FileSystem reads and writes the host disk.
type FileSystem struct{}

var _ fs.FileSystem = FileSystem{}

// New returns the host file system.
func New() FileSystem { return FileSystem{} }

// ReadFile returns the content of a file.
func (FileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(filepath.FromSlash(path))
}

// WriteFile creates or replaces a file and any missing parent directory.
func (FileSystem) WriteFile(path string, data []byte) error {
	native := filepath.FromSlash(path)
	if err := os.MkdirAll(filepath.Dir(native), 0o755); err != nil {
		return fmt.Errorf("create parent of %s: %w", path, err)
	}
	return os.WriteFile(native, data, 0o644)
}

// ReadDir lists the children of a directory sorted by name.
func (FileSystem) ReadDir(path string) ([]fs.Entry, error) {
	items, err := os.ReadDir(filepath.FromSlash(path))
	if err != nil {
		return nil, err
	}
	entries := make([]fs.Entry, 0, len(items))
	for _, item := range items {
		entries = append(entries, fs.Entry{Name: item.Name(), IsDir: item.IsDir()})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

// Stat describes a path.
func (FileSystem) Stat(path string) (fs.Info, error) {
	info, err := os.Stat(filepath.FromSlash(path))
	if err != nil {
		return fs.Info{}, err
	}
	return fs.Info{Size: info.Size(), IsDir: info.IsDir()}, nil
}

// MkdirAll creates a directory and any missing parent.
func (FileSystem) MkdirAll(path string) error {
	return os.MkdirAll(filepath.FromSlash(path), 0o755)
}

// Remove deletes a file or an empty directory.
func (FileSystem) Remove(path string) error {
	return os.Remove(filepath.FromSlash(path))
}
