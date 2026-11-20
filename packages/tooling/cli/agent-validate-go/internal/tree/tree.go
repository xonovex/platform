// Package tree gives a check read access to one repository through the
// file-system port, with every path spelled relative to the repository root.
package tree

import (
	"errors"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/discovery"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
)

// Tree reads the files below Root. A path that a method takes or returns is a
// slash path relative to Root, so a finding names the same file on every host.
type Tree struct {
	FS   fs.FileSystem
	Root string
	// Skip names directories that a walk does not enter, on top of the default
	// skip list of the contract.
	Skip []string
}

// Abs returns the file-system path of a repository path. An absolute path
// names itself, so a path the user spelled absolute never lands below Root.
func (t Tree) Abs(rel string) string {
	if path.IsAbs(rel) || t.Root == "" || t.Root == "." {
		return path.Clean(rel)
	}
	return path.Join(t.Root, rel)
}

// Rel returns the repository path of a file-system path below Root, or the
// path unchanged when it lies elsewhere.
func (t Tree) Rel(abs string) string {
	if t.Root == "" || t.Root == "." {
		return path.Clean(abs)
	}
	prefix := strings.TrimSuffix(t.Root, "/") + "/"
	if strings.HasPrefix(abs, prefix) {
		return strings.TrimPrefix(abs, prefix)
	}
	if abs == t.Root {
		return "."
	}
	return abs
}

// Read returns the text of a file.
func (t Tree) Read(rel string) (string, error) {
	data, err := t.FS.ReadFile(t.Abs(rel))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Optional returns the text of a file and whether it exists. An error other
// than a missing file is a fault.
func (t Tree) Optional(rel string) (string, bool, error) {
	text, err := t.Read(rel)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return text, true, nil
}

// IsFile reports whether the path is a file.
func (t Tree) IsFile(rel string) bool {
	info, err := t.FS.Stat(t.Abs(rel))
	return err == nil && !info.IsDir
}

// IsDir reports whether the path is a directory.
func (t Tree) IsDir(rel string) bool {
	info, err := t.FS.Stat(t.Abs(rel))
	return err == nil && info.IsDir
}

// Exists reports whether the path is a file or a directory.
func (t Tree) Exists(rel string) bool {
	_, err := t.FS.Stat(t.Abs(rel))
	return err == nil
}

// List returns the entries of a directory, or none when it does not exist.
func (t Tree) List(rel string) ([]fs.Entry, error) {
	entries, err := t.FS.ReadDir(t.Abs(rel))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return entries, err
}

// Walk returns the files below a directory that match, sorted. A path that
// names a file returns itself when it matches. A path that does not exist
// returns no file.
func (t Tree) Walk(rel string, match func(string) bool) ([]string, error) {
	switch {
	case t.IsFile(rel):
		if match == nil || match(rel) {
			return []string{path.Clean(rel)}, nil
		}
		return nil, nil
	case !t.IsDir(rel):
		return nil, nil
	}
	skip := append(discovery.DefaultSkip(), t.Skip...)
	found, err := discovery.Walk(t.FS, t.Abs(rel), discovery.Options{
		Skip: skip,
		Match: func(abs string) bool {
			return match == nil || match(t.Rel(abs))
		},
	})
	if err != nil {
		return nil, err
	}
	for i, file := range found {
		found[i] = t.Rel(file)
	}
	return found, nil
}

// WalkAll walks every path and returns the files once each, sorted.
func (t Tree) WalkAll(paths []string, match func(string) bool) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		files, err := t.Walk(p, match)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if !seen[file] {
				seen[file] = true
				out = append(out, file)
			}
		}
	}
	sortStrings(out)
	return out, nil
}
