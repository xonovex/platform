package orchestration

import (
	"sort"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

// sortFiles orders fixtures by path so a request does not depend on map order.
func sortFiles(files []harness.File) {
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
}
