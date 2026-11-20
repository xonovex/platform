// Package claude is the overlay dialect leaf of Claude Code. It places the
// artifacts below a .claude directory in the overlay root and adds the root with
// --add-dir, so the .claude directory of the user and of the project stay as
// they are.
package claude

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

const (
	skillsDirectory     = ".claude/skills"
	commandsDirectory   = ".claude/commands"
	agentsDirectory     = ".claude/agents"
	instructionFile     = "CLAUDE.md"
	instructionVariable = "CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1"
	addDirectoryFlag    = "--add-dir"
)

type dialect struct{}

// New returns the dialect.
func New() shared.Dialect { return dialect{} }

// Facts declares the project files Claude reads as its own instructions. The
// overlay is an added directory, so the home of the user stays in place.
func (dialect) Facts() shared.Facts {
	return shared.Facts{NativeInstructions: []string{"CLAUDE.md", ".claude/CLAUDE.md"}}
}

// Place builds the overlay of a run. Claude reads skills, commands and agents
// from the .claude directory of every added directory, and the CLAUDE.md of an
// added directory when the variable asks for it. The instruction is a copy of the
// text and not an @ import, because Claude asks the user to approve an import
// that leaves the work directory.
func (dialect) Place(input shared.Input) (shared.Placement, error) {
	if err := shared.CheckRoot(input.Root); err != nil {
		return shared.Placement{}, err
	}
	var placement shared.Placement
	for _, skill := range input.Artifacts.Skills {
		placement.Files = append(placement.Files, shared.SkillFiles(skillsDirectory, skill)...)
	}
	for _, command := range input.Artifacts.Commands {
		placement.Files = append(placement.Files, shared.DocumentFile(commandsDirectory+"/"+command.Name+".md", artifact.KindCommand, command.Document))
	}
	for _, agent := range input.Artifacts.Agents {
		placement.Files = append(placement.Files, shared.DocumentFile(agentsDirectory+"/"+agent.Name+".md", artifact.KindAgent, agent.Document))
	}
	if skip, native := shared.InstructionSkip(input); native {
		placement.Skipped = append(placement.Skipped, skip)
	} else if text := input.Artifacts.InstructionText(); text != "" {
		placement.Files = append(placement.Files, shared.File{Path: instructionFile, Kind: artifact.KindInstruction, Content: []byte(text)})
		placement.Environment = []string{instructionVariable}
	}
	if len(placement.Files) > 0 {
		placement.Arguments = []string{addDirectoryFlag, input.Root}
	}
	return shared.Finish(placement)
}
