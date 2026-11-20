// Package osexec is the process adapter that starts real programs. Only a
// composition root imports it.
package osexec

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/envutil"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/process"
)

// stopGrace is how long a stopped program has to exit before it is killed.
const stopGrace = 2 * time.Second

// Runner starts programs on the host.
type Runner struct{}

// New returns a Runner.
func New() Runner { return Runner{} }

// Run starts the program and feeds each line of its standard output to out. An
// Output that returns true, or a timeout, ends the program.
func (Runner) Run(ctx context.Context, request process.Request, out process.Output) (process.Result, error) {
	var runCtx context.Context
	var cancel context.CancelFunc
	if request.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, request.Timeout)
	} else {
		runCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	lines := process.NewSplitter(out, cancel)
	cmd := exec.CommandContext(runCtx, request.Name, request.Args...)
	cmd.Dir = request.Dir
	cmd.Env = envutil.EnvMapToSlice(envutil.MergeEnvMaps(envutil.ParseEnv(os.Environ()), request.Env))
	// WaitDelay bounds the wait for a pipe that a child of the program still
	// holds open after the stop or the timeout ended the program.
	cmd.WaitDelay = stopGrace
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = lines
	if err := cmd.Start(); err != nil {
		return process.Result{}, err
	}
	waitErr := cmd.Wait()
	_ = lines.Close()

	result := process.Result{Stderr: stderr.String(), Stopped: lines.Stopped()}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) && !lines.Stopped() {
		result.TimedOut = true
		return result, nil
	}
	if ctx.Err() != nil {
		return process.Result{}, ctx.Err()
	}
	var exit *exec.ExitError
	if errors.As(waitErr, &exit) {
		result.ExitCode = exit.ExitCode()
		return result, nil
	}
	if lines.Stopped() && errors.Is(waitErr, context.Canceled) {
		// The program ended by itself before the stop reached it: Wait reports the
		// cancelled context for a clean exit, and the stop is no fault.
		return result, nil
	}
	return result, waitErr
}
