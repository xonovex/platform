// Package oshosttest provides Fake, an in-memory double of oshost.System for
// unit tests: a file tree with symlinks, an environment, a user identity and
// recorded process outcomes. It touches no real file, process or variable.
package oshosttest

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)

// Outcome is the recorded result of one command line.
type Outcome struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Call is one process the code under test started.
type Call struct {
	Kind string // "run", "stream", "capture" or "output"
	Name string
	Args []string
	Env  []string
	Dir  string
}

// Line returns the command line of the call.
func (c Call) Line() string { return CommandLine(c.Name, c.Args...) }

type entry struct {
	dir  bool
	mode fs.FileMode
	data []byte
	link string
}

// Fake is an in-memory operating system. The zero value is unusable; use New.
type Fake struct {
	Env        map[string]string
	WorkingDir string
	Home       string
	HomeErr    error
	UID        int
	GID        int
	CPUs       int
	// Outcomes maps a command line, optionally prefixed by "<dir>: ", to its
	// result. A command with no outcome fails to start.
	Outcomes map[string]Outcome
	Calls    []Call

	// Dial and DialPublicFunc answer DialContext and DialPublic; ListenErr, when
	// set, fails Listen; ExecutablePath answers Executable.
	Dial           DialFunc
	DialPublicFunc DialFunc
	ListenErr      error
	ExecutablePath string
	// Clock is the instant Now returns, and OS the name GOOS returns.
	Clock time.Time
	OS    string

	// InterruptOn is the command line whose Output call delivers an interrupt:
	// the context of NotifyInterrupt ends while that command runs. Sleeps holds
	// each pause Sleep was asked for.
	InterruptOn string
	Sleeps      []time.Duration

	// Terminal is the answer of IsTerminal. Input holds the lines ReadLine
	// returns in order, then io.EOF; InputErr, when set, replaces that EOF.
	Terminal bool
	Input    []string
	InputErr error

	files     map[string]entry
	tempCount int
	interrupt context.CancelFunc

	netMu     sync.Mutex
	listeners map[string]*MemListener
}

// New returns a Linux Fake with an empty root directory, uid and gid 1000, 4
// CPUs, home /home/user and working directory /work.
func New() *Fake {
	return &Fake{
		Env:        map[string]string{},
		WorkingDir: "/work",
		Home:       "/home/user",
		UID:        1000,
		GID:        1000,
		CPUs:       4,
		Clock:      time.Date(2026, time.November, 20, 12, 0, 0, 0, time.UTC),
		OS:         "linux",
		Outcomes:   map[string]Outcome{},
		files:      map[string]entry{"/": {dir: true, mode: fs.ModeDir | 0o755}},
	}
}

// CommandLine joins a program and its arguments with spaces, the key of Outcomes.
func CommandLine(name string, args ...string) string {
	return strings.Join(append([]string{name}, args...), " ")
}

// InDir prefixes a command line with the directory it must run in.
func InDir(dir, line string) string { return dir + ": " + line }

// AddDir creates the directory p and every missing parent.
func (f *Fake) AddDir(p string) *Fake {
	p = path.Clean(p)
	for _, parent := range ancestors(p) {
		if _, ok := f.files[parent]; !ok {
			f.files[parent] = entry{dir: true, mode: fs.ModeDir | 0o755}
		}
	}
	return f
}

// AddFile creates the file p with the given content and permissions, and every
// missing parent directory.
func (f *Fake) AddFile(p string, data string, perm fs.FileMode) *Fake {
	p = path.Clean(p)
	f.AddDir(path.Dir(p))
	f.files[p] = entry{mode: perm, data: []byte(data)}
	return f
}

// AddExecutable creates an executable file p.
func (f *Fake) AddExecutable(p string) *Fake { return f.AddFile(p, "#!/bin/sh\n", 0o755) }

// AddLink creates the symlink p that points at target, and every missing
// parent directory. A relative target resolves against the directory of p.
func (f *Fake) AddLink(p, target string) *Fake {
	p = path.Clean(p)
	f.AddDir(path.Dir(p))
	f.files[p] = entry{mode: fs.ModeSymlink | 0o777, link: target}
	return f
}

// File returns the content of the file p and whether it exists.
func (f *Fake) File(p string) (string, bool) {
	e, ok := f.files[path.Clean(p)]
	if !ok || e.dir {
		return "", false
	}
	return string(e.data), true
}

// Exists reports whether p names an entry, without following a final symlink.
func (f *Fake) Exists(p string) bool {
	_, ok := f.files[path.Clean(p)]
	return ok
}

// Getenv returns the value of key in Env.
func (f *Fake) Getenv(key string) string { return f.Env[key] }

// IsTerminal reports the Terminal field.
func (f *Fake) IsTerminal() bool { return f.Terminal }

// ReadLine returns the next line of Input. When Input is used up it returns
// InputErr, or io.EOF when that is nil.
func (f *Fake) ReadLine() (string, error) {
	if len(f.Input) == 0 {
		if f.InputErr != nil {
			return "", f.InputErr
		}
		return "", io.EOF
	}
	line := f.Input[0]
	f.Input = f.Input[1:]
	return line, nil
}

// Environ returns Env as sorted KEY=VALUE entries.
func (f *Fake) Environ() []string {
	out := make([]string, 0, len(f.Env))
	for k, v := range f.Env {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

// Getwd returns WorkingDir.
func (f *Fake) Getwd() (string, error) { return f.WorkingDir, nil }

// UserHomeDir returns Home, or HomeErr when it is set.
func (f *Fake) UserHomeDir() (string, error) {
	if f.HomeErr != nil {
		return "", f.HomeErr
	}
	return f.Home, nil
}

// Getuid returns UID.
func (f *Fake) Getuid() int { return f.UID }

// Getgid returns GID.
func (f *Fake) Getgid() int { return f.GID }

// Now returns Clock.
func (f *Fake) Now() time.Time { return f.Clock }

// GOOS returns OS.
func (f *Fake) GOOS() string { return f.OS }

// NumCPU returns CPUs.
func (f *Fake) NumCPU() int { return f.CPUs }

// EvalSymlinks resolves every symlink in the absolute path p.
func (f *Fake) EvalSymlinks(p string) (string, error) {
	resolved, err := f.resolve(p, 0)
	if err != nil {
		return "", &fs.PathError{Op: "lstat", Path: p, Err: err}
	}
	return resolved, nil
}

// Stat returns the information of p after following symlinks.
func (f *Fake) Stat(p string) (fs.FileInfo, error) {
	resolved, err := f.resolve(p, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: p, Err: err}
	}
	e := f.files[resolved]
	return fileInfo{name: path.Base(p), entry: e}, nil
}

// Lstat returns the information of p without following a final symlink.
func (f *Fake) Lstat(p string) (fs.FileInfo, error) {
	p = path.Clean(p)
	parent, err := f.resolve(path.Dir(p), 0)
	if err != nil {
		return nil, &fs.PathError{Op: "lstat", Path: p, Err: err}
	}
	e, ok := f.files[path.Join(parent, path.Base(p))]
	if !ok {
		return nil, &fs.PathError{Op: "lstat", Path: p, Err: fs.ErrNotExist}
	}
	return fileInfo{name: path.Base(p), entry: e}, nil
}

// ReadDir returns the entries of the directory p, sorted by name.
func (f *Fake) ReadDir(p string) ([]fs.DirEntry, error) {
	resolved, err := f.resolve(p, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: p, Err: err}
	}
	if !f.files[resolved].dir {
		return nil, &fs.PathError{Op: "readdir", Path: p, Err: fmt.Errorf("not a directory")}
	}
	var entries []fs.DirEntry
	for name, e := range f.files {
		if name != resolved && path.Dir(name) == resolved {
			entries = append(entries, dirEntry{fileInfo{name: path.Base(name), entry: e}})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, nil
}

// ReadFile returns the content of the file p.
func (f *Fake) ReadFile(p string) ([]byte, error) {
	resolved, err := f.resolve(p, 0)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: p, Err: err}
	}
	e := f.files[resolved]
	if e.dir {
		return nil, &fs.PathError{Op: "read", Path: p, Err: fmt.Errorf("is a directory")}
	}
	return append([]byte(nil), e.data...), nil
}

// WriteFile replaces the content of the existing or new file p.
func (f *Fake) WriteFile(p string, data []byte, perm fs.FileMode) error {
	parent, err := f.resolve(path.Dir(p), 0)
	if err != nil || !f.files[parent].dir {
		return &fs.PathError{Op: "open", Path: p, Err: fs.ErrNotExist}
	}
	f.files[path.Join(parent, path.Base(p))] = entry{mode: perm, data: append([]byte(nil), data...)}
	return nil
}

// CreateTemp creates an empty private file under /tmp whose name replaces the
// last "*" of pattern with a counter.
func (f *Fake) CreateTemp(pattern string) (string, error) {
	f.tempCount++
	name := pattern
	if i := strings.LastIndex(pattern, "*"); i >= 0 {
		name = fmt.Sprintf("%s%d%s", pattern[:i], f.tempCount, pattern[i+1:])
	}
	p := path.Join("/tmp", name)
	f.AddFile(p, "", 0o600)
	return p, nil
}

// MkdirTemp creates an empty directory under /tmp whose name replaces the last
// "*" of pattern with a counter.
func (f *Fake) MkdirTemp(pattern string) (string, error) {
	f.tempCount++
	name := pattern + fmt.Sprint(f.tempCount)
	if i := strings.LastIndex(pattern, "*"); i >= 0 {
		name = fmt.Sprintf("%s%d%s", pattern[:i], f.tempCount, pattern[i+1:])
	}
	p := path.Join("/tmp", name)
	f.AddDir(p)
	return p, nil
}

// Rename moves the entry oldPath, and everything below it, to newPath. The parent
// of newPath must exist.
func (f *Fake) Rename(oldPath, newPath string) error {
	oldPath, newPath = path.Clean(oldPath), path.Clean(newPath)
	if _, ok := f.files[oldPath]; !ok {
		return &fs.PathError{Op: "rename", Path: oldPath, Err: fs.ErrNotExist}
	}
	if parent, ok := f.files[path.Dir(newPath)]; !ok || !parent.dir {
		return &fs.PathError{Op: "rename", Path: newPath, Err: fs.ErrNotExist}
	}
	moved := map[string]entry{}
	for name, e := range f.files {
		if name == oldPath || strings.HasPrefix(name, oldPath+"/") {
			moved[newPath+strings.TrimPrefix(name, oldPath)] = e
			delete(f.files, name)
		}
	}
	for name, e := range moved {
		f.files[name] = e
	}
	return nil
}

// RemoveAll deletes the entry p and everything below it. A missing entry is not
// an error.
func (f *Fake) RemoveAll(p string) error {
	p = path.Clean(p)
	for name := range f.files {
		if name == p || strings.HasPrefix(name, p+"/") {
			delete(f.files, name)
		}
	}
	return nil
}

// Remove deletes the entry p.
func (f *Fake) Remove(p string) error {
	p = path.Clean(p)
	if _, ok := f.files[p]; !ok {
		return &fs.PathError{Op: "remove", Path: p, Err: fs.ErrNotExist}
	}
	delete(f.files, p)
	return nil
}

// MkdirAll creates the directory p and every missing parent.
func (f *Fake) MkdirAll(p string, _ fs.FileMode) error {
	f.AddDir(p)
	return nil
}

// LookPath resolves file through the PATH entry of Env, as exec.LookPath does.
func (f *Fake) LookPath(file string) (string, error) {
	if strings.Contains(file, "/") {
		if f.isExecutable(file) {
			return file, nil
		}
		return "", fmt.Errorf("exec: %q: executable file not found", file)
	}
	for _, dir := range strings.Split(f.Env["PATH"], ":") {
		if dir == "" {
			continue
		}
		candidate := path.Join(dir, file)
		if f.isExecutable(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("exec: %q: executable file not found in $PATH", file)
}

// Run records the call and returns the exit code of its outcome.
func (f *Fake) Run(name string, args, env []string, dir string) (int, error) {
	outcome, err := f.record(Call{Kind: "run", Name: name, Args: args, Env: env, Dir: dir})
	if err != nil {
		return 1, err
	}
	return outcome.ExitCode, nil
}

// Stream records the call and returns the error of its outcome.
func (f *Fake) Stream(name string, args []string, dir string) error {
	_, err := f.record(Call{Kind: "stream", Name: name, Args: args, Dir: dir})
	return err
}

// Capture records the call and returns the standard output of its outcome.
func (f *Fake) Capture(name string, args []string, dir string) (string, error) {
	outcome, err := f.record(Call{Kind: "capture", Name: name, Args: args, Dir: dir})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(outcome.Stdout), nil
}

// Output records the call and returns the outputs and the exit code of its
// outcome. A call while ctx has ended, or the call InterruptOn names, returns
// the error of ctx, as a process that the context ended does.
func (f *Fake) Output(ctx context.Context, name string, args []string, dir string) (string, string, int, error) {
	outcome, err := f.record(Call{Kind: "output", Name: name, Args: args, Dir: dir})
	if f.interrupt != nil && f.InterruptOn != "" && CommandLine(name, args...) == f.InterruptOn {
		f.interrupt()
	}
	stdout, stderr := strings.TrimSpace(outcome.Stdout), strings.TrimSpace(outcome.Stderr)
	if ctx.Err() != nil {
		return stdout, stderr, 1, ctx.Err()
	}
	if err != nil {
		return "", "", 1, err
	}
	return stdout, stderr, outcome.ExitCode, nil
}

// NotifyInterrupt returns a context that ends when the command InterruptOn
// names runs, and its cancel function.
func (f *Fake) NotifyInterrupt(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	f.interrupt = cancel
	return ctx, cancel
}

// Sleep records the pause and returns at once, with the error of ctx when it
// has ended.
func (f *Fake) Sleep(ctx context.Context, d time.Duration) error {
	f.Sleeps = append(f.Sleeps, d)
	return ctx.Err()
}

func (f *Fake) record(call Call) (Outcome, error) {
	f.Calls = append(f.Calls, call)
	line := call.Line()
	outcome, ok := f.Outcomes[InDir(call.Dir, line)]
	if !ok {
		outcome, ok = f.Outcomes[line]
	}
	if !ok {
		return Outcome{}, fmt.Errorf("oshosttest: no outcome for %q in %q", line, call.Dir)
	}
	return outcome, outcome.Err
}

func (f *Fake) isExecutable(p string) bool {
	info, err := f.Stat(p)
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

// resolve follows every symlink in the absolute path p and returns the path of
// the entry it names.
func (f *Fake) resolve(p string, depth int) (string, error) {
	if depth > 40 {
		return "", fmt.Errorf("too many levels of symbolic links")
	}
	if !path.IsAbs(p) {
		return "", fmt.Errorf("oshosttest: relative path %q", p)
	}
	current := "/"
	parts := strings.Split(strings.TrimPrefix(path.Clean(p), "/"), "/")
	for i, part := range parts {
		if part == "" {
			continue
		}
		next := path.Join(current, part)
		e, ok := f.files[next]
		if !ok {
			return "", fs.ErrNotExist
		}
		if e.link != "" {
			target := e.link
			if !path.IsAbs(target) {
				target = path.Join(current, target)
			}
			resolved, err := f.resolve(target, depth+1)
			if err != nil {
				return "", err
			}
			next = resolved
			e = f.files[next]
		}
		if !e.dir && i < len(parts)-1 {
			return "", fmt.Errorf("not a directory")
		}
		current = next
	}
	return current, nil
}

func ancestors(p string) []string {
	var out []string
	for current := p; ; current = path.Dir(current) {
		out = append([]string{current}, out...)
		if current == "/" || current == "." {
			return out
		}
	}
}

type fileInfo struct {
	name  string
	entry entry
}

func (i fileInfo) Name() string { return i.name }
func (i fileInfo) Size() int64  { return int64(len(i.entry.data)) }
func (i fileInfo) Mode() fs.FileMode {
	if i.entry.dir {
		return fs.ModeDir | i.entry.mode.Perm()
	}
	return i.entry.mode
}
func (i fileInfo) ModTime() time.Time { return time.Time{} }
func (i fileInfo) IsDir() bool        { return i.entry.dir }
func (i fileInfo) Sys() any           { return nil }

type dirEntry struct{ info fileInfo }

func (d dirEntry) Name() string               { return d.info.name }
func (d dirEntry) IsDir() bool                { return d.info.IsDir() }
func (d dirEntry) Type() fs.FileMode          { return d.info.Mode().Type() }
func (d dirEntry) Info() (fs.FileInfo, error) { return d.info, nil }
