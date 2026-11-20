// Package stamp records a review: it sets the review date of the sources a
// person has checked again. It changes dates and nothing else, so it is the
// only step of the tool that writes a catalog.
package stamp

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

const dateLayout = "2006-01-02"

// Result lists what a stamp changed.
type Result struct {
	// Edited holds the ids of the sources whose date changed, in catalog order.
	Edited []string
	// Files holds the catalog files that were rewritten, sorted.
	Files []string
}

// Reviewed sets the review date of the sources of the catalog to the day of now.
// A filter limits the change to the sources whose title contains it, ignoring
// case. A catalog file with no change is left alone.
func Reviewed(fsys fs.FileSystem, catalog sources.Catalog, filter string, now time.Time) (Result, error) {
	today := now.Format(dateLayout)
	needle := strings.ToLower(filter)

	byFile := map[string][]string{}
	for _, source := range catalog.Sources {
		if needle == "" || strings.Contains(strings.ToLower(source.Title), needle) {
			byFile[source.File] = append(byFile[source.File], source.ID)
		}
	}

	var result Result
	files := make([]string, 0, len(byFile))
	for file := range byFile {
		files = append(files, file)
	}
	sort.Strings(files)

	edited := map[string]bool{}
	for _, file := range files {
		changed, err := stampFile(fsys, file, byFile[file], today)
		if err != nil {
			return Result{}, err
		}
		if len(changed) == 0 {
			continue
		}
		result.Files = append(result.Files, file)
		for _, id := range changed {
			edited[id] = true
		}
	}
	for _, source := range catalog.Sources {
		if edited[source.ID] {
			result.Edited = append(result.Edited, source.ID)
		}
	}
	return result, nil
}

// stampFile rewrites one catalog file and returns the ids whose date changed.
func stampFile(fsys fs.FileSystem, file string, ids []string, today string) ([]string, error) {
	data, err := fsys.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}
	document, err := sources.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	var changed []string
	for i := range document.Sources {
		source := &document.Sources[i]
		if wanted[source.ID] && source.ReviewedAt != today {
			source.ReviewedAt = today
			changed = append(changed, source.ID)
		}
	}
	if len(changed) == 0 {
		return nil, nil
	}
	encoded, err := sources.Encode(document)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if err := fsys.WriteFile(file, encoded); err != nil {
		return nil, fmt.Errorf("write %s: %w", file, err)
	}
	return changed, nil
}
