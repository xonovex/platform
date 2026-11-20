package cmd

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/spf13/cobra"

	libplugin "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/plugin"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/trust"
	origin "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/plugin"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/runconfig"
)

// PluginHost is the machine the plugin and marketplace commands consult.
type PluginHost interface {
	plugin.Files
	// Host reaches the approval ledger and the terminal of the trust gate.
	runconfig.Host
}

// PluginPorts are the ports of the plugin and marketplace commands.
type PluginPorts struct {
	Host    PluginHost
	Origins origin.Registry
	// Out receives the result of each command.
	Out io.Writer
}

// service builds the plugin service of the store the user named, or of the
// default store.
func (p PluginPorts) service(store string) (plugin.Service, error) {
	cwd, err := p.Host.Getwd()
	if err != nil {
		return plugin.Service{}, fmt.Errorf("resolve current working directory: %w", err)
	}
	dir := store
	switch {
	case dir == "":
		home, err := p.Host.UserHomeDir()
		if err != nil {
			return plugin.Service{}, fmt.Errorf("resolve user home directory: %w", err)
		}
		dir = plugin.DefaultDir(p.Host.Getenv, home)
	case !path.IsAbs(dir):
		dir = path.Join(cwd, dir)
	}
	dirStore, err := plugin.NewStore(p.Host, path.Clean(dir))
	if err != nil {
		return plugin.Service{}, err
	}
	return plugin.Service{Store: dirStore, Files: p.Host, Origins: p.Origins}, nil
}

// newPluginCommand builds `agent-cli plugin`.
func newPluginCommand(ports PluginPorts) *cobra.Command {
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
		newPluginInstallCommand(ports, &store),
		newPluginListCommand(ports, &store),
		newPluginRemoveCommand(ports, &store),
	)
	return command
}

func newPluginInstallCommand(ports PluginPorts, store *string) *cobra.Command {
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
		RunE: func(command *cobra.Command, args []string) error {
			capabilities := make([]libplugin.Capability, 0, len(allow))
			for _, name := range allow {
				capability, err := libplugin.ParseCapability(name)
				if err != nil {
					return err
				}
				capabilities = append(capabilities, capability)
			}
			service, err := ports.service(*store)
			if err != nil {
				return err
			}
			fetched, err := service.Fetch(args[0])
			if err != nil {
				return err
			}
			printPreview(ports.Out, fetched)
			subject := trust.Subject{
				Digest: fetched.Digest,
				Path:   fetched.Source,
				Notice: fmt.Sprintf("The plugin %s from %s is installed only after you approve its digest", fetched.Manifest.Name, fetched.Source),
			}
			if err := trust.Decide(subject, approveDigest, runconfig.Ledger(ports.Host), ports.Host, ports.Out); err != nil {
				return err
			}
			record, err := service.Commit(fetched, capabilities)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(ports.Out, "Installed %s %s (%s)\n", record.Name, record.Version, record.Digest)
			return err
		},
	}
	command.Flags().StringVar(&approveDigest, trust.ApproveFlag, "", "Digest of the reviewed plugin files (sha256:...); approves without a question")
	command.Flags().StringSliceVar(&allow, "allow", nil, "Grant a capability the plugin carries: "+strings.Join(capabilityNames(), ", ")+" (repeatable)")
	_ = command.RegisterFlagCompletionFunc("allow", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return capabilityNames(), cobra.ShellCompDirectiveNoFileComp
	})
	return command
}

func newPluginListCommand(ports PluginPorts, store *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the installed plugins",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			service, err := ports.service(*store)
			if err != nil {
				return err
			}
			records, err := service.List()
			if err != nil {
				return err
			}
			if len(records) == 0 {
				_, err = fmt.Fprintln(ports.Out, "No plugins installed.")
				return err
			}
			for _, record := range records {
				line := fmt.Sprintf("%s %s %s %s", record.Name, record.Version, record.Digest, record.Source)
				if record.Marketplace != "" {
					line += " (marketplace " + record.Marketplace + ")"
				}
				if _, err := fmt.Fprintln(ports.Out, line); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newPluginRemoveCommand(ports PluginPorts, store *string) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			service, err := ports.service(*store)
			if err != nil {
				return err
			}
			if err := service.Remove(args[0]); err != nil {
				return err
			}
			_, err = fmt.Fprintf(ports.Out, "Removed %s\n", args[0])
			return err
		},
	}
}

// newMarketplaceCommand builds `agent-cli marketplace`.
func newMarketplaceCommand(ports PluginPorts) *cobra.Command {
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
				service, err := ports.service(store)
				if err != nil {
					return err
				}
				source, listing, err := service.AddMarketplace(args[0])
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(ports.Out, "Registered marketplace %s (%s) with %d plugins\n", source.Name, source.Source, len(listing.Entries))
				return err
			},
		},
		&cobra.Command{
			Use:   "list [name]",
			Short: "List the registered marketplaces, or the plugins of one",
			Args:  cobra.MaximumNArgs(1),
			RunE: func(_ *cobra.Command, args []string) error {
				service, err := ports.service(store)
				if err != nil {
					return err
				}
				if len(args) == 1 {
					return printListing(ports.Out, service, args[0])
				}
				sources, err := service.Marketplaces()
				if err != nil {
					return err
				}
				if len(sources) == 0 {
					_, err = fmt.Fprintln(ports.Out, "No marketplaces registered.")
					return err
				}
				for _, source := range sources {
					if _, err := fmt.Fprintf(ports.Out, "%s %s\n", source.Name, source.Source); err != nil {
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
				service, err := ports.service(store)
				if err != nil {
					return err
				}
				if err := service.RemoveMarketplace(args[0]); err != nil {
					return err
				}
				_, err = fmt.Fprintf(ports.Out, "Removed marketplace %s\n", args[0])
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

func capabilityNames() []string {
	var names []string
	for _, capability := range libplugin.Capabilities() {
		names = append(names, string(capability))
	}
	return names
}
