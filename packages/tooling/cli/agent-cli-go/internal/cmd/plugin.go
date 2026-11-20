package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
)

// newPluginCommand builds `agent-cli plugin`.
func newPluginCommand(plugins plugin.Manager) *cobra.Command {
	var store string
	command := &cobra.Command{
		Use:   "plugin",
		Short: "Install, list and remove plugins that agent-cli overlays on a run",
		Long: `Plugins hold skills, agents, commands and instructions. An installed plugin is
stored by the digest of its files, recorded in a lock file with its source, and
laid over a run with agent-cli run --plugin <name>. An install fails closed: it
needs the digest you reviewed, and a grant for each hook, MCP server, language
server or settings file the plugin carries.`,
	}
	command.PersistentFlags().StringVar(&store, "store", "", "Plugin store directory (default: $XDG_DATA_HOME/agentic/plugins)")
	command.AddCommand(
		newPluginInstallCommand(plugins, &store),
		newPluginListCommand(plugins, &store),
		newPluginRemoveCommand(plugins, &store),
	)
	return command
}

func newPluginInstallCommand(plugins plugin.Manager, store *string) *cobra.Command {
	var approveDigest string
	var allow []string
	command := &cobra.Command{
		Use:   "install <source>",
		Short: "Install a plugin after you approve its digest",
		Long: `Install the plugin at <source>: a directory, a repository URL with an optional
#branch-or-tag, or name@marketplace for a plugin of a registered marketplace.
The command prints the plugin and its digest. On a terminal it then asks you to
approve that digest and installs only on y or yes. Without a terminal it stops;
review the files, then repeat the command with --approve-config-digest <digest>.
The approval ledger keeps an approved digest, so the same content installs again
without a question. A grant for a hook, MCP server, language server or settings file is never given by
the answer: pass --allow for each.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			request := plugin.InstallRequest{Store: *store, Source: args[0], Allow: allow, ApproveDigest: approveDigest}
			record, err := plugins.Install(request, func(fetched plugin.Fetched) { printPreview(plugins.Out, fetched) })
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(plugins.Out, "Installed %s %s (%s)\n", record.Name, record.Version, record.Digest)
			return err
		},
	}
	command.Flags().StringVar(&approveDigest, trust.ApproveFlag, "", "Digest of the reviewed plugin files (sha256:...); approves without a question")
	command.Flags().StringSliceVar(&allow, "allow", nil, "Grant a capability the plugin carries: "+strings.Join(plugin.CapabilityNames(), ", ")+" (repeatable)")
	_ = command.RegisterFlagCompletionFunc("allow", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return plugin.CapabilityNames(), cobra.ShellCompDirectiveNoFileComp
	})
	return command
}

func newPluginListCommand(plugins plugin.Manager, store *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the installed plugins",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			service, err := plugins.Open(*store)
			if err != nil {
				return err
			}
			records, err := service.List()
			if err != nil {
				return err
			}
			if len(records) == 0 {
				_, err = fmt.Fprintln(plugins.Out, "No plugins installed.")
				return err
			}
			for _, record := range records {
				line := fmt.Sprintf("%s %s %s %s", record.Name, record.Version, record.Digest, record.Source)
				if record.Marketplace != "" {
					line += " (marketplace " + record.Marketplace + ")"
				}
				if _, err := fmt.Fprintln(plugins.Out, line); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newPluginRemoveCommand(plugins plugin.Manager, store *string) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			service, err := plugins.Open(*store)
			if err != nil {
				return err
			}
			if err := service.Remove(args[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintf(plugins.Out, "Removed %s\n", args[0])
			return err
		},
	}
}

// newMarketplaceCommand builds `agent-cli marketplace`.
func newMarketplaceCommand(plugins plugin.Manager) *cobra.Command {
	var store string
	command := &cobra.Command{
		Use:   "marketplace",
		Short: "Register, list and remove plugin marketplaces",
		Long: `A marketplace lists plugins in .claude-plugin/marketplace.json or
.agents/plugins/marketplace.json. A registered marketplace lets you install a
plugin as name@marketplace.`,
	}
	command.PersistentFlags().StringVar(&store, "store", "", "Plugin store directory (default: $XDG_DATA_HOME/agentic/plugins)")
	command.AddCommand(
		&cobra.Command{
			Use:   "add <source>",
			Short: "Register the marketplace at a directory or repository URL",
			Args:  cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				service, err := plugins.Open(store)
				if err != nil {
					return err
				}
				source, listing, err := service.AddMarketplace(args[0])
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(plugins.Out, "Registered marketplace %s (%s) with %d plugins\n", source.Name, source.Source, len(listing.Entries))
				return err
			},
		},
		&cobra.Command{
			Use:   "list [name]",
			Short: "List the registered marketplaces, or the plugins of one",
			Args:  cobra.MaximumNArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				service, err := plugins.Open(store)
				if err != nil {
					return err
				}
				if len(args) == 1 {
					return printListing(plugins.Out, service, args[0])
				}
				sources, err := service.Marketplaces()
				if err != nil {
					return err
				}
				if len(sources) == 0 {
					_, err = fmt.Fprintln(plugins.Out, "No marketplaces registered.")
					return err
				}
				for _, source := range sources {
					if _, err := fmt.Fprintf(plugins.Out, "%s %s\n", source.Name, source.Source); err != nil {
						return err
					}
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "remove <name>",
			Short: "Unregister a marketplace",
			Args:  cobra.ExactArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				service, err := plugins.Open(store)
				if err != nil {
					return err
				}
				if err := service.RemoveMarketplace(args[0]); err != nil {
					return err
				}
				_, err = fmt.Fprintf(plugins.Out, "Removed marketplace %s\n", args[0])
				return err
			},
		},
	)
	return command
}

func printListing(out io.Writer, service plugin.Service, name string) error {
	listing, err := service.Browse(name)
	if err != nil {
		return err
	}
	for _, entry := range listing.Entries {
		line := entry.Name
		if entry.Description != "" {
			line += " - " + entry.Description
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

// printPreview shows what the user is asked to approve. The trust gate shows
// the digest after it.
func printPreview(out io.Writer, fetched plugin.Fetched) {
	lines := []string{
		fmt.Sprintf("Plugin: %s %s", fetched.Manifest.Name, fetched.Manifest.Version),
		"Source: " + fetched.Source,
	}
	if fetched.Resolved != "" {
		lines = append(lines, "Revision: "+fetched.Resolved)
	}
	if fetched.Marketplace != "" {
		lines = append(lines, "Marketplace: "+fetched.Marketplace)
	}
	lines = append(lines, fmt.Sprintf("Files: %d", fetched.FileCount))
	if len(fetched.Manifest.Dependencies) > 0 {
		lines = append(lines, "Depends on: "+strings.Join(fetched.Manifest.Dependencies, ", "))
	}
	if len(fetched.Capabilities) > 0 {
		names := make([]string, len(fetched.Capabilities))
		for index, capability := range fetched.Capabilities {
			names[index] = string(capability)
		}
		lines = append(lines, "Carries: "+strings.Join(names, ", ")+" (each needs --allow)")
	}
	// A failed write to the output stream does not change what is installed.
	_, _ = fmt.Fprintln(out, strings.Join(lines, "\n"))
}
