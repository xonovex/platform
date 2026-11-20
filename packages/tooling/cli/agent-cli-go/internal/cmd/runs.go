package cmd

import (
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// RecordPorts are the ports of the runs command.
type RecordPorts struct {
	// Records is nil for a build or a machine that keeps no run records.
	Records launch.RecordReader
	// Out receives the result of each command.
	Out io.Writer
}

func newRunsCommand(ports RecordPorts) *cobra.Command {
	runs := &cobra.Command{
		Use:   "runs",
		Short: "List and show the records of past runs",
		Long:  "Every run that is not a dry run leaves a record of the agent, the axes, the guarantees, the overlay digests and the configuration layers. A record holds credential names and never their values.",
	}
	var asJSON bool
	list := &cobra.Command{
		Use:   "list",
		Short: "List the run records, newest first",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			records, err := launch.ListRuns(ports.Records)
			if err != nil {
				return err
			}
			return writeRunList(ports.Out, records, asJSON)
		},
	}
	show := &cobra.Command{
		Use:   "show <id|" + launch.LatestRun + ">",
		Short: "Show one run record",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return launch.RunIDs(ports.Records), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(_ *cobra.Command, args []string) error {
			found, err := launch.ReadRun(ports.Records, args[0])
			if err != nil {
				return err
			}
			return writeRun(ports.Out, found, asJSON)
		},
	}
	list.Flags().BoolVar(&asJSON, "json", false, "Print the records as a JSON array")
	show.Flags().BoolVar(&asJSON, "json", false, "Print the record as JSON")
	runs.AddCommand(list, show)
	return runs
}

func writeRunList(out io.Writer, records []record.Record, asJSON bool) error {
	if asJSON {
		if records == nil {
			records = []record.Record{}
		}
		return writeJSON(out, records)
	}
	if len(records) == 0 {
		_, err := fmt.Fprintln(out, "No runs recorded.")
		return err
	}
	table := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(table, "ID\tSTATUS\tEXIT\tAGENT\tISOLATION\tWORK DIRECTORY")
	for _, r := range records {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\t%s\n", r.ID, r.Status, exitText(r), r.Agent.Name, r.Axes.Isolation, r.WorkDir)
	}
	return table.Flush()
}

func exitText(r record.Record) string {
	if r.ExitCode == nil {
		return "-"
	}
	return fmt.Sprint(*r.ExitCode)
}

func writeRun(out io.Writer, r record.Record, asJSON bool) error {
	if asJSON {
		data, err := record.Marshal(r)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		return err
	}
	var failure error
	line := func(format string, args ...any) {
		if _, err := fmt.Fprintf(out, format+"\n", args...); err != nil {
			failure = errors.Join(failure, err)
		}
	}
	line("Run %s: %s (exit %s)", r.ID, r.Status, exitText(r))
	line("  Started: %s", r.StartedAt.Format("2006-01-02 15:04:05 MST"))
	line("  Ended: %s", r.EndedAt.Format("2006-01-02 15:04:05 MST"))
	if r.Error != "" {
		line("  Error: %s", r.Error)
	}
	line("  Agent: %s (%s)", r.Agent.Name, r.Agent.VersionProbe)
	line("  Isolation: %s, provision: %s, network: %s", r.Axes.Isolation, r.Axes.Provision, r.Axes.Network)
	line("  Workspace: %s", r.Axes.Workspace.Mode)
	line("  Work directory: %s", r.WorkDir)
	for _, guarantee := range r.Guarantees {
		line("  Guarantee %s: required %t, provided %t", guarantee.Name, guarantee.Required, guarantee.Provided)
	}
	for _, plugin := range r.Overlay.Plugins {
		line("  Plugin %s: %s", plugin.Name, plugin.Digest)
	}
	for _, dir := range r.Overlay.Directories {
		line("  Overlay directory %s: %s", dir.Name, dir.Digest)
	}
	for _, name := range r.Credentials {
		line("  Credential: %s", name)
	}
	return failure
}
