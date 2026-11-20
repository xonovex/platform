package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// newProvidersCommand builds `agent-cli providers`.
func newProvidersCommand(deps launch.Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "providers",
		Short: "List the defined model providers and approve the ones a project supplies",
		Long: `A model provider names an agent, the credential and route it uses, and its model
slots. agent-cli ships no provider and no model id: you define each one in the
providers section of an agentic.json file, or as providers.d/<name>.json beside
the managed, the user or the project agentic.json. The run and providers
sections of a project agentic.json and each project provider file take effect
only after you approve the digest of the file, and any change to a file asks
again.`,
	}
	command.AddCommand(
		newProvidersListCommand(deps), newProvidersShowCommand(deps),
		newProvidersApproveCommand(deps), newProvidersRevokeCommand(deps),
	)
	return command
}

func newProvidersListCommand(deps launch.Dependencies) *cobra.Command {
	options := launch.Options{}
	command := &cobra.Command{
		Use:   "list",
		Short: "List the defined providers and the layer each one came from",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			listing, err := launch.ListProviders(deps, options)
			if err != nil {
				return err
			}
			return writeProviders(deps.Out, listing)
		},
	}
	bindProviderLocation(command, &options)
	return command
}

func writeProviders(out io.Writer, listing launch.ProviderListing) error {
	if len(listing.Rows) == 0 {
		lines := []string{"No provider is defined. Add a providers section to an agentic.json file, or a <name>.json file to one of these directories:"}
		for _, dir := range listing.Dirs {
			lines = append(lines, fmt.Sprintf("  %s: %s", dir.Layer, dir.Path))
		}
		_, err := fmt.Fprintln(out, strings.Join(lines, "\n"))
		return err
	}
	for _, row := range listing.Rows {
		line := fmt.Sprintf("%s %s %s %s", row.Name, row.Agent, row.Layer, row.Origin)
		if row.Extends != "" {
			line += " extends " + row.Extends
		}
		if row.Portable {
			line += " portable"
		}
		if row.Approval != "" {
			line += " (" + row.Approval + ")"
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func newProvidersShowCommand(deps launch.Dependencies) *cobra.Command {
	options := launch.Options{}
	command := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a provider: its file, route hosts, model slots and approval",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			listing, err := launch.ListProviders(deps, options)
			if err != nil {
				return err
			}
			rows := listing.Named(args[0])
			if len(rows) == 0 {
				return UsageError{Err: fmt.Errorf("provider %s is not defined; see agent-cli providers list", args[0])}
			}
			for index, row := range rows {
				if index > 0 {
					if _, err := fmt.Fprintln(deps.Out); err != nil {
						return err
					}
				}
				if _, err := fmt.Fprintln(deps.Out, strings.Join(describeProvider(row), "\n")); err != nil {
					return err
				}
			}
			return nil
		},
	}
	bindProviderLocation(command, &options)
	return command
}

// describeProvider is what a person reviews of one provider: where it is
// defined, the hosts its route reaches and its model slots.
func describeProvider(row launch.ProviderRow) []string {
	lines := []string{
		"Name: " + row.Name,
		"Agent: " + string(row.Agent),
		"Layer: " + row.Layer,
		"File: " + row.Origin,
	}
	if row.Extends != "" {
		lines = append(lines, "Extends: "+row.Extends)
	}
	lines = append(lines, "Route hosts: "+orNone(row.RouteHosts()), "Model slots: "+orNone(row.ModelSlots()))
	if row.Approval != "" {
		lines = append(lines, "Approval: "+row.Approval)
	}
	return lines
}

func orNone(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func newProvidersApproveCommand(deps launch.Dependencies) *cobra.Command {
	options := launch.Options{}
	command := &cobra.Command{
		Use:   "approve",
		Short: "Approve the project agentic.json, provider files and harness overrides",
		Long: `Show each project file that a run reads only after approval: the project
agentic.json when it sets the run or providers section, each provider file in
providers.d and each harness override in harnesses.d beside it. On a terminal the command shows each file that is not
approved and asks you to approve its digest. Without a terminal it stops; review
the file, then repeat the command with --approve-config-digest <digest>. Any
change to a file changes its digest and asks again.`,
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			approved, err := launch.ApproveProjectFiles(deps, options)
			if err != nil {
				return err
			}
			if len(approved) == 0 {
				_, err := fmt.Fprintln(deps.Out, "The project supplies no file that needs approval.")
				return err
			}
			for _, p := range approved {
				if _, err := fmt.Fprintf(deps.Out, "Approved %s (%s)\n", p.Subject.Digest, p.Subject.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
	bindProviderLocation(command, &options)
	bindApproval(command, &options)
	return command
}

func newProvidersRevokeCommand(deps launch.Dependencies) *cobra.Command {
	options := launch.Options{}
	command := &cobra.Command{
		Use:   "revoke [digest]",
		Short: "Remove the approval of the project files, or of one digest",
		Long: `Remove from the approval ledger the approval of each project file that a run
reads only after approval, or, with a digest, the approval of that digest. The
next run that reads a file whose approval is gone asks again.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 1 {
				if err := launch.RevokeDigest(deps, args[0]); err != nil {
					return err
				}
				_, err := fmt.Fprintf(deps.Out, "Revoked %s\n", args[0])
				return err
			}
			revoked, err := launch.RevokeProjectFiles(deps, options)
			if err != nil {
				return err
			}
			if len(revoked) == 0 {
				_, err := fmt.Fprintln(deps.Out, "The approval ledger holds no approval of a project file.")
				return err
			}
			for _, p := range revoked {
				if _, err := fmt.Fprintf(deps.Out, "Revoked %s (%s)\n", p.Subject.Digest, p.Subject.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
	bindProviderLocation(command, &options)
	return command
}

// bindProviderLocation binds the flags that locate the project, which the
// provider commands share with `run`.
func bindProviderLocation(command *cobra.Command, options *launch.Options) {
	command.Flags().StringVar(&options.Project, "project", "", "Project agentic.json (default: the nearest agentic.json from the work directory up to the repository root)")
	command.Flags().StringVarP(&options.WorkDir, "work-dir", "w", "", "Project directory")
}

// bindApproval binds the flag that approves project content by its digest.
func bindApproval(command *cobra.Command, options *launch.Options) {
	command.Flags().StringVar(&options.ApproveDigest, trust.ApproveFlag, "", "Approve the project file with this digest (sha256:...) without a question")
}
