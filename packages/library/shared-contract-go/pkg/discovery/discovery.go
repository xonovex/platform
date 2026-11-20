// Package discovery lists the files below a root, skipping the directories that
// hold generated or foreign content.
package discovery

import (
	"path"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
)

// DefaultSkip names the directories no check descends into: build output,
// installed packages, eval results and version-control data.
func DefaultSkip() []string {
	return []string{"dist", "node_modules", "eval-results", ".git"}
}

// Options narrow a walk.
type Options struct {
	// Skip names directories, by base name, that the walk does not enter.
	Skip []string
	// Match keeps a file when it returns true. A nil Match keeps every file.
	Match func(path string) bool
}

// Walk returns the slash paths of the files below root in lexical order. A
// directory named in Skip is not entered, wherever it sits.
func Walk(fsys fs.FileSystem, root string, opts Options) ([]string, error) {
	skip := make(map[string]struct{}, len(opts.Skip))
	for _, name := range opts.Skip {
		skip[name] = struct{}{}
	}
	var found []string
	if err := walk(fsys, root, skip, opts.Match, &found); err != nil {
		return nil, err
	}
	sort.Strings(found)
	return found, nil
}

func walk(fsys fs.FileSystem, dir string, skip map[string]struct{}, match func(string) bool, found *[]string) error {
	entries, err := fsys.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		full := path.Join(dir, entry.Name)
		if entry.IsDir {
			if _, skipped := skip[entry.Name]; skipped {
				continue
			}
			if err := walk(fsys, full, skip, match, found); err != nil {
				return err
			}
			continue
		}
		if match == nil || match(full) {
			*found = append(*found, full)
		}
	}
	return nil
}

// Named returns a Match that keeps the files with the given base name.
func Named(name string) func(string) bool {
	return func(p string) bool { return path.Base(p) == name }
}

// WithExtension returns a Match that keeps the files with the given extension,
// written with its dot.
func WithExtension(ext string) func(string) bool {
	return func(p string) bool { return path.Ext(p) == ext }
}
