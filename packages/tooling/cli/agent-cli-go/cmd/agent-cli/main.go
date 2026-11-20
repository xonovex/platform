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
	originplugins "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/origin/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/overlay"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/sandbox/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/system"
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
		Host:      host,
		Log:       log,
		Out:       os.Stdout,
		Sandboxes: plugins.DefaultRegistry(host, log),
		Overlays:  overlay.Builder{Host: host, Origins: origins},
		Workspaces: map[wsp.VCSType]wsshared.VCS{
			wsp.VCSGit:     git.New(host, log),
			wsp.VCSJujutsu: jj.New(host, log),
		},
		Terminals: map[termshared.TerminalType]termshared.TerminalExecutor{
			termshared.TerminalTmux: tmux.NewExecutor(host, log),
		},
	}

	ports := cmd.PluginPorts{Host: host, Origins: origins, Out: os.Stdout}
	err := cmd.NewRootCommand(deps, ports).Execute()
	var exit launch.ExitError
	switch {
	case errors.As(err, &exit):
		os.Exit(exit.Code)
	case err != nil:
		log.Error(err.Error())
		os.Exit(1)
	}
}
