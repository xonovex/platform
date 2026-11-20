// Package osexec is the process adapter that starts real programs. Only the
// composition root imports it.
package osexec

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/process"
)

// maxLine is the longest line of standard output the adapter reads. A harness
// streams whole messages as one JSON line.
const maxLine = 16 << 20

// stopGrace is how long a stopped program has to exit before it is killed.
const stopGrace = 2 * time.Second

// Runner starts programs on the host.
type Runner struct{}

// New returns a Runner.
func New() Runner { return Runner{} }

// Run starts the program and feeds each line of its standard output to the
// handler. A handler that returns true, or a timeout, ends the program.
func (Runner) Run(ctx context.Context, request process.Request, onLine process.LineHandler) (process.Result, error) {
	var runCtx context.Context
	var cancel context.CancelFunc
	if request.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, request.Timeout)
	} else {
		runCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	cmd := exec.CommandContext(runCtx, request.Name, request.Args...)
	cmd.Dir = request.Dir
	cmd.Env = environment(request.Env)
	cmd.WaitDelay = stopGrace
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return process.Result{}, err
	}
	if err := cmd.Start(); err != nil {
		return process.Result{}, err
	}

	stopped := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64<<10), maxLine)
	for scanner.Scan() {
		if onLine != nil && onLine(scanner.Text()) {
			stopped = true
			cancel()
			break
		}
	}
	// Drain what remains so the program is never blocked on a full pipe.
	for scanner.Scan() {
	}
	waitErr := cmd.Wait()

	result := process.Result{Stderr: stderr.String(), Stopped: stopped}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) && !stopped {
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
	if stopped && errors.Is(waitErr, context.Canceled) {
		// The program ended by itself before the stop reached it: Wait reports the
		// cancelled context for a clean exit, and the stop is no fault.
		return result, nil
	}
	return result, waitErr
}

// environment merges the overrides into the inherited environment.
func environment(overrides map[string]string) []string {
	merged := map[string]string{}
	for _, pair := range os.Environ() {
		if key, value, ok := strings.Cut(pair, "="); ok {
			merged[key] = value
		}
	}
	for key, value := range overrides {
		merged[key] = value
	}
	out := make([]string, 0, len(merged))
	for key, value := range merged {
		out = append(out, key+"="+value)
	}
	sort.Strings(out)
	return out
}
