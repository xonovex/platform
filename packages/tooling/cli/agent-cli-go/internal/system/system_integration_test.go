//go:build integration

package system

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func newTestSystem(stdin string) (System, *bytes.Buffer, *bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	return New(strings.NewReader(stdin), &stdout, &stderr), &stdout, &stderr
}

func TestRunAttachesTheStreamsAndReturnsTheExitCode(t *testing.T) {
	s, stdout, stderr := newTestSystem("input")
	dir := t.TempDir()

	code, err := s.Run("/bin/sh", []string{"-c", `read -r line; printf "$line $A $(pwd)"; printf err >&2`}, []string{"A=env"}, dir)
	if err != nil || code != 0 {
		t.Fatalf("Run() = %d, %v", code, err)
	}
	resolved, _ := filepath.EvalSymlinks(dir)
	if stdout.String() != "input env "+resolved || stderr.String() != "err" {
		t.Errorf("stdout, stderr = %q, %q", stdout.String(), stderr.String())
	}

	if code, err := s.Run("/bin/sh", []string{"-c", "exit 7"}, nil, ""); err != nil || code != 7 {
		t.Errorf("Run(exit 7) = %d, %v; want the code without an error", code, err)
	}
	if code, err := s.Run(filepath.Join(dir, "missing"), nil, nil, ""); err == nil || code != 1 {
		t.Errorf("Run(missing) = %d, %v; want 1 and an error", code, err)
	}
}

func TestRunInheritsTheEnvironmentWithANilEnv(t *testing.T) {
	t.Setenv("AGENT_CLI_SYSTEM_TEST", "inherited")
	s, stdout, _ := newTestSystem("")

	if code, err := s.Run("/bin/sh", []string{"-c", "printf %s \"$AGENT_CLI_SYSTEM_TEST\""}, nil, ""); err != nil || code != 0 {
		t.Fatalf("Run() = %d, %v", code, err)
	}
	if stdout.String() != "inherited" {
		t.Errorf("stdout = %q, want the inherited variable", stdout.String())
	}
}

func TestStreamWritesToTheOutputStreams(t *testing.T) {
	s, stdout, stderr := newTestSystem("ignored")

	if err := s.Stream("/bin/sh", []string{"-c", "printf out; printf err >&2"}, t.TempDir()); err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if stdout.String() != "out" || stderr.String() != "err" {
		t.Errorf("stdout, stderr = %q, %q", stdout.String(), stderr.String())
	}
	if err := s.Stream("/bin/sh", []string{"-c", "exit 3"}, ""); err == nil {
		t.Error("Stream(exit 3) error = nil")
	}
}

func TestCaptureTrimsTheOutputAndReportsTheStandardError(t *testing.T) {
	s, stdout, _ := newTestSystem("")

	out, err := s.Capture("/bin/sh", []string{"-c", "printf '  value \\n'"}, "")
	if err != nil || out != "value" || stdout.Len() != 0 {
		t.Errorf("Capture() = %q, %v; stdout = %q", out, err, stdout.String())
	}
	_, err = s.Capture("/bin/sh", []string{"-c", "printf ' reason \\n' >&2; exit 2"}, "")
	if err == nil || !strings.HasPrefix(err.Error(), "/bin/sh -c ") || !strings.HasSuffix(err.Error(), ": reason") {
		t.Errorf("Capture(failure) error = %v, want the command line and the standard error", err)
	}
}

func TestFileOperations(t *testing.T) {
	s, _, _ := newTestSystem("")
	t.Setenv("TMPDIR", t.TempDir())
	dir := t.TempDir()

	nested := filepath.Join(dir, "a", "b")
	if err := s.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if info, err := s.Stat(nested); err != nil || !info.IsDir() {
		t.Fatalf("Stat() = %v, %v", info, err)
	}

	path, err := s.CreateTemp("system-*.txt")
	if err != nil || !strings.HasPrefix(path, os.Getenv("TMPDIR")) || !strings.HasSuffix(path, ".txt") {
		t.Fatalf("CreateTemp() = %q, %v", path, err)
	}
	if err := s.WriteFile(path, []byte("content"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if data, err := s.ReadFile(path); err != nil || string(data) != "content" {
		t.Errorf("ReadFile() = %q, %v", data, err)
	}
	if err := s.Remove(path); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if _, err := s.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Stat(removed) error = %v", err)
	}

	t.Setenv("TMPDIR", filepath.Join(dir, "missing"))
	if _, err := s.CreateTemp("x-*"); err == nil {
		t.Error("CreateTemp(missing directory) error = nil")
	}

	link := filepath.Join(dir, "link")
	if err := os.Symlink(nested, link); err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}
	resolvedNested, _ := filepath.EvalSymlinks(nested)
	if got, err := s.EvalSymlinks(link); err != nil || got != resolvedNested {
		t.Errorf("EvalSymlinks() = %q, %v; want %q", got, err, resolvedNested)
	}
}

func TestProcessEnvironmentAndIdentity(t *testing.T) {
	s, _, _ := newTestSystem("")
	t.Setenv("AGENT_CLI_SYSTEM_TEST", "value")

	if s.Getenv("AGENT_CLI_SYSTEM_TEST") != "value" || !slices.Contains(s.Environ(), "AGENT_CLI_SYSTEM_TEST=value") {
		t.Error("Getenv or Environ does not read the process environment")
	}
	if wd, err := s.Getwd(); err != nil || !filepath.IsAbs(wd) {
		t.Errorf("Getwd() = %q, %v", wd, err)
	}
	if home, err := s.UserHomeDir(); err != nil || home == "" {
		t.Errorf("UserHomeDir() = %q, %v", home, err)
	}
	if s.Getuid() != os.Getuid() || s.Getgid() != os.Getgid() || s.NumCPU() < 1 {
		t.Errorf("identity = %d %d %d", s.Getuid(), s.Getgid(), s.NumCPU())
	}
	if path, err := s.LookPath("sh"); err != nil || !filepath.IsAbs(path) {
		t.Errorf("LookPath(sh) = %q, %v", path, err)
	}
	if _, err := s.LookPath("agent-cli-no-such-tool"); err == nil {
		t.Error("LookPath(missing) error = nil")
	}
}

func TestFileOperationsWorkOnTheRealDisk(t *testing.T) {
	s, _, _ := newTestSystem("")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "b"), []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("b", filepath.Join(root, "a")); err != nil {
		t.Fatal(err)
	}

	entries, err := s.ReadDir(root)
	if err != nil || len(entries) != 2 || entries[0].Name() != "a" {
		t.Fatalf("ReadDir() = %v, %v", entries, err)
	}
	if info, err := s.Lstat(filepath.Join(root, "a")); err != nil || info.Mode()&fs.ModeSymlink == 0 {
		t.Errorf("Lstat(link) = %v, %v", info, err)
	}

	dir, err := s.MkdirTemp("agent-cli-system-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if info, err := os.Stat(dir); err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 || !strings.Contains(filepath.Base(dir), "agent-cli-system-") {
		t.Errorf("MkdirTemp() = %q, %v", dir, err)
	}
	moved := filepath.Join(root, "moved")
	if err := s.Rename(dir, moved); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if err := s.RemoveAll(moved); err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}
	if _, err := os.Stat(moved); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("RemoveAll() left %s", moved)
	}
	if err := s.RemoveAll(moved); err != nil {
		t.Errorf("RemoveAll(missing) error = %v", err)
	}
}

func TestIsPublicAddress(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8": true, "2606:4700:4700::1111": true,
		"127.0.0.1": false, "::1": false, "10.0.0.1": false, "192.168.1.1": false, "172.16.0.1": false,
		"169.254.169.254": false, "224.0.0.1": false, "0.0.0.0": false, "fe80::1": false, "fc00::1": false,
	}
	for text, want := range cases {
		if got := IsPublicAddress(net.ParseIP(text)); got != want {
			t.Errorf("IsPublicAddress(%s) = %v, want %v", text, got, want)
		}
	}
}

func TestListenAndDialConnectOnLoopback(t *testing.T) {
	s, _, _ := newTestSystem("")
	listener, err := s.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_, _ = conn.Write([]byte("hello"))
			_ = conn.Close()
		}
	}()
	conn, err := s.DialContext(context.Background(), "tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data, _ := io.ReadAll(conn)
	if string(data) != "hello" {
		t.Errorf("read %q", data)
	}
}

func TestDialPublicRefusesALoopbackService(t *testing.T) {
	s, _, _ := newTestSystem("")
	listener, err := s.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	for _, address := range []string{listener.Addr().String(), "localhost:" + portOf(t, listener)} {
		if _, err := s.DialPublic(context.Background(), "tcp", address); !errors.Is(err, ErrNonPublicAddress) {
			t.Errorf("DialPublic(%s) = %v, want ErrNonPublicAddress", address, err)
		}
	}
}

func portOf(t *testing.T, listener net.Listener) string {
	t.Helper()
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	return port
}

func TestExecutableReturnsTheTestBinary(t *testing.T) {
	s, _, _ := newTestSystem("")
	path, err := s.Executable()
	if err != nil || !filepath.IsAbs(path) {
		t.Errorf("Executable() = %q, %v", path, err)
	}
}

func TestNowAndTheManagedConfigDirectory(t *testing.T) {
	s, _, _ := newTestSystem("")

	if drift := time.Since(s.Now()); drift < -time.Minute || drift > time.Minute {
		t.Errorf("Now() is %v away from the clock", drift)
	}
	dir := s.ManagedConfigDir()
	switch runtime.GOOS {
	case "darwin":
		if dir != "/Library/Application Support/agent-cli" {
			t.Errorf("ManagedConfigDir() = %q", dir)
		}
	case "windows":
		if filepath.Base(dir) != "agent-cli" {
			t.Errorf("ManagedConfigDir() = %q", dir)
		}
	default:
		if dir != "/etc/agent-cli" {
			t.Errorf("ManagedConfigDir() = %q", dir)
		}
	}
}

func TestReadLineReadsOneLineAtATime(t *testing.T) {
	s, _, _ := newTestSystem("yes\r\nno\nlast")
	for _, want := range []string{"yes", "no", "last"} {
		got, err := s.ReadLine()
		if err != nil || got != want {
			t.Fatalf("ReadLine() = (%q, %v), want %q", got, err, want)
		}
	}
	if got, err := s.ReadLine(); !errors.Is(err, io.EOF) || got != "" {
		t.Errorf("ReadLine() after the input = (%q, %v), want io.EOF", got, err)
	}
}

func TestReadLineWithoutInputIsEndOfInput(t *testing.T) {
	if _, err := New(nil, io.Discard, io.Discard).ReadLine(); !errors.Is(err, io.EOF) {
		t.Errorf("ReadLine() error = %v, want io.EOF", err)
	}
}

func TestIsTerminalRefusesPipesFilesAndTheNullDevice(t *testing.T) {
	if New(strings.NewReader(""), io.Discard, io.Discard).IsTerminal() {
		t.Error("a string input is a terminal")
	}
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = null.Close() }()
	if New(null, null, null).IsTerminal() {
		t.Error("the null device is a terminal")
	}
	file, err := os.Open(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	if New(file, file, file).IsTerminal() {
		t.Error("a regular file is a terminal")
	}
}
