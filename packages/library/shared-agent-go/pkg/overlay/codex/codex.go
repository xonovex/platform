// Package codex is the overlay dialect leaf of Codex. It makes the overlay root
// the CODEX_HOME of the run, which Codex reads its skills, agents and profile
// from. The home of the user stays as it is, and the run sees none of its
// configuration or login, so the run takes its model access from a provider.
package codex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

const (
	skillsDirectory  = "skills"
	agentsDirectory  = "agents"
	invocationPolicy = "agents/openai.yaml"
	profileName      = "agentic"
	profileFile      = profileName + ".config.toml"
	homeVariable     = "CODEX_HOME"
	profileFlag      = "-p"
)

// noImplicitInvocation is the policy file that keeps a skill from starting unless
// the user names it, which is how a command behaves.
const noImplicitInvocation = "policy:\n  allow_implicit_invocation: false\n"

type dialect struct{}

// New returns the dialect.
func New() shared.Dialect { return dialect{} }

// Facts declares the project file Codex reads as its own instructions, and that
// the overlay root replaces the Codex home.
func (dialect) Facts() shared.Facts {
	return shared.Facts{NativeInstructions: []string{"AGENTS.md"}, ReplacesHome: true}
}

// Place builds the overlay of a run. A command becomes a skill that only a
// mention starts, because Codex has no command files. An agent is the TOML file
// Codex documents for a custom agent. The instruction is the developer
// instructions of a profile that -p selects, so it adds to the AGENTS.md of the
// project and replaces nothing.
func (dialect) Place(input shared.Input) (shared.Placement, error) {
	if err := shared.CheckRoot(input.Root); err != nil {
		return shared.Placement{}, err
	}
	var placement shared.Placement
	skills := shared.SkillNames(input.Artifacts)
	for _, skill := range input.Artifacts.Skills {
		placement.Files = append(placement.Files, shared.SkillFiles(skillsDirectory, skill, "model")...)
	}
	for _, command := range input.Artifacts.Commands {
		files, skip := placeCommand(command, skills)
		placement.Files = append(placement.Files, files...)
		placement.Skipped = append(placement.Skipped, skip...)
	}
	for _, agent := range input.Artifacts.Agents {
		file, skip := placeAgent(agent)
		placement.Files = append(placement.Files, file...)
		placement.Skipped = append(placement.Skipped, skip...)
	}
	if skip, native := shared.InstructionSkip(input); native {
		placement.Skipped = append(placement.Skipped, skip)
	} else if text := input.Artifacts.InstructionText(); text != "" {
		placement.Files = append(placement.Files, shared.File{
			Path: profileFile, Kind: artifact.KindInstruction,
			Content: []byte("developer_instructions = " + tomlString(text) + "\n"),
		})
		placement.Arguments = []string{profileFlag, profileName}
	}
	if len(placement.Files) > 0 {
		placement.Environment = []string{homeVariable + "=" + input.Root}
		placement.RootWritable = true
	}
	return shared.Finish(placement)
}

func placeCommand(command artifact.Command, skills map[string]bool) ([]shared.File, []shared.Skip) {
	if skills[command.Name] {
		return nil, []shared.Skip{shared.CommandShadowed(command.Name)}
	}
	document, _ := artifact.ParseDocument(command.Document)
	if document.Fields["description"] == "" {
		return nil, []shared.Skip{{Kind: artifact.KindCommand, Name: command.Name, Reason: "codex delivers a command as a skill, and a skill needs a description"}}
	}
	text := "---\nname: " + command.Name + "\ndescription: " + strconv.Quote(document.Fields["description"]) + "\n---\n" + document.Body
	directory := skillsDirectory + "/" + command.Name + "/"
	return []shared.File{
		{Path: directory + artifact.SkillFile, Kind: artifact.KindCommand, Content: []byte(text)},
		{Path: directory + invocationPolicy, Kind: artifact.KindCommand, Content: []byte(noImplicitInvocation)},
	}, nil
}

func placeAgent(agent artifact.Agent) ([]shared.File, []shared.Skip) {
	document, _ := artifact.ParseDocument(agent.Document)
	instructions := strings.TrimSpace(document.Body)
	if instructions == "" {
		return nil, []shared.Skip{{Kind: artifact.KindAgent, Name: agent.Name, Reason: "codex agents take their instructions from the body, which is empty"}}
	}
	var text strings.Builder
	fmt.Fprintf(&text, "name = %s\n", tomlString(agent.Name))
	if description := document.Fields["description"]; description != "" {
		fmt.Fprintf(&text, "description = %s\n", tomlString(description))
	}
	fmt.Fprintf(&text, "developer_instructions = %s\n", tomlString(instructions))
	return []shared.File{{Path: agentsDirectory + "/" + agent.Name + ".toml", Kind: artifact.KindAgent, Content: []byte(text.String())}}, nil
}

// tomlString encodes s as a TOML basic string.
func tomlString(s string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, character := range s {
		switch {
		case character == '"':
			out.WriteString(`\"`)
		case character == '\\':
			out.WriteString(`\\`)
		case character == '\n':
			out.WriteString(`\n`)
		case character == '\t':
			out.WriteString(`\t`)
		case character == '\r':
			out.WriteString(`\r`)
		case character < ' ' || character == 0x7f:
			fmt.Fprintf(&out, `\u%04X`, character)
		default:
			out.WriteRune(character)
		}
	}
	out.WriteByte('"')
	return out.String()
}
