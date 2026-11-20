// Package cmd is the driving adapter of the tool: it turns command-line input
// into calls on the protocol, run, baseline, cost and migrate packages and
// writes the outcome. Every port reaches it through Dependencies, so a test runs
// the whole command tree against in-memory adapters and a fake harness.
package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/protocol"
)

// Name is the name of the binary.
const Name = "agent-eval"

// version is set at link time through -ldflags -X.
var version = "dev"

// ConfigFiles reads the layers of agentic.json and keeps the approval ledger.
type ConfigFiles interface {
	discover.Files
	trust.Store
}

// Dependencies are the ports and adapters the command tree runs on.
type Dependencies struct {
	FS        fs.FileSystem
	Validator jsonschema.Validator
	// ConfigFiles, Places and Prompter reach the layers of agentic.json and
	// ask the user to approve a project file.
	ConfigFiles ConfigFiles
	Places      discover.Places
	Prompter    trust.Prompter
	// Protocols holds the eval protocols by name.
	Protocols *protocol.Registry
	// Engine builds, over the facts of the harnesses, the executor that runs one
	// model request and the predicates that check what a run showed.
	Engine func(facts agents.Catalog) (harness.Executor, *predicate.Registry, error)
	// Env holds the environment variables the tool reads.
	Env map[string]string
	// Dir is the working directory that relative paths resolve against.
	Dir string
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
		Use:   Name,
		Short: "Evaluate skills, instructions and agent flows against real models",
		Long: "Evaluate skills, instructions and agent flows against real models.\n\n" +
			"Exit codes: 0 pass, 1 fail, 2 invalid input or infrastructure fault, 3 stopped by a budget with work left.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(deps.Out)
	root.SetErr(deps.Err)
	root.AddCommand(newProtocolCommands(deps)...)
	root.AddCommand(newBaselineCommand(deps), newCostCommand(deps), newMigrateCommand(deps), newConfigCommand(deps), newCompletionCommand())
	return root
}
