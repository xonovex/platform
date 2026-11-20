// Package fs is the file-system port. Core packages read and write files only
// through FileSystem, so a unit test drives them with the memfs adapter and
// only the osfs adapter touches a real disk.
//
// Paths are slash-separated. An adapter that backs a platform with another
// separator converts at its own edge.
package fs

import (
	"errors"
	iofs "io/fs"
)

// ErrNotExist reports a path that does not exist. Adapters wrap it, so callers
// test with errors.Is(err, fs.ErrNotExist).
var ErrNotExist = iofs.ErrNotExist

// ErrNotDir reports a directory operation on a path that is a file.
var ErrNotDir = errors.New("not a directory")

// Entry is one child of a directory.
type Entry struct {
	Name  string
	IsDir bool
}

// Info describes a path.
type Info struct {
	Size  int64
	IsDir bool
}

// FileSystem is the set of file operations the core needs.
type FileSystem interface {
	// ReadFile returns the content of a file.
	ReadFile(path string) ([]byte, error)
	// WriteFile creates or replaces a file and any missing parent directory.
	WriteFile(path string, data []byte) error
	// ReadDir lists the children of a directory sorted by name.
	ReadDir(path string) ([]Entry, error)
	// Stat describes a path.
	Stat(path string) (Info, error)
	// MkdirAll creates a directory and any missing parent.
	MkdirAll(path string) error
	// Remove deletes a file or an empty directory.
	Remove(path string) error
}

// Exists reports whether a path exists. An error other than ErrNotExist is
// returned unchanged.
func Exists(fsys FileSystem, path string) (bool, error) {
	_, err := fsys.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotExist) {
		return false, nil
	}
	return false, err
}
