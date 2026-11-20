// Command agent-eval is the composition root of the eval tool: it builds the
// operating-system adapters, binds them to the harness executor and the
// registries, and hands the result to the command tree.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/execute"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process/osexec"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/registry"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/scratch/osscratch"
)

// retryPause is how long the executor waits after a failed attempt, per attempt
// already made: a transient fault of the harness often clears in seconds.
const retryPause = 2 * time.Second

func main() { os.Exit(run()) }

func run() int {
	validator, err := wiring.NewValidator()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return findings.ExitUsage
	}
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return findings.ExitUsage
	}
	fsys := osfs.New()
	exec := execute.Executor{
		Harnesses: registry.Harnesses(), Runner: osexec.New(), Scratch: osscratch.New(), FS: fsys,
		Now:   time.Now,
		Pause: func(attempt int) { time.Sleep(time.Duration(attempt) * retryPause) },
	}
	deps := cmd.Dependencies{
		FS: fsys, Validator: validator, Protocols: registry.Protocols(), Exec: exec,
		Predicates: registry.Predicates(exec), Env: environment(), Dir: dir,
		Now: time.Now, Out: os.Stdout, Err: os.Stderr,
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

// environment returns the environment variables as a map.
func environment() map[string]string {
	env := map[string]string{}
	for _, pair := range os.Environ() {
		if key, value, ok := strings.Cut(pair, "="); ok {
			env[key] = value
		}
	}
	return env
}
