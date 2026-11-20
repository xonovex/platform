// Package gitfs is a read-only file-system adapter over the tree of a git ref.
// It lets the source catalog loader read the catalog as it was at a base ref
// without a checkout.
package gitfs

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process"
)

// ErrReadOnly reports a write to a tree of a git ref.
var ErrReadOnly = errors.New("the tree of a git ref is read-only")

// FS reads the tree of Ref. Every path is relative to Dir, the working
// directory the git commands run in, so a catalog directory is read as it is
// found on disk.
type FS struct {
	Runner process.Runner
	// Dir is the directory inside the repository that paths are relative to.
	Dir string
	// Ref is the commit, tag or branch whose tree is read.
	Ref string
}

func (f FS) git(args ...string) (process.Result, error) {
	return f.Runner.Run(context.Background(), process.Command{Dir: f.Dir, Name: "git", Args: args})
}

// object names a path of the tree, relative to the working directory.
func (f FS) object(p string) string { return f.Ref + ":./" + strings.TrimPrefix(path.Clean(p), "./") }

func notExist(op, p string) error { return fmt.Errorf("%s %s at %s: %w", op, p, "git", fs.ErrNotExist) }

// ReadFile returns the content of a file at the ref.
func (f FS) ReadFile(p string) ([]byte, error) {
	result, err := f.git("show", f.object(p))
	if err != nil {
		return nil, err
	}
	if !result.Succeeded() {
		return nil, notExist("read", p)
	}
	return result.Stdout, nil
}

// Stat describes a path at the ref.
func (f FS) Stat(p string) (fs.Info, error) {
	kind, err := f.git("cat-file", "-t", f.object(p))
	if err != nil {
		return fs.Info{}, err
	}
	if !kind.Succeeded() {
		return fs.Info{}, notExist("stat", p)
	}
	if strings.TrimSpace(string(kind.Stdout)) == "tree" {
		return fs.Info{IsDir: true}, nil
	}
	size, err := f.git("cat-file", "-s", f.object(p))
	if err != nil {
		return fs.Info{}, err
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(string(size.Stdout)), 10, 64)
	return fs.Info{Size: n}, nil
}

// ReadDir lists the children of a directory at the ref, sorted by name.
func (f FS) ReadDir(p string) ([]fs.Entry, error) {
	info, err := f.Stat(p)
	if err != nil {
		return nil, err
	}
	if !info.IsDir {
		return nil, fmt.Errorf("list %s: %w", p, fs.ErrNotDir)
	}
	result, err := f.git("ls-tree", "-z", f.object(p))
	if err != nil {
		return nil, err
	}
	var entries []fs.Entry
	for _, record := range strings.Split(string(result.Stdout), "\x00") {
		meta, name, found := strings.Cut(record, "\t")
		if !found {
			continue
		}
		fields := strings.Fields(meta)
		entries = append(entries, fs.Entry{Name: name, IsDir: len(fields) > 1 && fields[1] == "tree"})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries, nil
}

// WriteFile fails: the tree of a ref cannot change.
func (FS) WriteFile(string, []byte) error { return ErrReadOnly }

// MkdirAll fails: the tree of a ref cannot change.
func (FS) MkdirAll(string) error { return ErrReadOnly }

// Remove fails: the tree of a ref cannot change.
func (FS) Remove(string) error { return ErrReadOnly }
