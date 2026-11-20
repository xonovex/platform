package launch

import (
	"io"
	"io/fs"
	"time"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/egress"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	credshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/credential"
	isoshared "github.com/xonovex/platform/packages/library/shared-launch-go/pkg/isolation"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/sandbox"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

// Host is the port the executor consumes to locate the run and to time it: the
// working directory, the home directory, the configuration files and the
// terminal of the trust gate, the environment the provider route reads its
// credential from, and the clock.
type Host interface {
	runconfig.Host
	// ReadDir returns the entries of a directory, sorted by name.
	ReadDir(path string) ([]fs.DirEntry, error)
	Now() time.Time
}

// Recorder stores the record of a finished run and returns its id. NextID names
// the id a run that starts at an instant expects, so the run carries it while it
// runs.
type Recorder interface {
	NextID(at time.Time) (string, error)
	Write(record.Record) (string, error)
}

// EgressOpener starts the enforcement of a managed network run: the allowlist
// proxy for a transport that needs one. It returns the route the isolator
// applies, and a stop function that ends the enforcement.
type EgressOpener interface {
	Open(transport network.EgressTransport, allow egress.Allowlist) (isoshared.EgressRoute, func(), error)
}

// Dependencies are the ports and plugin sets one run executes through.
type Dependencies struct {
	Host Host
	Log  logging.Logger
	// Out receives the command line of a dry run.
	Out       io.Writer
	Sandboxes *sandbox.Registry[isoshared.Isolator]
	// Egress starts managed network enforcement; it may be nil for a build that
	// offers no managed egress, and a managed run then fails.
	Egress EgressOpener
	// Credentials maps a credential scheme to its store; it may be nil for a build
	// without credential stores, and a run that binds a credential then fails.
	Credentials *credshared.Registry
	// Overlays builds the artifact overlay; it may be nil for a build without one.
	Overlays Overlays
	// Records stores the run record; it may be nil for a build that keeps none.
	Records Recorder
	// Validator checks each agentic.json against the schemas of the contract.
	Validator  jsonschema.Validator
	Workspaces map[wsp.VCSType]wsshared.VCS
	Terminals  map[termshared.TerminalType]termshared.TerminalExecutor
	// AutoTerminals lists, among Terminals, the wrappers that --terminal auto
	// chooses from by the session the CLI runs in.
	AutoTerminals []termshared.TerminalType
}

// ConfigEnv is the environment the configuration of a run is read through.
// The trust gate asks its question on Out.
func (d Dependencies) ConfigEnv() runconfig.Env {
	return runconfig.Env{Host: d.Host, Validator: d.Validator, Out: d.Out}
}
