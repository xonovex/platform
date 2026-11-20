// Package herdrcore is the terminal leaf that runs the agent in a terminal of the
// herdr-core runtime. It speaks to the runtime through the hc command only,
// never to the app window or to its socket. It starts the agent from a private
// launch script, so no environment value appears on an hc command line or in the
// process list, and unless the run is detached it waits for the program to end
// and returns its exit code. It starts a terminal only when the CLI runs in one,
// so it never reaches a runtime the CLI does not run in.
package herdrcore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

// Name is the wrapper name of the leaf.
const Name termshared.TerminalType = "herdr-core"

const (
	// command is the herdr-core command line client.
	command = "hc"
	// insideVariable holds the address of the terminal in every process that
	// runs in a herdr-core terminal.
	insideVariable = "HC_TERMINAL_ID"
	// runLabel marks each terminal with the run record id, so the terminals of
	// agent-cli are found with hc ls --label.
	runLabel = "agent-cli.run"
	// namePrefix starts the name of each terminal the leaf starts.
	namePrefix = "agent-cli-"
	// nameTaken is the error kind of hc run for a name a live terminal holds.
	nameTaken = "name_taken"
	// interruptedCode is the exit code of a run the user interrupted, as a shell
	// reports a command that an interrupt ended.
	interruptedCode = 130
	// endWaitSeconds bounds the wait for a killed program before its terminal is
	// removed.
	endWaitSeconds = "10"
	// firstPause and longestPause bound the pause between waits on a program
	// that stopped working but runs on, because such a wait returns at once.
	firstPause   = 500 * time.Millisecond
	longestPause = 10 * time.Second
)

// ErrOutside marks a run that does not run in a herdr-core terminal, so there is
// no terminal to start the agent beside and no runtime the CLI may use.
var ErrOutside = errors.New("agent-cli does not run inside a herdr-core terminal; start it from a herdr-core terminal or use another terminal")

// Host is the port the herdr-core executor consumes: the hc command it runs,
// the interrupt it honours, the environment and clock it reads, and the launch
// script it writes.
type Host interface {
	LookPath(file string) (string, error)
	Getenv(key string) string
	Now() time.Time
	CreateTemp(pattern string) (string, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Remove(path string) error
	// Output runs name until it exits or ctx ends, and returns its trimmed
	// standard output and standard error and its exit code; a non-zero exit is
	// a result, not an error.
	Output(ctx context.Context, name string, args []string, dir string) (stdout, stderr string, code int, err error)
	// NotifyInterrupt returns a context that ends on an interrupt of the CLI.
	NotifyInterrupt(parent context.Context) (context.Context, context.CancelFunc)
	// Sleep pauses for d, or until ctx ends.
	Sleep(ctx context.Context, d time.Duration) error
}

// Executor implements the herdr-core terminal wrapper.
type Executor struct {
	host Host
	log  logging.Logger
}

var _ termshared.TerminalExecutor = (*Executor)(nil)

// NewExecutor creates a herdr-core executor over host, writing through log.
func NewExecutor(host Host, log logging.Logger) *Executor {
	return &Executor{host: host, log: log}
}

// IsAvailable checks that hc is installed.
func (e *Executor) IsAvailable() bool {
	_, err := e.host.LookPath(command)
	return err == nil
}

// IsInside checks that the CLI runs in a herdr-core terminal.
func (e *Executor) IsInside() bool { return e.host.Getenv(insideVariable) != "" }

// Execute starts the command in a new terminal, a child of the terminal the CLI
// runs in. A detached run returns once the terminal has started, and its
// terminal outlives the CLI and removes itself when the program exits. Any other
// run waits for the program, removes its terminal and returns its exit code; an
// interrupt of the CLI ends the program and removes its terminal. Every failure
// removes the launch script, and ends the terminal when it may have started. The
// outcome reports the program ended only when hc confirmed it, so a detached run
// and a program that did not end in time report that it runs on. The session
// and window names of the configuration apply to tmux only.
func (e *Executor) Execute(config *termshared.TerminalConfig, command []string, env []string, workDir string) (termshared.Outcome, error) {
	if !e.IsInside() {
		return termshared.Outcome{Code: 1, Ended: true}, ErrOutside
	}
	script, err := termshared.LaunchScript(e.host, command, env)
	if err != nil {
		return termshared.Outcome{Code: 1, Ended: true}, err
	}
	ctx, stop := e.host.NotifyInterrupt(context.Background())
	defer stop()

	address, ended, err := e.start(ctx, config, script, workDir)
	if err != nil {
		_ = e.host.Remove(script)
		return termshared.Outcome{Code: 1, Ended: ended}, err
	}
	if config.Detach {
		e.log.Info("Started the agent in herdr-core terminal " + address)
		return termshared.Outcome{}, nil
	}
	e.log.Info("Running the agent in herdr-core terminal " + address)
	code, err := e.wait(ctx, address)
	switch {
	case ctx.Err() != nil:
		_ = e.host.Remove(script)
		e.log.Warning("Interrupted: ending herdr-core terminal " + address)
		return termshared.Outcome{Code: interruptedCode, Ended: e.end(address)}, nil
	case err != nil:
		_ = e.host.Remove(script)
		return termshared.Outcome{Code: 1, Ended: e.end(address)}, err
	}
	e.remove(address)
	return termshared.Outcome{Code: code, Ended: true}, nil
}

// start runs hc run in the background and returns the address of the new
// terminal. A name that a live terminal holds is replaced once by a new name.
// An interrupt during the start ends the terminal by its name, because it may
// have started. On a failure, the flag reports that no program of the start
// runs.
func (e *Executor) start(ctx context.Context, config *termshared.TerminalConfig, script, workDir string) (string, bool, error) {
	for attempt := 1; ; attempt++ {
		name := e.name(attempt)
		stdout, stderr, code, err := e.hc(ctx, runArgs(config, name, script, workDir)...)
		if err != nil {
			return "", ctx.Err() == nil || e.end(name), fmt.Errorf("hc run: %w", err)
		}
		if code != 0 {
			failure := parseFailure(stdout, stderr, code)
			if failure.kind == nameTaken && attempt == 1 {
				e.log.Debug("The herdr-core terminal name " + name + " is taken; starting under a new name")
				continue
			}
			return "", true, fmt.Errorf("hc run: %s", failure.text)
		}
		address, err := parseAddress(stdout)
		if err != nil {
			return "", e.end(name), err
		}
		return address, false, nil
	}
}

// name returns the terminal name of a start attempt: unique by the start
// instant, and by the attempt when an earlier name was taken. It stays within
// the 32 characters of a-z, 0-9, - and _ that hc accepts.
func (e *Executor) name(attempt int) string {
	name := namePrefix + strconv.FormatInt(e.host.Now().UnixNano(), 36)
	if attempt > 1 {
		name += "-" + strconv.Itoa(attempt)
	}
	return name
}

// runArgs returns the arguments of hc run. A detached terminal outlives the CLI
// and removes itself once its program exits. An attached terminal sends no
// notes into the terminal of the CLI, because the CLI waits for it itself.
func runArgs(config *termshared.TerminalConfig, name, script, workDir string) []string {
	args := []string{"run", "--bg", "--json", "--name", name}
	if workDir != "" {
		args = append(args, "--cwd", workDir)
	}
	if config.RunID != "" {
		args = append(args, "--label", runLabel+"="+config.RunID)
	}
	if config.Detach {
		args = append(args, "--rm", "--outlive-parent")
	} else {
		args = append(args, "--no-notify")
	}
	return append(args, "--", "sh", script)
}

// wait waits until the program of the terminal exits and returns its exit code.
// A wait that checks in after its timeout waits again. A wait that ends because
// the program stopped working while it runs on, idle, done, blocked or error
// alike, asks whether the terminal has finished and otherwise waits again after
// a pause that grows while nothing changes, because the next wait returns at
// once.
func (e *Executor) wait(ctx context.Context, address string) (int, error) {
	var last waitResult
	pause := time.Duration(0)
	for {
		stdout, stderr, code, err := e.hc(ctx, "wait", "--json", address)
		if err != nil {
			return 1, fmt.Errorf("hc wait %s: %w", address, err)
		}
		result, err := parseWait(stdout)
		if err != nil {
			return 1, fmt.Errorf("hc wait %s: %s", address, parseFailure(stdout, stderr, code).text)
		}
		if result.Reason == "exit" {
			return exitCode(address, result.ExitCode, result.Signal)
		}
		if !result.Ended {
			last, pause = result, 0
			continue
		}
		state, err := e.get(ctx, address)
		if err != nil {
			return 1, err
		}
		if state.State == "finished" {
			return exitCode(address, state.ExitCode, state.Signal)
		}
		pause = nextPause(pause, result.unchanged(last))
		last = result
		if err := e.host.Sleep(ctx, pause); err != nil {
			return 1, err
		}
	}
}

// nextPause doubles the pause while the program reports the same state, from
// firstPause up to longestPause, and starts at none when the state changed.
func nextPause(pause time.Duration, unchanged bool) time.Duration {
	if !unchanged {
		return 0
	}
	return min(max(2*pause, firstPause), longestPause)
}

// get returns the state of the terminal.
func (e *Executor) get(ctx context.Context, address string) (terminalState, error) {
	stdout, stderr, code, err := e.hc(ctx, "get", "--json", address)
	if err != nil {
		return terminalState{}, fmt.Errorf("hc get %s: %w", address, err)
	}
	if code != 0 {
		return terminalState{}, fmt.Errorf("hc get %s: %s", address, parseFailure(stdout, stderr, code).text)
	}
	return parseTerminal(stdout)
}

// end ends the program of the terminal, waits a bounded time for it to finish
// and removes the terminal. It reports whether hc confirmed that the program
// exited. It runs after the interrupt, so it does not use the context the
// interrupt ended.
func (e *Executor) end(target string) bool {
	ctx := context.Background()
	if _, stderr, code, err := e.hc(ctx, "kill", "--json", target); err != nil || code != 0 {
		e.log.Debug(fmt.Sprintf("hc kill %s: %v %s", target, err, stderr))
	}
	stdout, stderr, _, err := e.hc(ctx, "wait", "--json", "--timeout", endWaitSeconds, target)
	result, parseErr := parseWait(stdout)
	exited := err == nil && parseErr == nil && result.Reason == "exit"
	if !exited {
		e.log.Debug(fmt.Sprintf("hc wait %s: %v %s", target, err, stderr))
	}
	e.remove(target)
	return exited
}

// remove removes the finished terminal, and tells the user how to remove it
// when that fails.
func (e *Executor) remove(target string) {
	stdout, stderr, code, err := e.hc(context.Background(), "rm", "--json", target)
	if err == nil && code == 0 {
		return
	}
	reason := parseFailure(stdout, stderr, code).text
	if err != nil {
		reason = err.Error()
	}
	e.log.Warning(fmt.Sprintf("Could not remove herdr-core terminal %s (%s); remove it with hc kill %s and hc rm %s", target, reason, target, target))
}

func (e *Executor) hc(ctx context.Context, args ...string) (string, string, int, error) {
	e.log.Debug(command + " " + strings.Join(args, " "))
	return e.host.Output(ctx, command, args, "")
}

// exitCode returns the exit code of a program that exited, or 128 plus the
// signal that ended it, as a shell reports it.
func exitCode(address string, code, signal *int) (int, error) {
	switch {
	case code != nil:
		return *code, nil
	case signal != nil:
		return 128 + *signal, nil
	default:
		return 1, fmt.Errorf("herdr-core terminal %s ended without an exit code", address)
	}
}
