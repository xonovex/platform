// Package plugins is the sandbox composition root: it assembles the built-in
// isolators and provisioners into a Registry. This is the ONLY place that imports
// the concrete leaf packages (isolation/{none,bwrap,docker,seatbelt,
// dockersandboxes,openshell}, provision/{none,command,nix}, credential/{env,file,
// keychain,secretservice,pass}) — the core sandbox package depends only on the axis
// ports. Adding
// a plugin means a new leaf package plus one Register call here; the selection and
// policy code never changes.
package plugins

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	credenv "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/env"
	credfile "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/file"
	credkeychain "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/keychain"
	credpass "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/pass"
	credsecretservice "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/secretservice"
	credshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/credential/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/bwrap"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/docker"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/dockersandboxes"
	isonone "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/none"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/openshell"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/seatbelt"
	isoshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/isolation/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/command"
	provnix "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/nix"
	provnone "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/none"
	provshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/provision/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox"
)

// Host is the union of the ports the built-in plugins consume.
type Host interface {
	docker.Host
	provnix.Host
	openshell.Host
	credfile.Host
	credenv.Host
}

// DefaultRegistry returns a Registry with the built-in isolators and
// provisioners over host, each writing through log.
func DefaultRegistry(host Host, log logging.Logger) *sandbox.Registry {
	return sandbox.NewRegistry().
		RegisterIsolator(isolation.IsolationNone, func() isoshared.Isolator { return isonone.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationBwrap, func() isoshared.Isolator { return bwrap.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationDocker, func() isoshared.Isolator { return docker.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationSeatbelt, func() isoshared.Isolator { return seatbelt.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationDockerSandboxes, func() isoshared.Isolator { return dockersandboxes.NewIsolator(host, log) }).
		RegisterIsolator(isolation.IsolationOpenShell, func() isoshared.Isolator { return openshell.NewIsolator(host, log) }).
		RegisterProvisioner(provision.ProvisionNone, func() provshared.Provisioner { return provnone.New() }).
		RegisterProvisioner(provision.ProvisionCommand, func() provshared.Provisioner { return command.New() }).
		RegisterProvisioner(provision.ProvisionNix, func() provshared.Provisioner { return provnix.New(host) })
}

// DefaultCredentials returns a credential registry with the built-in stores over
// host. The scheme of a binding selects the store.
func DefaultCredentials(host Host) *credshared.Registry {
	return credshared.NewRegistry().
		Register("env", func() credshared.Store { return credenv.New(host) }).
		Register("file", func() credshared.Store { return credfile.New(host) }).
		Register("keychain", func() credshared.Store { return credkeychain.New(host) }).
		Register("secret-service", func() credshared.Store { return credsecretservice.New(host) }).
		Register("pass", func() credshared.Store { return credpass.New(host) })
}
