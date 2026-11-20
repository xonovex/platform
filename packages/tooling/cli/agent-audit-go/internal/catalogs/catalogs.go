// Package catalogs finds the source catalogs below a directory.
package catalogs

import (
	"path"
	"sort"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/discovery"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

// Discover returns the directories below root that hold a source catalog: a
// sources.json, or shards in a sources.d directory. A directory appears once
// whatever number of files it holds, and the result is sorted.
func Discover(fsys fs.FileSystem, root string) ([]string, error) {
	files, err := discovery.Walk(fsys, root, discovery.Options{
		Skip:  discovery.DefaultSkip(),
		Match: isCatalogFile,
	})
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var dirs []string
	for _, file := range files {
		dir := path.Dir(file)
		if path.Base(dir) == sources.ShardDir {
			dir = path.Dir(dir)
		}
		if !seen[dir] {
			seen[dir] = true
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func isCatalogFile(file string) bool {
	if path.Base(file) == sources.CatalogFile {
		return true
	}
	return path.Ext(file) == ".json" && path.Base(path.Dir(file)) == sources.ShardDir
}
