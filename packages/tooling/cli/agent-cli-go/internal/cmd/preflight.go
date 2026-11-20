package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/launch"
)

// Exit codes of `agent-cli preflight`: 0 when the run could start, ExitNotReady
// when it could not, and ExitUsage for a command line the CLI cannot read.
const (
	ExitNotReady = 1
	ExitUsage    = 2
)

// UsageError marks a command line the CLI cannot read: an unknown flag or an
// argument the command does not take.
type UsageError struct{ Err error }

func (e UsageError) Error() string { return e.Err.Error() }
func (e UsageError) Unwrap() error { return e.Err }

// newPreflightCommand builds `agent-cli preflight`: it resolves the run its flags
// describe, as `run` would, and reports whether the run could start, without
// starting it.
func newPreflightCommand(deps launch.Dependencies) *cobra.Command {
	options := defaultOptions()
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Check that a run could start, without starting it",
		Long: `Resolve the run that the flags and the configuration files describe and report each axis choice and the layer that made it, each guarantee with the reason it holds or fails, the tools, credential stores and plugins the run needs, and where each configuration value came from.
Nothing runs, no credential is read and no overlay is written.
Exit code 0 means the run could start, 1 means it could not, 2 means the command line was wrong.`,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return UsageError{Err: fmt.Errorf("preflight takes no arguments, got %q", args)}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			run := options
			run.Verbose, _ = cmd.Flags().GetBool("verbose")
			run.Explicit = chosenKeys(cmd)
			report := launch.Preflight(deps, run)
			if err := writeReport(deps.Out, report, asJSON); err != nil {
				return err
			}
			if !report.Ready {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return launch.ExitError{Code: ExitNotReady}
			}
			return nil
		},
	}
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return UsageError{Err: err} })
	bindRunFlags(cmd, &options, deps)
	cmd.Flags().BoolVar(&asJSON, "json", false, "Print the report as JSON")
	return cmd
}

func writeReport(out io.Writer, report launch.Report, asJSON bool) error {
	if asJSON {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("encode the preflight report: %w", err)
		}
		_, err = fmt.Fprintf(out, "%s\n", data)
		return err
	}
	return writeReportText(out, report)
}

func writeReportText(out io.Writer, report launch.Report) error {
	var failure error
	line := func(format string, args ...any) {
		if _, err := fmt.Fprintf(out, format+"\n", args...); err != nil {
			failure = errors.Join(failure, err)
		}
	}
	if report.Ready {
		line("Preflight: ready")
	} else {
		line("Preflight: not ready")
	}
	if report.Problem != "" {
		line("  Problem: %s", report.Problem)
	}
	if len(report.Axes) > 0 {
		line("Choices:")
		for _, axis := range report.Axes {
			line("  %-10s %s (%s)", axis.Axis, axis.Choice, axis.Source)
		}
	}
	if len(report.Guarantees) > 0 {
		line("Guarantees:")
		for _, guarantee := range report.Guarantees {
			line("  %-26s %s", guarantee.Guarantee, describeGuarantee(guarantee))
		}
	}
	if len(report.Checks) > 0 {
		line("Checks:")
		for _, check := range report.Checks {
			if check.OK {
				line("  ok    %s", check.Name)
			} else {
				line("  FAIL  %s: %s", check.Name, check.Detail)
			}
		}
	}
	if len(report.Configuration) > 0 {
		line("Configuration:")
		for _, value := range report.Configuration {
			line("  %s", value)
		}
	}
	return failure
}

// describeGuarantee says whether a guarantee holds and whether the run needs it.
func describeGuarantee(guarantee launch.GuaranteeStatus) string {
	switch {
	case guarantee.Provided && guarantee.Required:
		return "required, provided"
	case guarantee.Provided:
		return "provided"
	case guarantee.Required:
		return "required, NOT provided: " + guarantee.Reason
	default:
		return "not provided, not required"
	}
}
