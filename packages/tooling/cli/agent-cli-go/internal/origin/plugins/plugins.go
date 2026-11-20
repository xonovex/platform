// Package plugins is the origin composition root: the only package that imports
// the concrete origins. Add an origin as a package plus one entry here.
package plugins

import (
	"io/fs"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/directory"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/git"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
)

// Host is the machine the registered origins consult.
type Host interface {
	Getwd() (string, error)
	Stat(path string) (fs.FileInfo, error)
	MkdirTemp(pattern string) (string, error)
	RemoveAll(path string) error
	Stream(name string, args []string, dir string) error
	Capture(name string, args []string, dir string) (string, error)
}

// DefaultRegistry returns the origins of this build over host.
func DefaultRegistry(host Host) shared.Registry {
	return shared.Registry{
		"directory": directory.New(host),
		"git":       git.New(host),
	}
}
