// Package kiro is the overlay dialect leaf of Kiro CLI. It makes the overlay root
// the KIRO_HOME of the run, where Kiro reads its skills, prompts, steering and
// agents. That home also holds the login of the user, so a run with an overlay
// cannot use the login of Kiro: the launcher refuses the combination.
package kiro

import (
	"encoding/json"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/artifact"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/overlay/shared"
)

const (
	skillsDirectory   = "skills"
	promptsDirectory  = "prompts"
	steeringFile      = "steering/instructions.md"
	agentsDirectory   = "agents"
	settingsFile      = "settings/cli.json"
	homeVariable      = "KIRO_HOME"
	steeringFrontPage = "---\ninclusion: always\n---\n"
)

// settings is the cli.json of the home. Kiro reads telemetry and update checks
// from it, so a generated home turns both off.
const settings = "{\n  \"app.disableAutoupdates\": true,\n  \"telemetry.enabled\": false\n}\n"

type dialect struct{}

// New returns the dialect.
func New() shared.Dialect { return dialect{} }

// Facts declares the project file Kiro reads as its own instructions, and that
// the overlay root replaces the Kiro home.
func (dialect) Facts() shared.Facts {
	return shared.Facts{NativeInstructions: []string{"AGENTS.md"}, ReplacesHome: true}
}

// Place builds the configuration home of a run. A prompt is the body of a command
// with none of its fields, because Kiro sends a prompt file as it is. Kiro offers
// a prompt and a skill of one name as one slash command, so the skill wins. Each
// agent lists the skills and the steering file of the home as its resources.
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
		document, _ := artifact.ParseDocument(command.Document)
		placement.Files = append(placement.Files, shared.File{Path: promptsDirectory + "/" + command.Name + ".md", Kind: artifact.KindCommand, Content: []byte(document.Body)})
	}
	steering := false
	if skip, native := shared.InstructionSkip(input); native {
		placement.Skipped = append(placement.Skipped, skip)
	} else if text := input.Artifacts.InstructionText(); text != "" {
		steering = true
		placement.Files = append(placement.Files, shared.File{Path: steeringFile, Kind: artifact.KindInstruction, Content: []byte(steeringFrontPage + text)})
	}
	resources := resourcesOf(input.Root, len(skills) > 0, steering)
	for _, agent := range input.Artifacts.Agents {
		placement.Files = append(placement.Files, placeAgent(agent, resources))
	}
	if len(placement.Files) > 0 {
		placement.Files = append(placement.Files, shared.File{Path: settingsFile, Kind: artifact.KindSupport, Content: []byte(settings)})
		placement.Environment = []string{homeVariable + "=" + input.Root}
		placement.RootWritable = true
	}
	return shared.Finish(placement)
}

// resourcesOf lists the files an agent loads from the home.
func resourcesOf(root string, skills, steering bool) []string {
	resources := []string{}
	if skills {
		resources = append(resources, "file://"+shared.Join(root, skillsDirectory)+"/**/SKILL.md")
	}
	if steering {
		resources = append(resources, "file://"+shared.Join(root, steeringFile))
	}
	return resources
}

// placeAgent writes an agent as the JSON file Kiro reads for a custom agent.
func placeAgent(agent artifact.Agent, resources []string) shared.File {
	document, _ := artifact.ParseDocument(agent.Document)
	content, _ := json.MarshalIndent(struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Prompt      string   `json:"prompt"`
		Resources   []string `json:"resources"`
	}{agent.Name, document.Fields["description"], strings.TrimSpace(document.Body), resources}, "", "  ")
	return shared.File{Path: agentsDirectory + "/" + agent.Name + ".json", Kind: artifact.KindAgent, Content: append(content, '\n')}
}
