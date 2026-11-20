// Package osexec is the process adapter that starts real programs. Only the
// composition root imports it.
package osexec

import (
	"bytes"
	"context"
	"errors"
	"os/exec"

	"github.com/xonovex/platform/packages/tooling/cli/agent-audit-go/internal/process"
)

// Runner starts programs on the host.
type Runner struct{}

// New returns a Runner.
func New() Runner { return Runner{} }

// Run starts the program, waits for it and returns its standard output and exit
// status. Standard error is discarded.
func (Runner) Run(ctx context.Context, command process.Command) (process.Result, error) {
	cmd := exec.CommandContext(ctx, command.Name, command.Args...)
	cmd.Dir = command.Dir
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return process.Result{Stdout: stdout.Bytes(), ExitCode: exit.ExitCode()}, nil
	}
	if err != nil {
		return process.Result{}, err
	}
	return process.Result{Stdout: stdout.Bytes()}, nil
}
