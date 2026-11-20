// Package pi is the overlay dialect leaf of Pi. It places the artifacts in the
// overlay root and names each one in a repeatable flag, so the Pi directory of
// the user and of the project stay as they are, and the trust gate of the project
// directories is not involved.
package pi

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

const (
	skillsDirectory    = "skills"
	promptsDirectory   = "prompts"
	instructionFile    = "AGENTS.md"
	skillFlag          = "--skill"
	promptTemplateFlag = "--prompt-template"
	instructionFlag    = "--append-system-prompt"
)

type dialect struct{}

// New returns the dialect.
func New() shared.Dialect { return dialect{} }

// Facts declares the project files Pi reads as its own instructions.
func (dialect) Facts() shared.Facts {
	return shared.Facts{NativeInstructions: []string{"AGENTS.md", "CLAUDE.md"}}
}

// Place builds the overlay of a run. Pi loads a skill from a directory, a prompt
// template from a file and appends the text of one file to its system prompt. It
// has no agents.
func (dialect) Place(input shared.Input) (shared.Placement, error) {
	if err := shared.CheckRoot(input.Root); err != nil {
		return shared.Placement{}, err
	}
	var placement shared.Placement
	for _, skill := range input.Artifacts.Skills {
		placement.Files = append(placement.Files, shared.SkillFiles(skillsDirectory, skill, "model")...)
		placement.Arguments = append(placement.Arguments, skillFlag, shared.Join(input.Root, skillsDirectory+"/"+skill.Name))
	}
	for _, command := range input.Artifacts.Commands {
		file := promptsDirectory + "/" + command.Name + ".md"
		placement.Files = append(placement.Files, shared.DocumentFile(file, artifact.KindCommand, command.Document, "model"))
		placement.Arguments = append(placement.Arguments, promptTemplateFlag, shared.Join(input.Root, file))
	}
	for _, agent := range input.Artifacts.Agents {
		placement.Skipped = append(placement.Skipped, shared.Unsupported(artifact.KindAgent, agent.Name, "pi"))
	}
	if skip, native := shared.InstructionSkip(input); native {
		placement.Skipped = append(placement.Skipped, skip)
	} else if text := input.Artifacts.InstructionText(); text != "" {
		placement.Files = append(placement.Files, shared.File{Path: instructionFile, Kind: artifact.KindInstruction, Content: []byte(text)})
		placement.Arguments = append(placement.Arguments, instructionFlag, shared.Join(input.Root, instructionFile))
	}
	return shared.Finish(placement)
}
