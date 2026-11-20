package cmd

import (
	"path"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// repositoryPaths spells each path the user gave as the checks spell it: an
// absolute path below the root becomes relative to the root, so a finding or a
// budget names the same file however the user spelled it, and every other path
// stays as given.
func repositoryPaths(t tree.Tree, paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = t.Rel(path.Clean(p))
	}
	return out
}
