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

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/osfiles"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs/osfs"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/wiring"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process/osexec"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/scratch/osscratch"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/cmd"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/execute"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/predicate"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/registry"
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
	engine := func(facts agents.Catalog) (harness.Executor, *predicate.Registry, error) {
		harnesses, err := registry.Harnesses(facts.Get)
		if err != nil {
			return nil, nil, err
		}
		exec := execute.Executor{
			Harnesses: harnesses, Runner: osexec.New(), Scratch: osscratch.New(), FS: fsys,
			Now:   time.Now,
			Pause: func(attempt int) { time.Sleep(time.Duration(attempt) * retryPause) },
		}
		return exec, registry.Predicates(exec), nil
	}
	deps := cmd.Dependencies{
		FS: fsys, Validator: validator, ConfigFiles: osfiles.Files{}, Places: osfiles.Places{},
		Prompter: osfiles.NewPrompter(), Protocols: registry.Protocols(), Engine: engine,
		Env: environment(), Dir: dir, Now: time.Now, Out: os.Stdout, Err: os.Stderr,
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
