// Command agent-cli is the composition root of the CLI: it builds the
// operating-system adapter and the logger, wires every plugin to them, and
// hands the result to the command tree.
package main

import (
	"errors"
	"os"

	wsp "github.com/xonovex/platform/packages/library/shared-agent-go/pkg/workspace"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/network/managed"
	originplugins "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/overlay"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/record"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/herdr"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/tmux"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/git"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/jj"
	wsshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/workspace/shared"
)

func main() {
	host := system.New(os.Stdin, os.Stdout, os.Stderr)
	log := logging.FromEnvironment(host.Getenv, os.Stdout, os.Stderr)
	origins := originplugins.DefaultRegistry(host)
	deps := launch.Dependencies{
		Host:        host,
		Log:         log,
		Out:         os.Stdout,
		Sandboxes:   plugins.DefaultRegistry(host, log),
		Egress:      managed.Service{Host: host, Log: log},
		Credentials: plugins.DefaultCredentials(host),
		Overlays:    overlay.Builder{Host: host, Origins: origins},
		Workspaces: map[wsp.VCSType]wsshared.VCS{
			wsp.VCSGit:     git.New(host, log),
			wsp.VCSJujutsu: jj.New(host, log),
		},
		Terminals: map[termshared.TerminalType]termshared.TerminalExecutor{
			termshared.TerminalTmux:  tmux.NewExecutor(host, log),
			termshared.TerminalHerdr: herdr.NewExecutor(host, log),
		},
	}
	var records cmd.RecordReader
	if store, err := runStore(host); err != nil {
		log.Warning("Run records are off: " + err.Error())
	} else {
		deps.Records = store
		records = store
	}

	ports := cmd.PluginPorts{Host: host, Origins: origins, Out: os.Stdout}
	err := cmd.NewRootCommand(deps, ports, cmd.RecordPorts{Records: records, Out: os.Stdout}, managed.Bridge{Host: host}).Execute()
	var exit launch.ExitError
	var usage cmd.UsageError
	switch {
	case errors.As(err, &exit):
		os.Exit(exit.Code)
	case errors.As(err, &usage):
		log.Error(err.Error())
		os.Exit(cmd.ExitUsage)
	case err != nil:
		log.Error(err.Error())
		os.Exit(1)
	}
}

// runStore opens the run record store in the state directory of the user.
func runStore(host system.System) (record.Store, error) {
	home, err := host.UserHomeDir()
	if err != nil {
		return record.Store{}, err
	}
	return record.NewStore(host, record.DefaultDir(host.Getenv, home))
}
