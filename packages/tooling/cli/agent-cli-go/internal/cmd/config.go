package cmd

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/explain"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// newConfigCommand builds `agent-cli config`.
func newConfigCommand(deps launch.Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Inspect the agentic.json configuration of a run",
		Long: `agent-cli reads the run and providers sections of agentic.json in three layers:
the managed file (/etc/agentic/agentic.json on Linux), the user file
($XDG_CONFIG_HOME/agentic/agentic.json) and the nearest agentic.json of the
project. The managed layer wins and the command line cannot contradict it, a
project file sets only the keys that cannot widen what the sandbox reaches, and
the run and providers sections of a project file take effect only after you
approve its digest.`,
		Args: cobra.NoArgs,
	}
	command.AddCommand(newConfigShowCommand(deps))
	return command
}

func newConfigShowCommand(deps launch.Dependencies) *cobra.Command {
	options := launch.Options{}
	var sources bool
	command := &cobra.Command{
		Use:   "show",
		Short: "Print the merged run and providers sections",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			view, err := launch.ShowConfig(deps, options)
			if err != nil {
				return err
			}
			if sources {
				return explain.Render(deps.Out, view.Merged)
			}
			// The merged sections are valid JSON, which always encodes.
			data, _ := json.Marshal(view.Sections)
			var out bytes.Buffer
			_ = json.Indent(&out, data, "", "  ")
			out.WriteByte('\n')
			_, err = out.WriteTo(deps.Out)
			return err
		},
	}
	bindProviderLocation(command, &options)
	bindApproval(command, &options)
	command.Flags().BoolVar(&sources, "sources", false, "Name the layer and the file of each value")
	return command
}
