package bwrap

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/network"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/provision"
	"github.com/xonovex/platform/packages/library/shared-launch-go/pkg/oshost/oshosttest"
)

// elfImage returns the start of a 64-bit ELF executable: its file header and
// one program header, plus the interpreter path when interpreter is not empty.
func elfImage(order binary.ByteOrder, interpreter string) string {
	header := elf.Header64{Phoff: 64, Phentsize: 56, Phnum: 1}
	copy(header.Ident[:], elf.ELFMAG)
	header.Ident[elf.EI_CLASS] = byte(elf.ELFCLASS64)
	header.Ident[elf.EI_DATA] = byte(elf.ELFDATA2LSB)
	if order == binary.BigEndian {
		header.Ident[elf.EI_DATA] = byte(elf.ELFDATA2MSB)
	}
	program := elf.Prog64{Type: uint32(elf.PT_LOAD)}
	if interpreter != "" {
		program = elf.Prog64{Type: uint32(elf.PT_INTERP), Off: 120, Filesz: uint64(len(interpreter) + 1)}
	}
	var out bytes.Buffer
	_ = binary.Write(&out, order, header)
	_ = binary.Write(&out, order, program)
	if interpreter != "" {
		out.WriteString(interpreter + "\x00")
	}
	return out.String()
}

// elf32Image is elfImage for a 32-bit little-endian executable.
func elf32Image(interpreter string) string {
	header := elf.Header32{Phoff: 52, Phentsize: 32, Phnum: 1}
	copy(header.Ident[:], elf.ELFMAG)
	header.Ident[elf.EI_CLASS] = byte(elf.ELFCLASS32)
	header.Ident[elf.EI_DATA] = byte(elf.ELFDATA2LSB)
	program := elf.Prog32{Type: uint32(elf.PT_INTERP), Off: 84, Filesz: uint32(len(interpreter) + 1)}
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, header)
	_ = binary.Write(&out, binary.LittleEndian, program)
	out.WriteString(interpreter + "\x00")
	return out.String()
}

var staticBinary = elfImage(binary.LittleEndian, "")

// nixHost is a host whose /bin/sh and dynamic loader live in /nix/store, which
// passthrough does not bind, as on NixOS.
func nixHost() *oshosttest.Fake {
	host := newHost()
	host.AddFile("/nix/store/abc-bash/bin/sh", staticBinary, 0o755).AddLink("/bin/sh", "/nix/store/abc-bash/bin/sh")
	host.AddFile("/nix/store/def-glibc/lib/ld-linux-x86-64.so.2", staticBinary, 0o755)
	return host
}

func TestPassthroughRejectsAnInterpreterOutsideTheBinds(t *testing.T) {
	host := nixHost().AddFile("/usr/bin/claude", "#!/bin/sh\nexec agent\n", 0o755)

	_, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{})

	if err == nil {
		t.Fatal("Command() error = nil, want an unreachable interpreter")
	}
	for _, want := range []string{"interpreter /bin/sh of /usr/bin/claude resolves to /nix/store/abc-bash/bin/sh, outside", "provision=nix", "static agent binary"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}
}

func TestPassthroughAcceptsAnInterpreterInsideTheBinds(t *testing.T) {
	host := newHost().AddFile("/usr/bin/dash", staticBinary, 0o755).AddLink("/bin/sh", "/usr/bin/dash")
	host.AddFile("/usr/bin/claude", "#!/bin/sh\nexec agent\n", 0o755)

	args := bwrapCommand(t, host, claudeCfg(network.NetworkNone, true), provision.Contribution{})

	if !argHasTriple(args, "--ro-bind", "/bin", "/bin") {
		t.Errorf("args = %v, want the passthrough binds", args)
	}
}

func TestPassthroughAcceptsAStaticBinary(t *testing.T) {
	host := nixHost().AddFile("/usr/bin/claude", staticBinary, 0o755)

	if _, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{}); err != nil {
		t.Fatalf("Command(static) error = %v", err)
	}
}

func TestPassthroughChecksTheLoaderOfADynamicBinary(t *testing.T) {
	cases := map[string]struct {
		image string
		fails bool
	}{
		"loader in the store":         {elfImage(binary.LittleEndian, "/nix/store/def-glibc/lib/ld-linux-x86-64.so.2"), true},
		"big-endian loader in store":  {elfImage(binary.BigEndian, "/nix/store/def-glibc/lib/ld-linux-x86-64.so.2"), true},
		"32-bit loader in the store":  {elf32Image("/nix/store/def-glibc/lib/ld-linux-x86-64.so.2"), true},
		"loader in a bound directory": {elfImage(binary.LittleEndian, "/usr/lib/ld-linux-x86-64.so.2"), false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			host := nixHost().AddFile("/usr/lib/ld-linux-x86-64.so.2", staticBinary, 0o755).AddFile("/usr/bin/claude", c.image, 0o755)

			_, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{})

			if fails := err != nil; fails != c.fails {
				t.Fatalf("Command() error = %v, want failure %v", err, c.fails)
			}
		})
	}
}

func TestWithoutPassthroughTheProgramsAreNotChecked(t *testing.T) {
	host := nixHost().AddFile("/usr/bin/claude", "#!/bin/sh\n", 0o755)

	if _, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, false), provision.Contribution{InitCommands: []string{"true"}}); err != nil {
		t.Fatalf("Command(no passthrough) error = %v", err)
	}
}

func TestPassthroughRejectsAnAgentOutsideTheBinds(t *testing.T) {
	host := nixHost().AddFile("/nix/store/ghi-claude/bin/claude", staticBinary, 0o755).AddLink("/usr/bin/claude", "/nix/store/ghi-claude/bin/claude")

	_, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{})

	if err == nil || !strings.Contains(err.Error(), "agent executable /usr/bin/claude resolves to /nix/store/ghi-claude/bin/claude, outside") {
		t.Fatalf("Command() error = %v, want the unreachable agent", err)
	}
}

func TestPassthroughNamesAnAgentPathOutsideTheBinds(t *testing.T) {
	host := newHost().AddFile("/home/user/.local/bin/claude", staticBinary, 0o755)
	host.Env["PATH"] = "/home/user/.local/bin"

	_, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{})

	if err == nil || !strings.Contains(err.Error(), "agent executable /home/user/.local/bin/claude is outside") || !strings.Contains(err.Error(), "a read-only bind of /home/user/.local/bin") {
		t.Fatalf("Command() error = %v, want the unbound agent path", err)
	}
}

func TestPassthroughTakesTheFirstReachableCandidate(t *testing.T) {
	host := nixHost().AddFile("/nix/store/ghi-claude/bin/claude", staticBinary, 0o755).AddLink("/usr/local/bin/claude", "/nix/store/ghi-claude/bin/claude")
	host.AddFile("/usr/bin/claude", staticBinary, 0o755).AddFile("/usr/bin/notes", "text", 0o644)

	if _, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{}); err != nil {
		t.Fatalf("Command() error = %v, want the reachable /usr/bin/claude", err)
	}
}

func TestPassthroughFollowsEnvToTheInterpreter(t *testing.T) {
	cases := map[string]struct {
		node  func(*oshosttest.Fake)
		fails bool
	}{
		"node in the store": {func(h *oshosttest.Fake) {
			h.AddFile("/nix/store/jkl-node/bin/node", staticBinary, 0o755).AddLink("/usr/bin/node", "/nix/store/jkl-node/bin/node")
		}, true},
		"node in a bound directory": {func(h *oshosttest.Fake) { h.AddFile("/usr/bin/node", staticBinary, 0o755) }, false},
		"no node":                   {func(*oshosttest.Fake) {}, false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			host := nixHost().AddFile("/usr/bin/env", staticBinary, 0o755).AddFile("/usr/bin/claude", "#!/usr/bin/env -S NODE_OPTIONS=x node --flag\n", 0o755)
			c.node(host)

			_, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{})

			if fails := err != nil; fails != c.fails {
				t.Fatalf("Command() error = %v, want failure %v", err, c.fails)
			}
			if c.fails && !strings.Contains(err.Error(), "interpreter /usr/bin/node of /usr/bin/claude resolves to /nix/store/jkl-node/bin/node") {
				t.Errorf("error %q does not name node", err)
			}
		})
	}
}

func TestPassthroughChecksTheShellOfTheInitCommands(t *testing.T) {
	host := nixHost().AddFile("/usr/bin/claude", staticBinary, 0o755)

	_, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{InitCommands: []string{"true"}})

	if err == nil || !strings.Contains(err.Error(), "agent executable /bin/sh resolves to /nix/store/abc-bash/bin/sh") {
		t.Fatalf("Command(init commands) error = %v, want the unreachable shell", err)
	}
}

func TestPassthroughResolvesARelativeProgramInTheWorkDirectory(t *testing.T) {
	host := nixHost().AddFile("/work/agent.sh", "#!/bin/sh\n", 0o755)
	host.Env["PATH"] = "tools"
	host.AddFile("/work/tools/helper", staticBinary, 0o755)
	cfg := claudeCfg(network.NetworkNone, true)

	cfg.Process.Argv = []string{"./agent.sh"}
	if _, err := NewIsolator(host, testLog).Command(cfg, provision.Contribution{}); err == nil || !strings.Contains(err.Error(), "interpreter /bin/sh of /work/agent.sh") {
		t.Errorf("Command(./agent.sh) error = %v, want the unreachable interpreter", err)
	}
	cfg.Process.Argv = []string{"helper"}
	if _, err := NewIsolator(host, testLog).Command(cfg, provision.Contribution{}); err != nil {
		t.Errorf("Command(helper on a relative PATH entry) error = %v", err)
	}
}

func TestPassthroughStopsAtTheInterpreterDepth(t *testing.T) {
	host := newHost().AddFile("/usr/bin/self", "#!/usr/bin/self\n", 0o755).AddFile("/usr/bin/claude", "#!/usr/bin/self\n", 0o755)

	if _, err := NewIsolator(host, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{}); err != nil {
		t.Fatalf("Command(looping interpreter) error = %v", err)
	}
}

func TestPassthroughSkipsAnEmptyArgumentVector(t *testing.T) {
	cfg := claudeCfg(network.NetworkNone, true)
	cfg.Process.Argv = nil

	if _, err := NewIsolator(newHost(), testLog).Command(cfg, provision.Contribution{}); err != nil {
		t.Fatalf("Command(no argv) error = %v", err)
	}
}

// faultyHost fails one file operation of the fake host.
type faultyHost struct {
	*oshosttest.Fake
	readFails, resolveFails bool
}

func (h faultyHost) ReadPrefix(path string, limit int) ([]byte, error) {
	if h.readFails {
		return nil, errors.New("permission denied")
	}
	return h.Fake.ReadPrefix(path, limit)
}

func (h faultyHost) EvalSymlinks(path string) (string, error) {
	if h.resolveFails && path == "/usr/bin/claude" {
		return "", errors.New("too many links")
	}
	return h.Fake.EvalSymlinks(path)
}

func TestPassthroughWithAFileItCannotReadOrResolve(t *testing.T) {
	unreadable := faultyHost{Fake: nixHost().AddFile("/usr/bin/claude", "#!/bin/sh\n", 0o711), readFails: true}
	if _, err := NewIsolator(unreadable, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{}); err != nil {
		t.Errorf("Command(execute-only agent) error = %v, want no claim", err)
	}
	unresolved := faultyHost{Fake: nixHost().AddFile("/usr/bin/claude", staticBinary, 0o755), resolveFails: true}
	if _, err := NewIsolator(unresolved, testLog).Command(claudeCfg(network.NetworkNone, true), provision.Contribution{}); err == nil || !strings.Contains(err.Error(), "resolve agent executable /usr/bin/claude") {
		t.Errorf("Command(unresolvable agent) error = %v", err)
	}
}

func TestInterpreterOf(t *testing.T) {
	cases := []struct {
		header, interpreter, program string
	}{
		{"#!/bin/sh\n", "/bin/sh", ""},
		{"#! /bin/bash -e\nrest", "/bin/bash", ""},
		{"#!\t/usr/bin/env\tnode", "/usr/bin/env", "node"},
		{"#!/usr/bin/env -i\n", "/usr/bin/env", ""},
		{"plain text", "", ""},
		{"\x7fEL", "", ""},
		{staticBinary, "", ""},
		{elfImage(binary.LittleEndian, "/lib/ld.so")[:100], "", ""},
		{elfImage(binary.LittleEndian, "/lib/ld.so")[:125], "", ""},
		{elfImage(binary.LittleEndian, "/lib/ld.so")[:70], "", ""},
		{elfImage(binary.LittleEndian, "/lib/ld.so")[:20], "", ""},
		{elf32Image("/lib/ld.so")[:30], "", ""},
		{elfImage(binary.LittleEndian, "/lib/ld.so"), "/lib/ld.so", ""},
		{elf32Image("/lib/ld.so"), "/lib/ld.so", ""},
	}
	for _, c := range cases {
		interpreter, program := interpreterOf([]byte(c.header))
		if interpreter != c.interpreter || program != c.program {
			t.Errorf("interpreterOf(%q) = %q, %q, want %q, %q", c.header, interpreter, program, c.interpreter, c.program)
		}
	}
}

func TestElfInterpreterIgnoresProgramHeadersPastTheHeader(t *testing.T) {
	image := []byte(elfImage(binary.LittleEndian, "/lib/ld.so"))
	binary.LittleEndian.PutUint64(image[32:], 1<<20)

	if interpreter := elfInterpreter(image); interpreter != "" {
		t.Errorf("elfInterpreter(far program headers) = %q, want none", interpreter)
	}
}

func TestElfInterpreterRejectsAnOverflowingInterpreterRange(t *testing.T) {
	image := []byte(elfImage(binary.LittleEndian, "/lib/ld.so"))
	// The PT_INTERP program header starts at 64: p_offset at 72, p_filesz at 96.
	binary.LittleEndian.PutUint64(image[72:], ^uint64(0))
	binary.LittleEndian.PutUint64(image[96:], 2)

	if interpreter := elfInterpreter(image); interpreter != "" {
		t.Errorf("elfInterpreter(overflowing range) = %q, want none", interpreter)
	}
}
