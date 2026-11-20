// Package shared is the origin axis core: where a plugin or a marketplace comes
// from. An Origin makes a source available as a directory, and ReadTree reads a
// directory into the files a digest covers. The core names no concrete origin
// (a directory or a git repository), and reaches the machine through the Files
// port only.
package shared

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
)

// Limits on what one source may hold. A source over a limit is refused and not
// truncated, because a digest of a part of the source would approve less than the
// user sees.
const (
	MaxFiles     = 4096
	MaxFileBytes = 8 << 20
	MaxTreeBytes = 64 << 20
)

// Files is the read-only view of the file system an origin consumes.
type Files interface {
	Lstat(path string) (fs.FileInfo, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	ReadFile(path string) ([]byte, error)
}

// Checkout is a source made available as a directory.
type Checkout struct {
	// Dir is the absolute directory that holds the source.
	Dir string
	// Source names the source as the lock file records it: an absolute path, or
	// the URL of a repository with its reference.
	Source string
	// Resolved names the exact revision of the source, empty for a directory.
	Resolved string
	// Release removes what Fetch created. It is never nil.
	Release func()
}

// Origin is one way a source reaches the machine.
type Origin interface {
	// Accepts reports whether spec names a source of this origin.
	Accepts(spec string) bool
	// Fetch makes the source available as a directory.
	Fetch(spec string) (Checkout, error)
}

// Registry maps the name of each origin to its implementation.
type Registry map[string]Origin

// Select returns the origin that accepts spec. It fails closed when none does,
// and when more than one does, because the spec would then name two sources.
func (r Registry) Select(spec string) (Origin, error) {
	names := make([]string, 0, len(r))
	for name := range r {
		names = append(names, name)
	}
	sort.Strings(names)
	var accepted []string
	for _, name := range names {
		if r[name].Accepts(spec) {
			accepted = append(accepted, name)
		}
	}
	switch len(accepted) {
	case 0:
		return nil, fmt.Errorf("source %q is neither an existing directory nor a repository URL (https://, ssh://, file:// or user@host:path, with an optional #branch-or-tag)", spec)
	case 1:
		return r[accepted[0]], nil
	}
	return nil, fmt.Errorf("source %q is ambiguous: %s all accept it", spec, strings.Join(accepted, ", "))
}

// ReadTree reads the regular files below dir. It skips the .git directory and
// fails closed on a symlink, a file that is neither regular nor a directory, and
// a source over the limits.
func ReadTree(files Files, dir string) (artifact.Tree, error) {
	var collected []artifact.File
	total := 0
	var walk func(current, relative string) error
	walk = func(current, relative string) error {
		entries, err := files.ReadDir(current)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			name := entry.Name()
			if relative == "" && name == ".git" {
				continue
			}
			full := path.Join(current, name)
			rel := path.Join(relative, name)
			info, err := files.Lstat(full)
			if err != nil {
				return err
			}
			switch {
			case info.Mode()&fs.ModeSymlink != 0:
				return fmt.Errorf("%s is a symlink, which a plugin may not hold", rel)
			case info.IsDir():
				if err := walk(full, rel); err != nil {
					return err
				}
			case info.Mode().IsRegular():
				if len(collected) >= MaxFiles {
					return fmt.Errorf("the source holds more than %d files", MaxFiles)
				}
				if info.Size() > MaxFileBytes {
					return fmt.Errorf("%s is larger than %d bytes", rel, MaxFileBytes)
				}
				content, err := files.ReadFile(full)
				if err != nil {
					return err
				}
				if total += len(content); total > MaxTreeBytes {
					return fmt.Errorf("the source holds more than %d bytes", MaxTreeBytes)
				}
				collected = append(collected, artifact.File{Path: rel, Executable: info.Mode().Perm()&0o111 != 0, Content: content})
			default:
				return fmt.Errorf("%s is neither a regular file nor a directory", rel)
			}
		}
		return nil
	}
	if err := walk(dir, ""); err != nil {
		return nil, err
	}
	return artifact.NewTree(collected)
}

// ReadFiles reads the named files below dir that exist, as a tree. A name that is
// absent is left out, and a name that is a directory or a symlink fails closed.
func ReadFiles(files Files, dir string, names []string) (artifact.Tree, error) {
	var collected []artifact.File
	for _, name := range names {
		full := path.Join(dir, name)
		info, err := files.Lstat(full)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() || info.Size() > MaxFileBytes {
			return nil, fmt.Errorf("%s is not a regular file within the size limit", name)
		}
		content, err := files.ReadFile(full)
		if err != nil {
			return nil, err
		}
		collected = append(collected, artifact.File{Path: name, Content: content})
	}
	return artifact.NewTree(collected)
}

// CheckNoSymlinks fails when a component of relative, below base, is a symlink,
// so that a path a marketplace names cannot lead out of the checkout.
func CheckNoSymlinks(files Files, base, relative string) error {
	current := base
	for _, segment := range slices.DeleteFunc(strings.Split(relative, "/"), func(s string) bool { return s == "" }) {
		current = path.Join(current, segment)
		info, err := files.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink, which a marketplace path may not contain", segment)
		}
	}
	return nil
}
