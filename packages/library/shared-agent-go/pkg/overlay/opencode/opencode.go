// Package opencode is the overlay dialect leaf of OpenCode. It makes the overlay
// root the OPENCODE_CONFIG_DIR, which OpenCode searches after the configuration
// of the user and of the project, so both stay as they are and the overlay wins
// on a name they share.
package opencode

import (
	"encoding/json"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

const (
	skillsDirectory   = "skills"
	commandsDirectory = "commands"
	agentsDirectory   = "agents"
	instructionFile   = "AGENTS.md"
	configurationFile = "opencode.json"
	ignoreFile        = ".gitignore"
	directoryVariable = "OPENCODE_CONFIG_DIR"
	schemaURL         = "https://opencode.ai/config.json"
)

// ignoreContent is the .gitignore OpenCode writes into a configuration directory
// that lacks one. Placing it keeps OpenCode from a write that an isolation
// denies.
const ignoreContent = "node_modules\npackage.json\npackage-lock.json\nbun.lock\n.gitignore"

type dialect struct{}

// New returns the dialect.
func New() shared.Dialect { return dialect{} }

// Facts declares the project file OpenCode reads as its own instructions.
func (dialect) Facts() shared.Facts {
	return shared.Facts{NativeInstructions: []string{instructionFile}}
}

// Place builds the overlay of a run. The model names of an artifact belong to
// another harness, and the tool list of an agent has another form, so the dialect
// drops them. The instruction is a file that an opencode.json beside it lists,
// and the text never enters a variable.
func (dialect) Place(input shared.Input) (shared.Placement, error) {
	if err := shared.CheckRoot(input.Root); err != nil {
		return shared.Placement{}, err
	}
	var placement shared.Placement
	for _, skill := range input.Artifacts.Skills {
		placement.Files = append(placement.Files, shared.SkillFiles(skillsDirectory, skill, "model")...)
	}
	for _, command := range input.Artifacts.Commands {
		placement.Files = append(placement.Files, shared.DocumentFile(commandsDirectory+"/"+command.Name+".md", artifact.KindCommand, command.Document, "model"))
	}
	for _, agent := range input.Artifacts.Agents {
		placement.Files = append(placement.Files, shared.DocumentFile(agentsDirectory+"/"+agent.Name+".md", artifact.KindAgent, agent.Document, "model", "tools"))
	}
	if skip, native := shared.InstructionSkip(input); native {
		placement.Skipped = append(placement.Skipped, skip)
	} else if text := input.Artifacts.InstructionText(); text != "" {
		configuration, _ := json.MarshalIndent(struct {
			Schema       string   `json:"$schema"`
			Instructions []string `json:"instructions"`
		}{schemaURL, []string{shared.Join(input.Root, instructionFile)}}, "", "  ")
		placement.Files = append(placement.Files,
			shared.File{Path: instructionFile, Kind: artifact.KindInstruction, Content: []byte(text)},
			shared.File{Path: configurationFile, Kind: artifact.KindSupport, Content: append(configuration, '\n')},
		)
	}
	if len(placement.Files) > 0 {
		placement.Files = append(placement.Files, shared.File{Path: ignoreFile, Kind: artifact.KindSupport, Content: []byte(ignoreContent)})
		placement.Environment = []string{directoryVariable + "=" + input.Root}
		placement.RootWritable = true
	}
	return shared.Finish(placement)
}
