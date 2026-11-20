package bwrap

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// headerLimit covers the #! line, which the Linux kernel reads from the
	// first 256 bytes, and the ELF program headers with the interpreter path of
	// a dynamically linked executable.
	headerLimit = 4096
	// interpreterDepth is the number of nested interpreters the Linux kernel
	// follows before it refuses to start a program.
	interpreterDepth = 4
)

// agentRole describes, in an error, a program that the sandbox starts by name.
// A role is a format with one %s for the path.
const agentRole = "agent executable %s"

// sandboxView is what a passthrough sandbox sees of the host: the destinations
// of its binds, the PATH it starts the agent with, and its work directory.
type sandboxView struct {
	binds   []string
	path    []string
	workDir string
}

// contains reports whether p lies under the destination of a bind.
func (v sandboxView) contains(p string) bool {
	for _, bind := range v.binds {
		if p == bind || strings.HasPrefix(p, strings.TrimSuffix(bind, "/")+"/") {
			return true
		}
	}
	return false
}

// candidates lists the paths the sandbox tries for program, in the order that
// execvp tries them.
func (v sandboxView) candidates(program string) []string {
	if strings.Contains(program, "/") {
		return []string{v.absolute(program)}
	}
	out := make([]string, 0, len(v.path))
	for _, dir := range v.path {
		out = append(out, filepath.Join(v.absolute(dir), program))
	}
	return out
}

func (v sandboxView) absolute(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(v.workDir, p)
}

// checkReachable fails when the executable that the sandbox starts for program,
// or an interpreter that starts it, resolves outside the binds of the sandbox.
// Host passthrough binds the system directories but not /nix/store, so on NixOS
// a #!/bin/sh script or a dynamically linked binary would end in execvp ENOENT
// inside the sandbox. A program or an interpreter that the host does not have
// fails on the host too, so the check makes no claim about it.
func (i *Isolator) checkReachable(view sandboxView, program string) error {
	path, target, err := i.locate(view, agentRole, program)
	if err != nil || path == "" {
		return err
	}
	return i.checkInterpreters(view, path, target, interpreterDepth)
}

// locate returns the first candidate of program that the sandbox reaches and its
// final symlink target. It returns empty paths when the host has no candidate,
// and an error naming the first candidate when the sandbox reaches none.
func (i *Isolator) locate(view sandboxView, role, program string) (string, string, error) {
	var first, firstTarget string
	for _, candidate := range view.candidates(program) {
		info, err := i.host.Stat(candidate)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
			continue
		}
		target, err := i.host.EvalSymlinks(candidate)
		if err != nil {
			return "", "", fmt.Errorf("resolve %s: %w", fmt.Sprintf(role, candidate), err)
		}
		if view.contains(candidate) && view.contains(target) {
			return candidate, target, nil
		}
		if first == "" {
			first, firstTarget = candidate, target
		}
	}
	if first == "" {
		return "", "", nil
	}
	return "", "", unreachable(role, first, firstTarget)
}

// checkInterpreters follows the #! line or the ELF interpreter of the file at
// target and requires each interpreter to resolve inside the binds.
func (i *Isolator) checkInterpreters(view sandboxView, path, target string, depth int) error {
	if depth == 0 {
		return nil
	}
	header, err := i.host.ReadPrefix(target, headerLimit)
	if err != nil {
		// An execute-only file cannot be read, and the kernel still starts it:
		// the check makes no claim about it.
		return nil
	}
	interpreter, program := interpreterOf(header)
	if interpreter == "" {
		return nil
	}
	role := "interpreter %s of " + path
	next, nextTarget, err := i.locate(view, role, interpreter)
	if err == nil && next != "" && program != "" {
		// #!/usr/bin/env names the real interpreter as its argument, and env
		// finds it on the PATH of the sandbox.
		next, nextTarget, err = i.locate(view, role, program)
	}
	if err != nil || next == "" {
		return err
	}
	return i.checkInterpreters(view, next, nextTarget, depth-1)
}

// interpreterOf returns the interpreter that the kernel starts for a file with
// header, and, for #!/usr/bin/env, the program that env runs.
func interpreterOf(header []byte) (string, string) {
	if line, ok := bytes.CutPrefix(header, []byte("#!")); ok {
		if end := bytes.IndexByte(line, '\n'); end >= 0 {
			line = line[:end]
		}
		interpreter, argument, _ := strings.Cut(strings.TrimSpace(strings.ReplaceAll(string(line), "\t", " ")), " ")
		if filepath.Base(interpreter) == "env" {
			return interpreter, envProgram(argument)
		}
		return interpreter, ""
	}
	return elfInterpreter(header), ""
}

// envProgram returns the first argument of env that is neither an option nor
// an assignment.
func envProgram(argument string) string {
	for _, field := range strings.Fields(argument) {
		if !strings.HasPrefix(field, "-") && !strings.Contains(field, "=") {
			return field
		}
	}
	return ""
}

// elfInterpreter returns the PT_INTERP path of an ELF header, or "" for a
// static executable, a file that is no ELF, and a header that ends before the
// path. It reads only the file header and the program headers, which come first
// in the file, so a prefix of the file is enough.
func elfInterpreter(header []byte) string {
	if len(header) < elf.EI_NIDENT || string(header[:4]) != elf.ELFMAG {
		return ""
	}
	var order binary.ByteOrder = binary.LittleEndian
	if elf.Data(header[elf.EI_DATA]) == elf.ELFDATA2MSB {
		order = binary.BigEndian
	}
	reader := bytes.NewReader(header)
	var programs []elf.Prog64
	switch elf.Class(header[elf.EI_CLASS]) {
	case elf.ELFCLASS64:
		var file elf.Header64
		if binary.Read(reader, order, &file) != nil {
			return ""
		}
		programs = readPrograms[elf.Prog64](header, order, file.Phoff, file.Phnum, func(p elf.Prog64) elf.Prog64 { return p })
	case elf.ELFCLASS32:
		var file elf.Header32
		if binary.Read(reader, order, &file) != nil {
			return ""
		}
		programs = readPrograms[elf.Prog32](header, order, uint64(file.Phoff), file.Phnum, func(p elf.Prog32) elf.Prog64 {
			return elf.Prog64{Type: p.Type, Off: uint64(p.Off), Filesz: uint64(p.Filesz)}
		})
	}
	for _, program := range programs {
		if elf.ProgType(program.Type) != elf.PT_INTERP {
			continue
		}
		size := uint64(len(header))
		if program.Off > size || program.Filesz > size-program.Off {
			return ""
		}
		return string(bytes.TrimRight(header[program.Off:program.Off+program.Filesz], "\x00"))
	}
	return ""
}

// readPrograms decodes the program headers that lie inside header.
func readPrograms[P elf.Prog32 | elf.Prog64](header []byte, order binary.ByteOrder, offset uint64, count uint16, widen func(P) elf.Prog64) []elf.Prog64 {
	if offset > uint64(len(header)) {
		return nil
	}
	reader := bytes.NewReader(header[offset:])
	out := make([]elf.Prog64, 0, count)
	for range count {
		var program P
		if binary.Read(reader, order, &program) != nil {
			break
		}
		out = append(out, widen(program))
	}
	return out
}

func unreachable(role, path, target string) error {
	where := "is outside"
	if target != path {
		where = "resolves to " + target + ", outside"
	}
	remedy := "use provision=nix or a static agent binary"
	if role == agentRole {
		remedy = "use provision=nix, a static agent binary, or a read-only bind of " + filepath.Dir(target)
	}
	return fmt.Errorf("%s %s the paths that bwrap binds with host passthrough, which never include /nix/store; %s", fmt.Sprintf(role, path), where, remedy)
}
