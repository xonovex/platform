package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/explain"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
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
	command.AddCommand(newConfigShowCommand(deps), newConfigMigrateCommand(deps))
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
			loaded, err := launch.LoadConfig(deps, options)
			if err != nil {
				return err
			}
			merged := loaded.Resolved.Merged()
			if sources {
				return explain.Render(deps.Out, merged)
			}
			sections := map[string]json.RawMessage{}
			for _, section := range runconfig.Sections() {
				sections[section.Name] = merged.Section(section.Name)
			}
			// The merged sections are valid JSON, which always encodes.
			data, _ := json.Marshal(sections)
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

func newConfigMigrateCommand(deps launch.Dependencies) *cobra.Command {
	var workDir string
	var managed bool
	command := &cobra.Command{
		Use:   "migrate",
		Short: "Write agentic.json from the configuration files of the previous release",
		Long: `Write the user ~/.config/sandboxed-claude/config as the user agentic.json, the
project .agent-cli.yaml, .agent-cli.yml or .agent-cli.toml as the project
agentic.json, and each legacy providers/<name>.json as providers.d/<name>.json
beside it. With --managed the command also writes the managed managed.yaml,
managed.yml or managed.toml as the managed agentic.json, which needs the rights
of an administrator. The command checks each file against the schema, never
replaces a file, and leaves the legacy files in place: remove them after you
check the result, because a layer with both is an error. A project file that
the command writes needs approval again before a run reads it.`,
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			if deps.Legacy == nil {
				return errors.New("this build reads no legacy configuration")
			}
			dir := workDir
			if !filepath.IsAbs(dir) {
				cwd, err := deps.Host.Getwd()
				if err != nil {
					return fmt.Errorf("find the working directory: %w", err)
				}
				dir = filepath.Join(cwd, dir)
			}
			migrated, err := deps.Legacy.Migrate(deps.ConfigEnv(), runconfig.Migration{ProjectDir: dir, Managed: managed})
			for _, m := range migrated {
				if _, werr := fmt.Fprintf(deps.Out, "Wrote %s from %s\n", m.Path, strings.Join(m.From, ", ")); werr != nil {
					return werr
				}
			}
			if err != nil {
				return err
			}
			if len(migrated) == 0 {
				_, err := fmt.Fprintln(deps.Out, "No legacy configuration file was found.")
				return err
			}
			_, err = fmt.Fprintln(deps.Out, "Check the new files, then remove the legacy files.")
			return err
		},
	}
	command.Flags().StringVarP(&workDir, "work-dir", "w", "", "Project directory")
	command.Flags().BoolVar(&managed, "managed", false, "Also migrate the managed file")
	return command
}
