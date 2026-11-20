package rules

import (
	"fmt"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
)

// Mapping returns the findings about the files a source feeds: it must feed at
// least one file, or all references, and each file it names must exist.
func Mapping(source sources.Source, exists func(path string) bool) []findings.Finding {
	var out []findings.Finding
	if !source.CoversAllReferences && len(source.Feeds) == 0 {
		out = append(out, finding(findings.Fail, CodeReferenceMissing, source.File,
			fmt.Sprintf("source %q names no file it feeds", source.ID)))
	}
	for _, feed := range source.Feeds {
		if !exists(feed) {
			out = append(out, finding(findings.Fail, CodeReferenceDangling, source.File,
				fmt.Sprintf("source %q feeds %s, which does not exist", source.ID, feed)))
		}
	}
	return out
}

// Pin returns the finding for a source that pins a version that nothing can
// check: it has a web address but neither a content digest nor a checkout with
// a commit and watched paths. A pin that cannot move is not a pin.
func Pin(source sources.Source) []findings.Finding {
	if source.Version == "" || len(source.URLs) == 0 || source.ContentSHA256 != "" {
		return nil
	}
	if source.Checkout != "" && source.Commit != "" && len(source.Watch) > 0 {
		return nil
	}
	return []findings.Finding{finding(findings.Fail, CodePinMissing, source.File,
		fmt.Sprintf("source %q pins version %s but records no content digest, and no checkout with a commit and watched paths", source.ID, source.Version))}
}

// Coverage returns the findings that hold across the sources of one catalog: it
// lists a source, and every reference file is backed by one. The reference
// files are the slash paths below the catalog directory.
func Coverage(catalog sources.Catalog, referenceFiles []string) []findings.Finding {
	if len(catalog.Sources) == 0 {
		return []findings.Finding{finding(findings.Fail, CodeNoSources, "", "the catalog lists no source")}
	}
	covered := map[string]bool{}
	for _, source := range catalog.Sources {
		if source.CoversAllReferences {
			return nil
		}
		for _, feed := range source.Feeds {
			covered[feed] = true
		}
		for _, w := range source.Watch {
			for _, feed := range w.Feeds {
				covered[feed] = true
			}
		}
	}
	var out []findings.Finding
	for _, file := range referenceFiles {
		if !covered[file] {
			out = append(out, finding(findings.Fail, CodeReferenceUncovered, "",
				fmt.Sprintf("no source backs %s", file)))
		}
	}
	return out
}
