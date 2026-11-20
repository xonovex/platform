// Package artifact is the neutral model of what a plugin or an overlay source
// gives an agent: skills, agents, commands and instructions, read from a
// harness-independent directory layout. It is pure: it parses text and bytes it
// is given and never touches the machine.
package artifact

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Kind names a class of artifact.
type Kind string

const (
	KindSkill       Kind = "skill"
	KindAgent       Kind = "agent"
	KindCommand     Kind = "command"
	KindInstruction Kind = "instruction"
	// KindSupport is a file a dialect adds so the harness accepts the overlay,
	// such as a manifest or a settings file. No artifact holds it.
	KindSupport Kind = "support"
)

// The directory layout of a source. A skill is a directory with a SKILL.md, an
// agent and a command are one Markdown file, and the instruction is the
// AGENTS.md of the source root.
const (
	SkillsDirectory   = "skills"
	AgentsDirectory   = "agents"
	CommandsDirectory = "commands"
	InstructionFile   = "AGENTS.md"
	SkillFile         = "SKILL.md"
)

var namePattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// maxNameLength is the longest name of a skill, an agent, a command or a plugin.
const maxNameLength = 64

// ValidName reports whether name is a lowercase kebab-case name of at most 64
// characters.
func ValidName(name string) bool {
	return len(name) <= maxNameLength && namePattern.MatchString(name)
}

// ConfinedPath checks that p is a clean, relative, slash-separated path that
// stays below the directory it is joined to.
func ConfinedPath(p string) error {
	if p == "" {
		return errors.New("empty path")
	}
	if strings.HasPrefix(p, "/") {
		return fmt.Errorf("path %q is absolute", p)
	}
	for _, character := range p {
		if character < ' ' || character == 0x7f || character == '\\' {
			return fmt.Errorf("path %q holds a control or backslash character", p)
		}
	}
	for _, segment := range strings.Split(p, "/") {
		switch segment {
		case "", ".", "..":
			return fmt.Errorf("path %q is not clean and confined", p)
		}
	}
	return nil
}

// File is one regular file of a source.
type File struct {
	// Path is relative to the directory that holds it, in slash form.
	Path       string
	Executable bool
	Content    []byte
}

// Skill is a directory with a SKILL.md and the files beside it.
type Skill struct {
	Name   string
	Source string
	// Document is the text of SKILL.md.
	Document string
	// Files are the other files of the skill directory, relative to it.
	Files []File
}

// Agent is a Markdown file that defines a sub-agent.
type Agent struct {
	Name     string
	Source   string
	Document string
}

// Command is a Markdown file that defines a prompt the user starts by name.
type Command struct {
	Name     string
	Source   string
	Document string
}

// Instruction is the text of an AGENTS.md.
type Instruction struct {
	Source  string
	Content string
}

// Set holds the artifacts of one or more sources, each list sorted by name.
type Set struct {
	Skills       []Skill
	Agents       []Agent
	Commands     []Command
	Instructions []Instruction
}

// Empty reports whether the set holds nothing.
func (s Set) Empty() bool {
	return len(s.Skills)+len(s.Agents)+len(s.Commands)+len(s.Instructions) == 0
}

// InstructionText joins the instructions of the set in source order, separated
// by a blank line, and returns an empty string when there is none.
func (s Set) InstructionText() string {
	parts := make([]string, 0, len(s.Instructions))
	for _, instruction := range s.Instructions {
		parts = append(parts, strings.TrimRight(instruction.Content, "\n"))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n") + "\n"
}
