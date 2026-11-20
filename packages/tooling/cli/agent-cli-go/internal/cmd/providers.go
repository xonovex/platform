package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/approval"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/providers"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/trust"
)

// approveFlag is the flag of `providers approve` that carries the reviewed digest.
const approveFlag = "--approve-digest"

// runApproveFlag is the flag of `run` that carries the reviewed digest.
const runApproveFlag = "--approve-project-digest"

// trustStore opens the approval ledger of the user.
func trustStore(ports PluginPorts) (trust.Store, error) {
	home, err := ports.Host.UserHomeDir()
	if err != nil {
		return trust.Store{}, fmt.Errorf("resolve user home directory: %w", err)
	}
	return trust.NewStore(ports.Host, trust.DefaultFile(ports.Host.Getenv, home))
}

// approveMaterial shows the project provider material, asks the user to approve
// its digest, and records the approval. A digest from the flag approves without a
// question, and without a flag and a terminal it fails and names the digest.
func approveMaterial(ports PluginPorts, material providers.Material, flagged, flag string) error {
	digest := material.Digest()
	printMaterial(ports.Out, material)
	approved, err := trust.Decide(ports.Host, ports.Out, digest, flagged)
	if err != nil {
		return err
	}
	if err := approval.Check(digest, approved, flag); err != nil {
		return err
	}
	store, err := trustStore(ports)
	if err != nil {
		return err
	}
	cwd, err := ports.Host.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current working directory: %w", err)
	}
	return store.Approve(approval.Approved{Kind: launch.ApprovalKind, Subject: cwd, Digest: digest})
}

// printMaterial shows what the user is asked to approve: each provider file with
// the hosts its route reaches and its model slots, the model overrides, and the
// digest.
func printMaterial(out io.Writer, material providers.Material) {
	lines := []string{"Project provider material:"}
	for _, description := range material.Describe() {
		lines = append(lines, "  File: "+description.Origin, fmt.Sprintf("    Provider: %s (agent %s)", description.Name, description.Agent))
		if description.Extends != "" {
			lines = append(lines, "    Extends: "+description.Extends)
		}
		lines = append(lines, "    Route hosts: "+orNone(description.Hosts), "    Model slots: "+orNone(description.Slots))
	}
	if len(material.Models) > 0 {
		lines = append(lines, "  Model overrides: "+strings.Join(material.Models, ", "))
	}
	lines = append(lines, "Digest: "+material.Digest())
	// A failed write to the output stream does not change what is approved.
	_, _ = fmt.Fprintln(out, strings.Join(lines, "\n"))
}

func orNone(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

// executeWithApproval runs the run and, when it stops because the project
// provider material is not approved, shows the material, asks for the approval
// and runs once more.
func executeWithApproval(deps launch.Dependencies, ports PluginPorts, options launch.Options, flagged string) error {
	err := launch.Execute(deps, options)
	var needed *launch.ApprovalRequiredError
	if !errors.As(err, &needed) || ports.Host == nil {
		return err
	}
	if err := approveMaterial(ports, needed.Material, flagged, runApproveFlag); err != nil {
		return err
	}
	return launch.Execute(deps, options)
}

// newProvidersCommand builds `agent-cli providers`.
func newProvidersCommand(deps launch.Dependencies, ports PluginPorts) *cobra.Command {
	command := &cobra.Command{
		Use:   "providers",
		Short: "List the defined model providers and approve the ones a project supplies",
		Long: `A model provider is a JSON file that names an agent, the credential and route it
uses, and its model slots. agent-cli ships no provider and no model id: you define
each one in a providers directory next to your configuration file, in the
providers directory of the managed configuration, or in .agent-cli/providers of a
project. A project file or a project model override takes effect only after you
approve its digest, and any change to the content asks again.`,
	}
	command.AddCommand(newProvidersListCommand(deps), newProvidersApproveCommand(deps, ports))
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
		lines := []string{"No provider is defined. Add a <name>.json file to one of these directories:"}
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

func newProvidersApproveCommand(deps launch.Dependencies, ports PluginPorts) *cobra.Command {
	options := launch.Options{}
	var flagged string
	command := &cobra.Command{
		Use:   "approve",
		Short: "Approve the provider files and model overrides of the project",
		Long: `Show the provider files and model overrides that the project supplies, with the
hosts each route reaches and each model slot, and record the digest of that
content after you approve it. On a terminal the command asks you to approve the
digest. Without a terminal it stops; review the content, then repeat the command
with --approve-digest <digest>. Any change to a file or an override changes the
digest and asks again.`,
		Args: cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			material, err := launch.ProjectMaterial(deps, options)
			if err != nil {
				return err
			}
			if material.Empty() {
				_, err := fmt.Fprintln(ports.Out, "The project supplies no provider files and no model overrides.")
				return err
			}
			if err := approveMaterial(ports, material, flagged, approveFlag); err != nil {
				return err
			}
			_, err = fmt.Fprintf(ports.Out, "Approved %s\n", material.Digest())
			return err
		},
	}
	bindProviderLocation(command, &options)
	command.Flags().StringVar(&flagged, "approve-digest", "", "Digest of the reviewed content (sha256:...); approves without a question")
	return command
}

// bindProviderLocation binds the flags that locate the configuration and the
// project, which the provider commands share with `run`.
func bindProviderLocation(command *cobra.Command, options *launch.Options) {
	command.Flags().StringVarP(&options.Config, "config", "c", "", "Configuration file")
	command.Flags().StringVarP(&options.WorkDir, "work-dir", "w", "", "Project directory")
}

// definedProviders names the providers defined for an agent type, for
// completion. A failure to read them offers none.
func definedProviders(deps launch.Dependencies, options launch.Options, agent string) []string {
	listing, err := launch.ListProviders(deps, options)
	if err != nil {
		return nil
	}
	var names []string
	for _, row := range listing.Rows {
		if string(row.Agent) == agent {
			names = append(names, row.Name)
		}
	}
	return names
}
