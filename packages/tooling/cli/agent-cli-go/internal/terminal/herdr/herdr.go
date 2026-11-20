// Package herdr is the terminal leaf that runs the agent in a pane of herdr, the
// session the CLI itself runs in. It splits the current pane and types a command
// into the new pane that runs a private launch script, so no environment value
// appears in a herdr command line or in the process list. It never attaches to a
// herdr session the CLI does not run in.
package herdr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/logging"
	"github.com/xonovex/platform/packages/library/shared-core-go/pkg/shell"
	termshared "github.com/xonovex/platform/packages/tooling/cli/agent-cli-go/internal/terminal/shared"
)

// insideVariable is set to 1 in every process that runs in a herdr pane.
const insideVariable = "HERDR_ENV"

// ErrOutside marks a run that does not run in a herdr pane, so there is no pane
// to split and no session the CLI may use.
var ErrOutside = errors.New("agent-cli does not run inside a herdr pane; start it from a herdr pane or use another terminal")

// Host is the port the herdr executor consumes: the herdr command it runs, the
// environment it reads, and the launch script it writes.
type Host interface {
	LookPath(file string) (string, error)
	Getenv(key string) string
	// Capture runs name in dir and returns its trimmed standard output.
	Capture(name string, args []string, dir string) (string, error)
	CreateTemp(pattern string) (string, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Remove(path string) error
}

// Executor implements the herdr terminal wrapper.
type Executor struct {
	host Host
	log  logging.Logger
}

var _ termshared.TerminalExecutor = (*Executor)(nil)

// NewExecutor creates a herdr executor over host, writing through log.
func NewExecutor(host Host, log logging.Logger) *Executor {
	return &Executor{host: host, log: log}
}

// IsAvailable checks that herdr is installed.
func (e *Executor) IsAvailable() bool {
	_, err := e.host.LookPath("herdr")
	return err == nil
}

// IsInside checks that the CLI runs in a herdr pane.
func (e *Executor) IsInside() bool { return e.host.Getenv(insideVariable) == "1" }

// Execute opens a pane beside the current one and starts the command in it. The
// command runs on in that pane, so Execute reports success once herdr has
// accepted the command and does not wait for the agent. A pane that cannot be
// given its command is closed again, so a failure leaves no empty pane behind.
// The session and window names of the configuration apply to tmux only.
func (e *Executor) Execute(config *termshared.TerminalConfig, command []string, env []string, workDir string) (int, error) {
	if !e.IsInside() {
		return 1, ErrOutside
	}
	script, err := termshared.LaunchScript(e.host, command, env)
	if err != nil {
		return 1, err
	}
	pane, err := e.split(config, workDir)
	if err != nil {
		_ = e.host.Remove(script)
		return 1, err
	}
	if _, err := e.herdr("pane", "run", pane, "sh "+shell.Quote(script)); err != nil {
		_ = e.host.Remove(script)
		_, _ = e.herdr("pane", "close", pane)
		return 1, err
	}
	e.log.Info("Started the agent in herdr pane " + pane)
	return 0, nil
}

// split opens a pane beside the current one without taking the focus of the user
// when the run is detached, and returns its identifier.
func (e *Executor) split(config *termshared.TerminalConfig, workDir string) (string, error) {
	args := []string{"pane", "split", "--current", "--direction", "right"}
	if workDir != "" {
		args = append(args, "--cwd", workDir)
	}
	if config.Detach {
		args = append(args, "--no-focus")
	}
	output, err := e.herdr(args...)
	if err != nil {
		return "", err
	}
	var reply struct {
		Result struct {
			Pane struct {
				ID string `json:"pane_id"`
			} `json:"pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(output), &reply); err != nil || reply.Result.Pane.ID == "" {
		return "", fmt.Errorf("herdr pane split named no pane: %s", output)
	}
	return reply.Result.Pane.ID, nil
}

func (e *Executor) herdr(args ...string) (string, error) {
	e.log.Debug("herdr " + strings.Join(args, " "))
	output, err := e.host.Capture("herdr", args, "")
	if err != nil {
		return "", fmt.Errorf("herdr %s: %w", args[0]+" "+args[1], err)
	}
	return output, nil
}
