// Package cmd builds the command tree of agent-validate. It knows the check
// port, the plan and the report, and reaches the disk and git only through the
// Open port that the composition root binds.
package cmd

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"
	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/jsonschema"
	contractreport "github.com/xonovex/platform/packages/library/shared-contract-go/pkg/report"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/project"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/report"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/runner"
	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/tree"
)

// Environment is what a run reads: the files below a root and the git state.
type Environment struct {
	Tree       tree.Tree
	Repository check.Repository
}

// Opener builds the environment for a root directory.
type Opener func(root string) (Environment, error)

// Dependencies are the ports of the command tree.
type Dependencies struct {
	Open      Opener
	Validator jsonschema.Validator
	Checks    []check.Check
	Out       io.Writer
	Version   string
}

// ExitError carries the exit code of the process. Err is empty when the
// findings already say all there is to say.
type ExitError struct {
	Code int
	Err  error
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit %d", e.Code)
	}
	return e.Err.Error()
}

// Unwrap returns the cause.
func (e ExitError) Unwrap() error { return e.Err }

// ExitCode maps the error of a command to the exit code of the process. An
// error that carries no code is a usage error, such as a misspelled command.
func ExitCode(err error) int {
	var exit ExitError
	switch {
	case err == nil:
		return findings.ExitClean
	case errors.As(err, &exit):
		return exit.Code
	default:
		return findings.ExitUsage
	}
}

const (
	toolName       = "agent-validate"
	informationURI = "https://github.com/xonovex/xonovex-platform/tree/main/packages/tooling/cli/agent-validate-go"
)

type flags struct {
	root    string
	project string
	format  string
	strict  bool
	paths   []string
}

// NewRootCommand builds the command tree.
func NewRootCommand(d Dependencies) *cobra.Command {
	var f flags
	root := &cobra.Command{
		Use:   toolName + " [check...]",
		Short: "Validate skills, plugins, instructions and repository files",
		Long: "Run the checks of the project file, or the named checks, and report what they find.\n\n" +
			"Exit codes: 0 when nothing fails, 1 when a check finds a failure or drift (or a warning with --strict), 2 on a usage error or an infrastructure fault.",
		Args:          cobra.ArbitraryArgs,
		Version:       d.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, d, f, args)
		},
	}
	root.SetOut(d.Out)
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return ExitError{Code: findings.ExitUsage, Err: err} })
	root.PersistentFlags().StringVar(&f.root, "root", ".", "repository root to validate")
	root.Flags().StringVar(&f.project, "project", project.File, "project file, relative to the root")
	root.Flags().StringVar(&f.format, "format", string(contractreport.Text), "output format: "+contractreport.Names(report.Formats()...))
	root.Flags().BoolVar(&f.strict, "strict", false, "fail on warnings too")
	root.Flags().StringArrayVar(&f.paths, "path", nil, "path to check instead of the paths of the project file, repeatable")
	root.AddCommand(listCommand(d))
	root.AddCommand(seedCommand(d, &f.root))
	return root
}

func usage(err error) error { return ExitError{Code: findings.ExitUsage, Err: err} }

func run(cmd *cobra.Command, d Dependencies, f flags, names []string) error {
	format, err := contractreport.Parse(f.format, report.Formats()...)
	if err != nil {
		return usage(err)
	}
	meta := report.Meta{Tool: toolName, Version: d.Version, URI: informationURI}
	fail := func(cause error) error {
		if writeErr := report.WriteFailure(cmd.OutOrStdout(), format, meta, cause); writeErr != nil {
			cause = errors.Join(cause, writeErr)
		}
		return usage(cause)
	}
	env, err := d.Open(f.root)
	if err != nil {
		return fail(err)
	}
	proj, err := project.Load(env.Tree, f.project, d.Validator, cmd.Flags().Changed("project"))
	if err != nil {
		return fail(err)
	}
	env.Tree.Skip = slices.Concat(env.Tree.Skip, proj.Skip)
	steps, err := project.Plan(d.Checks, proj, names, f.paths)
	if err != nil {
		return fail(err)
	}
	all, err := runner.Run(check.Context{Tree: env.Tree, Validator: d.Validator, Repository: env.Repository}, steps)
	if err != nil {
		return fail(err)
	}
	ran := make([]string, len(steps))
	for i, s := range steps {
		ran[i] = s.Check.Name()
	}
	if err := report.Write(cmd.OutOrStdout(), format, meta, ran, all); err != nil {
		return usage(err)
	}
	policy := findings.DefaultPolicy()
	if f.strict {
		policy = findings.StrictPolicy()
	}
	if code := findings.ExitCode(all, nil, policy); code != findings.ExitClean {
		return ExitError{Code: code}
	}
	return nil
}

func listCommand(d Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the checks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			checks := slices.Clone(d.Checks)
			sort.Slice(checks, func(i, j int) bool { return checks[i].Name() < checks[j].Name() })
			for _, c := range checks {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", c.Name(), c.Summary()); err != nil {
					return usage(err)
				}
			}
			return nil
		},
	}
}
