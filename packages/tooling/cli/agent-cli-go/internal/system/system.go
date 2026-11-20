// Package system is the driven adapter over the operating system: the file
// system, the process environment, child processes and the user identity. The
// axis cores and leaves declare the subset they consume as their own port, and
// System satisfies each of them; no other internal package calls os, os/exec
// or runtime.
package system

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// System launches child processes on the given standard streams.
type System struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// New returns the operating-system adapter over the given standard streams.
func New(stdin io.Reader, stdout, stderr io.Writer) System {
	return System{stdin: stdin, stdout: stdout, stderr: stderr}
}

// Getenv returns the value of an environment variable of this process.
func (System) Getenv(key string) string { return os.Getenv(key) }

// Environ returns the environment of this process as KEY=VALUE entries.
func (System) Environ() []string { return os.Environ() }

// Getwd returns the working directory of this process.
func (System) Getwd() (string, error) { return os.Getwd() }

// UserHomeDir returns the home directory of the current user.
func (System) UserHomeDir() (string, error) { return os.UserHomeDir() }

// Getuid returns the user id of this process.
func (System) Getuid() int { return os.Getuid() }

// Getgid returns the group id of this process.
func (System) Getgid() int { return os.Getgid() }

// NumCPU returns the number of logical CPUs of this host.
func (System) NumCPU() int { return runtime.NumCPU() }

// Stat returns the file information of path, following symlinks.
func (System) Stat(path string) (fs.FileInfo, error) { return os.Stat(path) }

// EvalSymlinks returns path with every symlink resolved.
func (System) EvalSymlinks(path string) (string, error) { return filepath.EvalSymlinks(path) }

// ReadFile returns the content of the file at path.
func (System) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

// WriteFile replaces the content of the file at path.
func (System) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// CreateTemp creates an empty, private file in the temporary directory whose
// name follows pattern, and returns its path.
func (System) CreateTemp(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// Remove deletes the file at path.
func (System) Remove(path string) error { return os.Remove(path) }

// MkdirAll creates the directory at path and every missing parent.
func (System) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

// LookPath resolves an executable name through the PATH of this process.
func (System) LookPath(file string) (string, error) { return exec.LookPath(file) }

// Run starts name attached to the standard streams of System, waits for it and
// returns its exit code. A nil env inherits the environment of this process.
// A process that starts and exits non-zero is a result, not an error.
func (s System) Run(name string, args, env []string, dir string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stdin = s.stdin
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

// Stream runs name in dir with its output on the output streams of System and
// no input.
func (s System) Stream(name string, args []string, dir string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	return cmd.Run()
}

// Capture runs name in dir and returns its trimmed standard output. A failure
// carries the command line and the trimmed standard error.
func (System) Capture(name string, args []string, dir string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}
