package plugin

import (
	"fmt"
	"io"
	"path"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// Host is the machine the plugin and marketplace commands reach: the files of
// the store, the working and home directories, the environment, and the
// approval ledger and terminal of the trust gate.
type Host interface {
	Files
	runconfig.Host
}

// Manager carries out the plugin and marketplace commands: it opens the store a
// command names, and it installs a plugin only after its digest passes the
// trust gate.
type Manager struct {
	Host    Host
	Origins origin.Registry
	// Out receives the question and the notice of the trust gate, and the result
	// of each command.
	Out io.Writer
}

// Open returns the service of the store at dir, or of the default store when dir
// is empty. A relative dir resolves against the working directory.
func (m Manager) Open(dir string) (Service, error) {
	cwd, err := m.Host.Getwd()
	if err != nil {
		return Service{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	switch {
	case dir == "":
		home, err := m.Host.UserHomeDir()
		if err != nil {
			return Service{}, fmt.Errorf("resolve user home directory: %w", err)
		}
		dir = DefaultDir(m.Host.Getenv, home)
	case !path.IsAbs(dir):
		dir = path.Join(cwd, dir)
	}
	store, err := NewStore(m.Host, path.Clean(dir))
	if err != nil {
		return Service{}, err
	}
	return Service{Store: store, Files: m.Host, Origins: m.Origins}, nil
}

// InstallRequest asks for the install of one plugin.
type InstallRequest struct {
	// Store is the store directory, empty for the default one.
	Store string
	// Source is a directory, a repository URL or name@marketplace.
	Source string
	// Allow names each capability the user grants the plugin.
	Allow []string
	// ApproveDigest is the digest of the reviewed plugin, which approves it
	// without a question.
	ApproveDigest string
}

// Install fetches the plugin of the request, shows it through preview, lets its
// digest through the trust gate and installs it with the granted capabilities.
// The capabilities are parsed before anything is fetched, and the grant is
// never given by the answer of the gate.
func (m Manager) Install(request InstallRequest, preview func(Fetched)) (plugin.Locked, error) {
	capabilities := make([]plugin.Capability, 0, len(request.Allow))
	for _, name := range request.Allow {
		capability, err := plugin.ParseCapability(name)
		if err != nil {
			return plugin.Locked{}, err
		}
		capabilities = append(capabilities, capability)
	}
	service, err := m.Open(request.Store)
	if err != nil {
		return plugin.Locked{}, err
	}
	fetched, err := service.Fetch(request.Source)
	if err != nil {
		return plugin.Locked{}, err
	}
	preview(fetched)
	subject := trust.Subject{
		Digest: fetched.Digest,
		Path:   fetched.Source,
		Notice: fmt.Sprintf("The plugin %s from %s is installed only after you approve its digest", fetched.Manifest.Name, fetched.Source),
	}
	if err := trust.Decide(subject, request.ApproveDigest, runconfig.Ledger(m.Host), m.Host, m.Out); err != nil {
		return plugin.Locked{}, err
	}
	return service.Commit(fetched, capabilities)
}

// CapabilityNames lists every capability a user may grant a plugin.
func CapabilityNames() []string {
	return capabilityNames(plugin.Capabilities())
}
