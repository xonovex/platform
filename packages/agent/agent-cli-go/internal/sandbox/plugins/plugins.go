// Package plugins is the sandbox composition root: it assembles the built-in
// isolators and provisioners into a Registry. This is the ONLY place that imports
// the concrete leaf packages (isolation/{none,bwrap,docker}, provision/{none,
// command,nix}) — the core sandbox package depends only on the axis ports. Adding
// a plugin means a new leaf package plus one Register call here; the selection and
// policy code never changes.
package plugins

import (
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/bwrap"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/docker"
	isonone "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/none"
	isoshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/command"
	provnix "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/nix"
	provnone "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/none"
	provshared "github.com/xonovex/platform/packages/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/cli/agent-cli-go/internal/sandbox"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/shared/shared-agent-go/pkg/provision"
)

// DefaultRegistry returns a Registry with the built-in isolators and provisioners.
func DefaultRegistry() *sandbox.Registry {
	return sandbox.NewRegistry().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Isolator { return isonone.NewIsolator() }).
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Isolator { return bwrap.NewIsolator() }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return docker.NewIsolator() }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return provnone.New() }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return command.New() }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return provnix.New() })
}
