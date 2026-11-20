// Package osfiles binds the configuration ports to the host: the disk, the
// directories of the platform and the terminal. Only a composition root
// imports it.
package osfiles

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
)

// Files reads and writes the real disk.
type Files struct{}

// ReadFile returns the content of a file.
func (Files) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

// Stat describes a path, following a symbolic link.
func (Files) Stat(path string) (discover.Info, error) {
	info, err := os.Stat(path)
	if err != nil {
		return discover.Info{}, err
	}
	return discover.Info{Mode: info.Mode()}, nil
}

// WriteFile replaces a file whole: it writes a private temporary file beside it
// and renames it over the target, so a reader never sees a partial file.
func (Files) WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(temp.Name()) }()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

// Places names the directories of this host.
type Places struct{}

func host() discover.Host {
	home, _ := os.UserHomeDir()
	return discover.Host{GOOS: runtime.GOOS, Getenv: os.Getenv, Home: home}
}

// ManagedDir is the directory of the managed file of this host.
func (Places) ManagedDir() string { return host().ManagedDir() }

// UserDir is the directory of the user file.
func (Places) UserDir() (string, error) { return host().UserDir() }

// StateDir is the state directory of the user.
func (Places) StateDir() (string, error) { return host().StateDir() }

// Getwd is the working directory.
func (Places) Getwd() (string, error) { return os.Getwd() }

// Prompter reads answers from the standard input.
type Prompter struct {
	in       *bufio.Reader
	terminal bool
}

// NewPrompter returns a Prompter over the standard input.
func NewPrompter() *Prompter {
	info, err := os.Stdin.Stat()
	terminal := err == nil && info.Mode()&os.ModeCharDevice != 0
	return &Prompter{in: bufio.NewReader(os.Stdin), terminal: terminal}
}

// IsTerminal reports whether the standard input is a terminal.
func (p *Prompter) IsTerminal() bool { return p.terminal }

// ReadLine returns one line without its line ending. A last line without a
// newline is still a line.
func (p *Prompter) ReadLine() (string, error) {
	line, err := p.in.ReadString('\n')
	if err == io.EOF && line != "" {
		err = nil
	}
	return strings.TrimRight(line, "\r\n"), err
}
