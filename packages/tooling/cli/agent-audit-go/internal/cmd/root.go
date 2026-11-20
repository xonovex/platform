// Package cmd is the driving adapter of the tool: it turns command-line input
// into calls on the audit, compare, stamp and convert packages and writes the
// outcome. Every port reaches it through Dependencies, so a test runs the whole
// command tree against in-memory adapters.
package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/sources"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/observe"
)

// Name is the name of the binary.
const Name = "agent-audit"

// version is set at link time through -ldflags -X.
var version = "dev"

// Dependencies are the ports and adapters the command tree runs on.
type Dependencies struct {
	FS fs.FileSystem
	// ConfigFiles and Places reach the layers of agentic.json.
	ConfigFiles discover.Files
	Places      discover.Places
	// Runner runs git for the comparison against a base ref.
	Runner process.Runner
	// Transports reads the current state of each source.
	Transports *observe.Registry
	// Validator checks a converted catalog against its schema.
	Validator jsonschema.Validator
	// LoaderFor returns a catalog loader that reads through the file system. The
	// comparison uses it to read a catalog from a git ref.
	LoaderFor func(fs.FileSystem) (sources.Loader, error)
	// Now returns the current instant.
	Now func() time.Time
	// Out receives the report, Err the notes about what the tool did.
	Out io.Writer
	Err io.Writer
}

// ExitError carries the exit status of a run that finished with a result the
// caller must act on. The binary exits with Code and prints nothing more.
type ExitError struct{ Code int }

// Error describes the status.
func (e ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Code) }

// NewRootCommand builds the command tree.
func NewRootCommand(deps Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:           Name,
		Short:         "Audit the sources that skills and documents draw on",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)
	root.AddCommand(newSourcesCommand(deps), newConvertCommand(deps), newConfigCommand(deps), newCompletionCommand())
	return root
}
