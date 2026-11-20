// Package configtest holds the in-memory doubles that unit tests use in place
// of the disk, the host directories and the terminal.
package configtest

import (
	"fmt"
	"io"
	iofs "io/fs"
	"path/filepath"
	"sync"

	"github.com/xonovex/platform/packages/library/shared-config-go/pkg/discover"
)

// Files is an in-memory file tree. Every parent of a file is a directory.
type Files struct {
	mu    sync.Mutex
	files map[string][]byte
	dirs  map[string]bool
}

// NewFiles returns a tree that holds files, keyed by path.
func NewFiles(files map[string]string) *Files {
	f := &Files{files: map[string][]byte{}, dirs: map[string]bool{}}
	for path, content := range files {
		f.put(path, []byte(content))
	}
	return f
}

func (f *Files) put(path string, data []byte) {
	path = filepath.Clean(path)
	f.files[path] = data
	for dir := filepath.Dir(path); !f.dirs[dir]; dir = filepath.Dir(dir) {
		f.dirs[dir] = true
		if filepath.Dir(dir) == dir {
			break
		}
	}
}

func notExist(op, path string) error {
	return &iofs.PathError{Op: op, Path: path, Err: iofs.ErrNotExist}
}

// ReadFile returns the content of a file.
func (f *Files) ReadFile(path string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, ok := f.files[filepath.Clean(path)]
	if !ok {
		return nil, notExist("read", path)
	}
	return append([]byte(nil), data...), nil
}

// Stat describes a file as a regular file and a parent as a directory.
func (f *Files) Stat(path string) (discover.Info, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path = filepath.Clean(path)
	_, file := f.files[path]
	switch {
	case file:
		return discover.Info{Mode: 0o644}, nil
	case f.dirs[path]:
		return discover.Info{Mode: iofs.ModeDir | 0o755}, nil
	default:
		return discover.Info{}, notExist("stat", path)
	}
}

// WriteFile creates or replaces a file.
func (f *Files) WriteFile(path string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.dirs[filepath.Clean(path)] {
		return fmt.Errorf("write %s: is a directory", path)
	}
	f.put(path, append([]byte{}, data...))
	return nil
}

// Places names fixed directories.
type Places struct {
	Managed, User, State, Wd string
	// Err fails UserDir, StateDir and Getwd.
	Err error
}

// ManagedDir returns Managed.
func (p Places) ManagedDir() string { return p.Managed }

// UserDir returns User.
func (p Places) UserDir() (string, error) { return p.User, p.Err }

// StateDir returns State.
func (p Places) StateDir() (string, error) { return p.State, p.Err }

// Getwd returns Wd.
func (p Places) Getwd() (string, error) { return p.Wd, p.Err }

// Prompter is a terminal that gives scripted answers, then the end of the
// input.
type Prompter struct {
	Terminal bool
	Answers  []string
}

// IsTerminal returns Terminal.
func (p *Prompter) IsTerminal() bool { return p.Terminal }

// ReadLine returns the next answer, or io.EOF when none is left.
func (p *Prompter) ReadLine() (string, error) {
	if len(p.Answers) == 0 {
		return "", io.EOF
	}
	answer := p.Answers[0]
	p.Answers = p.Answers[1:]
	return answer, nil
}
