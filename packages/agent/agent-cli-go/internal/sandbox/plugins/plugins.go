// Package plugins is the sandbox composition root: it assembles the built-in
// isolators and provisioners into a Registry. This is the ONLY place that
// imports the concrete plugin packages (bwrap/docker/none/nixprov) — the core
// sandbox package depends only on the interfaces. Adding a plugin means a new
// package plus one Register call here; the selection and policy code never
// changes.
package plugins

import (
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/bwrap"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/docker"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/nixprov"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox/none"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/types"
)

// DefaultRegistry returns a Registry with the built-in isolators and provisioners.
func DefaultRegistry() *sandbox.Registry {
	return sandbox.NewRegistry().
		RegisterIsolator(types.IsolationNone, func() sandbox.Isolator { return none.NewIsolator() }).
		RegisterIsolator(types.IsolationBwrap, func() sandbox.Isolator { return bwrap.NewIsolator() }).
		RegisterIsolator(types.IsolationDocker, func() sandbox.Isolator { return docker.NewIsolator() }).
		RegisterProvisioner(types.ProvisionNone, func() sandbox.Provisioner { return sandbox.NewNoneProvisioner() }).
		RegisterProvisioner(types.ProvisionCommand, func() sandbox.Provisioner { return sandbox.NewCommandProvisioner() }).
		RegisterProvisioner(types.ProvisionNix, func() sandbox.Provisioner { return nixprov.New() })
}
