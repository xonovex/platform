// Package plugins is the one place that names the concrete transports. A new
// transport is a package under internal/transport plus one Register call here.
package plugins

import (
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/filesystem"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/git"
	httptransport "github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/http"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web"
)

// Ports are the adapters the transports read through.
type Ports struct {
	FS     fs.FileSystem
	Runner process.Runner
	Web    web.Client
}

// DefaultRegistry returns a registry with every transport this tool ships.
func DefaultRegistry(ports Ports) *observe.Registry {
	return observe.NewRegistry().
		Register(filesystem.Transport{FS: ports.FS}).
		Register(git.Transport{FS: ports.FS, Runner: ports.Runner}).
		Register(httptransport.Transport{Client: ports.Web})
}
