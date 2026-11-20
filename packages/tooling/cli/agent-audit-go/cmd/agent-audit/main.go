// Command agent-audit is the composition root of the audit tool: it builds the
// operating-system adapters, binds them to the transports and hands the result
// to the command tree.
package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/osfiles"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/osexec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/transport/plugins"
	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/web/nethttp"
)

func main() { os.Exit(run()) }

func run() int {
	validator, err := wiring.NewValidator()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return findings.ExitUsage
	}
	fsys := osfs.New()
	runner := osexec.New()
	deps := cmd.Dependencies{
		FS:          fsys,
		ConfigFiles: osfiles.Files{},
		Places:      osfiles.Places{},
		Runner:      runner,
		Transports:  plugins.DefaultRegistry(plugins.Ports{FS: fsys, Runner: runner, Web: nethttp.New()}),
		Validator:   validator,
		LoaderFor:   wiring.NewSourcesLoader,
		Now:         time.Now,
		Out:         os.Stdout,
		Err:         os.Stderr,
	}

	err = cmd.NewRootCommand(deps).Execute()
	var exit cmd.ExitError
	switch {
	case err == nil:
		return findings.ExitClean
	case errors.As(err, &exit):
		return exit.Code
	default:
		fmt.Fprintln(os.Stderr, "error:", err)
		return findings.ExitUsage
	}
}
