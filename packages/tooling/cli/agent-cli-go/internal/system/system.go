// Package system is the driven adapter over the operating system: the file
// system, the process environment, child processes and the user identity. The
// axis cores and leaves declare the subset they consume as their own port, and
// System satisfies each of them; no other internal package calls os, os/exec
// or runtime.
package system

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
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

// IsTerminal reports whether both the input and the output of System are a
// terminal, so a person is there to answer a question.
func (s System) IsTerminal() bool { return isTerminal(s.stdin) && isTerminal(s.stdout) }

// ReadLine reads one line from the input of System without its line ending. It
// reads a byte at a time, so it consumes nothing beyond the line. Input that
// ends before a line ending returns what it read, or io.EOF if it read nothing.
func (s System) ReadLine() (string, error) {
	if s.stdin == nil {
		return "", io.EOF
	}
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := s.stdin.Read(buf)
		if n == 1 {
			if buf[0] == '\n' {
				return strings.TrimSuffix(string(line), "\r"), nil
			}
			line = append(line, buf[0])
		}
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) > 0 {
				return string(line), nil
			}
			return "", err
		}
	}
}

// isTerminal reports whether stream is a character device other than the null
// device, so a terminal and not a redirection. A character device that is not a
// terminal makes ReadLine return no answer, and the caller declines.
func isTerminal(stream any) bool {
	file, ok := stream.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	null, err := os.Stat(os.DevNull)
	return err != nil || !os.SameFile(info, null)
}

// Getenv returns the value of an environment variable of this process.
func (System) Getenv(key string) string { return os.Getenv(key) }

// Environ returns the environment of this process as KEY=VALUE entries.
func (System) Environ() []string { return os.Environ() }

// Getwd returns the working directory of this process.
func (System) Getwd() (string, error) { return os.Getwd() }

// UserHomeDir returns the home directory of the current user.
func (System) UserHomeDir() (string, error) { return os.UserHomeDir() }

// Now returns the current time.
func (System) Now() time.Time { return time.Now() }

// ManagedConfigDir returns the directory of the system-wide configuration that
// an administrator manages: /etc/agent-cli, or the machine-wide application data
// directory of macOS and Windows.
func (System) ManagedConfigDir() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Library/Application Support/agent-cli"
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "agent-cli")
	default:
		return "/etc/agent-cli"
	}
}

// Getuid returns the user id of this process.
func (System) Getuid() int { return os.Getuid() }

// Getgid returns the group id of this process.
func (System) Getgid() int { return os.Getgid() }

// NumCPU returns the number of logical CPUs of this host.
func (System) NumCPU() int { return runtime.NumCPU() }

// Executable returns the absolute path of the executable that started this
// process, with symlinks resolved.
func (System) Executable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}

// Listen announces on the local network address.
func (System) Listen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

// DialContext connects to address on the named network.
func (System) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	var dialer net.Dialer
	return dialer.DialContext(ctx, network, address)
}

// ErrNonPublicAddress reports a connection to an address that is not a public
// unicast address.
var ErrNonPublicAddress = errors.New("connection to a non-public address refused")

// DialPublic connects to address and refuses every resolved address that is
// loopback, private, link-local, multicast or unspecified, so an allowed host
// name cannot reach a service of this machine or its network.
func (System) DialPublic(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := net.Dialer{Control: func(_, resolved string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(resolved)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil || !IsPublicAddress(ip) {
			return fmt.Errorf("%w: %s", ErrNonPublicAddress, host)
		}
		return nil
	}}
	return dialer.DialContext(ctx, network, address)
}

// IsPublicAddress reports whether ip is a public unicast address.
func IsPublicAddress(ip net.IP) bool {
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast() && !ip.IsUnspecified() && !ip.IsInterfaceLocalMulticast()
}

// Stat returns the file information of path, following symlinks.
func (System) Stat(path string) (fs.FileInfo, error) { return os.Stat(path) }

// EvalSymlinks returns path with every symlink resolved.
func (System) EvalSymlinks(path string) (string, error) { return filepath.EvalSymlinks(path) }

// Lstat returns the file information of path without following a final symlink.
func (System) Lstat(path string) (fs.FileInfo, error) { return os.Lstat(path) }

// ReadDir returns the entries of the directory at path, sorted by name.
func (System) ReadDir(path string) ([]fs.DirEntry, error) { return os.ReadDir(path) }

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

// MkdirTemp creates a private directory in the temporary directory whose name
// follows pattern, and returns its path.
func (System) MkdirTemp(pattern string) (string, error) { return os.MkdirTemp("", pattern) }

// Rename moves oldPath to newPath, replacing a file at newPath.
func (System) Rename(oldPath, newPath string) error { return os.Rename(oldPath, newPath) }

// RemoveAll deletes path and everything below it. A path that does not exist is
// not an error.
func (System) RemoveAll(path string) error { return os.RemoveAll(path) }

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
