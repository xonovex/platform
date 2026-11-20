// Package copilot is the overlay dialect leaf of GitHub Copilot CLI. It places
// the artifacts in a plugin directory that Copilot loads for one run, and the
// instruction in a directory that a variable names, so the Copilot directory of
// the user and the .github directory of the project stay as they are.
package copilot

import (
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

const (
	pluginDirectory       = "plugin"
	pluginManifest        = pluginDirectory + "/plugin.json"
	skillsDirectory       = pluginDirectory + "/skills"
	commandsDirectory     = pluginDirectory + "/commands"
	agentsDirectory       = pluginDirectory + "/agents"
	instructionsDirectory = "instructions"
	instructionFile       = instructionsDirectory + "/AGENTS.md"
	pluginFlag            = "--plugin-dir"
	instructionVariable   = "COPILOT_CUSTOM_INSTRUCTIONS_DIRS"
)

// manifest is the plugin.json Copilot requires before it loads a plugin
// directory. It names the plugin and carries nothing else.
const manifest = `{"name":"agent-cli-overlay","description":"The artifacts of one run","version":"1.0.0"}` + "\n"

type dialect struct{}

// New returns the dialect.
func New() shared.Dialect { return dialect{} }

// Facts declares the project files Copilot reads as its own instructions.
func (dialect) Facts() shared.Facts {
	return shared.Facts{NativeInstructions: []string{"AGENTS.md", "CLAUDE.md", "GEMINI.md", ".github/copilot-instructions.md"}}
}

// Place builds the overlay of a run. Copilot loads skills, commands and agents
// from the plugin directory that --plugin-dir names, which neither writes the
// Copilot home nor grants file access. It loads a command and a skill of one name
// as one entry, so the skill wins.
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
		if skills[command.Name] {
			placement.Skipped = append(placement.Skipped, shared.CommandShadowed(command.Name))
			continue
		}
		placement.Files = append(placement.Files, shared.DocumentFile(commandsDirectory+"/"+command.Name+".md", artifact.KindCommand, command.Document, "model"))
	}
	for _, agent := range input.Artifacts.Agents {
		placement.Files = append(placement.Files, shared.DocumentFile(agentsDirectory+"/"+agent.Name+".md", artifact.KindAgent, agent.Document, "model"))
	}
	if len(placement.Files) > 0 {
		placement.Files = append(placement.Files, shared.File{Path: pluginManifest, Kind: artifact.KindSupport, Content: []byte(manifest)})
		placement.Arguments = []string{pluginFlag, shared.Join(input.Root, pluginDirectory)}
	}
	if skip, native := shared.InstructionSkip(input); native {
		placement.Skipped = append(placement.Skipped, skip)
	} else if text := input.Artifacts.InstructionText(); text != "" {
		placement.Files = append(placement.Files, shared.File{Path: instructionFile, Kind: artifact.KindInstruction, Content: []byte(text)})
		placement.Environment = []string{instructionVariable + "=" + shared.Join(input.Root, instructionsDirectory)}
	}
	return shared.Finish(placement)
}
