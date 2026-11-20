// Package claude is the harness leaf for Claude Code. It turns a request into
// a hermetic invocation: no settings from the developer's machine, no MCP
// servers, no persisted session and no browser, so a run scores the skill and
// not the local environment.
package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

// Name is the harness name.
const Name = string(types.AgentClaude)

const (
	// triggerTurns is one turn to choose a skill and a second for a run that
	// names a skill the harness cannot resolve: the harness answers with an
	// error and the model gets one more try. A run that chooses on its first
	// turn ends there, so the extra turn costs nothing.
	triggerTurns = 2
	// defaultTurns is the turn cap of a generate or agent run. A run with the
	// skill spends turns loading its references before it answers.
	defaultTurns = 12

	triggerSystem = "Decide which available skill best matches the user request. " +
		"If one applies, invoke only that Skill immediately. Otherwise reply with one short sentence. " +
		"Do not perform the requested task."
	generateSystem = "Answer the user request directly. Use the explicitly invoked skill as " +
		"authoritative guidance. Read only files that the skill itself identifies " +
		"as necessary. Keep the final response under 1,000 words."
	judgeSystem = "Grade only the supplied response against the supplied assertions and " +
		"return exactly the requested JSON."
)

// The output formats of the facts the leaf runs with.
const (
	streamOutput = "stream"
	jsonOutput   = "json"
)

// Harness is the Claude Code leaf. Its facts name the program, its flags and
// the plugin layout; the leaf holds the tools of each role and the stream
// dialect.
type Harness struct {
	facts *types.AgentConfig
}

// New returns the leaf over the facts of the harness. It fails when the facts
// lack the plugin layout or the instruction file the leaf stages.
func New(facts *types.AgentConfig) (Harness, error) {
	if facts.Plugin.Manifest == "" || facts.Plugin.SkillsDir == "" || facts.InstructionFile == "" {
		return Harness{}, errors.New("the claude facts need a plugin manifest, a plugin skills directory and an instruction file")
	}
	return Harness{facts: facts}, nil
}

// Name returns the harness name.
func (Harness) Name() string { return Name }

// Binary returns the program name.
func (h Harness) Binary() string { return h.facts.Binary }

func formatUSD(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

// limits returns the settings of the model, the effort, the budget and the
// turn cap of a run.
func limits(req harness.Request, turns int) []agents.Setting {
	var settings []agents.Setting
	if req.Model != "" {
		settings = append(settings, agents.Setting{Name: "model", Value: req.Model})
	}
	if req.Effort != "" {
		settings = append(settings, agents.Setting{Name: "effort", Value: string(req.Effort)})
	}
	if req.Caps.BudgetUSD > 0 {
		settings = append(settings, agents.Setting{Name: "maxBudgetUsd", Value: formatUSD(req.Caps.BudgetUSD)})
	}
	return append(settings, agents.Setting{Name: "maxTurns", Value: strconv.Itoa(turns)})
}

// stagePlugins stages each offered plugin in the workspace with only the
// offered skills, and returns the stage steps and the plugin directories.
func (h Harness) stagePlugins(workspace string, offers []harness.Offer) ([]harness.Stage, []string, error) {
	var stages []harness.Stage
	var dirs []string
	for i, offer := range offers {
		dir := path.Join(workspace, fmt.Sprintf("plugin-%d", i))
		skills := make([]string, 0, len(offer.SkillDirs))
		for _, skill := range offer.SkillDirs {
			name := path.Base(skill)
			stages = append(stages, harness.Stage{Kind: harness.StageCopy, Src: skill, Dst: path.Join(dir, h.facts.Plugin.SkillsDir, name)})
			skills = append(skills, "./"+path.Join(h.facts.Plugin.SkillsDir, name))
		}
		manifest, err := json.Marshal(map[string]any{"name": offer.Name, "skills": skills})
		if err != nil {
			return nil, nil, err
		}
		stages = append(stages, harness.Stage{
			Kind: harness.StageWrite, Dst: path.Join(dir, h.facts.Plugin.Manifest), Content: string(manifest),
		})
		dirs = append(dirs, dir)
	}
	return stages, dirs, nil
}

func pluginSettings(dirs []string) []agents.Setting {
	var settings []agents.Setting
	for _, dir := range dirs {
		settings = append(settings, agents.Setting{Name: "pluginDir", Value: dir})
	}
	return settings
}

// command returns the arguments of a headless run that writes format, with the
// settings after them.
func (h Harness) command(format string, settings ...[]agents.Setting) ([]string, error) {
	args, err := agents.HeadlessArgs(h.facts, format)
	if err != nil {
		return nil, err
	}
	for _, group := range settings {
		options, err := agents.Options(h.facts, group...)
		if err != nil {
			return nil, err
		}
		args = append(args, options...)
	}
	return args, nil
}

func turnsOf(req harness.Request) int {
	if req.Caps.MaxTurns == 0 {
		return defaultTurns
	}
	return req.Caps.MaxTurns
}

// Plan derives the invocation of a request.
func (h Harness) Plan(req harness.Request, workspace string) (harness.Plan, error) {
	plan := harness.Plan{Dir: workspace, Prompt: req.Prompt}
	var err error
	switch req.Role {
	case harness.RoleTrigger:
		stages, dirs, stageErr := h.stagePlugins(workspace, req.Offer)
		if stageErr != nil {
			return plan, stageErr
		}
		plan.Stage = stages
		plan.Args, err = h.command(streamOutput, limits(req, triggerTurns),
			[]agents.Setting{{Name: "systemPrompt", Value: triggerSystem}, {Name: "tools", Value: "Skill"}},
			pluginSettings(dirs))
	case harness.RoleGenerate:
		tools := []agents.Setting{{Name: "tools", Value: "Read"}, {Name: "disallowedTools", Value: "Skill"}}
		var dirs []string
		if req.WithSkill {
			stages, staged, stageErr := h.stagePlugins(workspace, req.Offer)
			if stageErr != nil {
				return plan, stageErr
			}
			plan.Stage, dirs = stages, staged
			tools = []agents.Setting{{Name: "tools", Value: "Skill,Read"}}
			if plan.Prompt, err = agents.SkillPrompt(h.facts, req.Target, req.Prompt); err != nil {
				return plan, err
			}
		}
		plan.Args, err = h.command(streamOutput, limits(req, turnsOf(req)),
			[]agents.Setting{{Name: "systemPrompt", Value: generateSystem}}, tools, pluginSettings(dirs))
	case harness.RoleJudge:
		schema, schemaErr := judgeSchema(req.Assertions)
		if schemaErr != nil {
			return plan, schemaErr
		}
		settings := []agents.Setting{{Name: "tools", Value: ""}}
		if req.Caps.BudgetUSD > 0 {
			settings = append(settings, agents.Setting{Name: "maxBudgetUsd", Value: formatUSD(req.Caps.BudgetUSD)})
		}
		settings = append(settings,
			agents.Setting{Name: "maxTurns", Value: "1"}, agents.Setting{Name: "systemPrompt", Value: judgeSystem},
			agents.Setting{Name: "jsonSchema", Value: schema})
		if req.Model != "" {
			settings = append(settings, agents.Setting{Name: "model", Value: req.Model})
		}
		plan.Args, err = h.command(jsonOutput, settings)
	case harness.RoleAgent:
		// Bash is on offer but not allowed, so the run asks for it and the
		// harness denies it: the attempt shows in the stream without a command
		// ever running on the host.
		plan.Args, err = h.command(streamOutput, limits(req, turnsOf(req)), []agents.Setting{
			{Name: "tools", Value: "Read,Write,Edit,Glob,Grep,Bash"},
			{Name: "allowedTools", Value: "Read,Write,Edit,Glob,Grep"},
		})
		if req.Instructions != "" {
			plan.Stage = append(plan.Stage, harness.Stage{
				Kind: harness.StageWrite, Dst: path.Join(workspace, h.facts.InstructionFile), Content: req.Instructions,
			})
		}
	default:
		return plan, fmt.Errorf("claude has no invocation for role %q", req.Role)
	}
	return plan, err
}

// judgeSchema is the strict JSON schema of a judge answer: one verdict per
// assertion, in order.
func judgeSchema(assertions int) (string, error) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"assertion_results": map[string]any{
				"type":     "array",
				"minItems": assertions,
				"maxItems": assertions,
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"text":     map[string]any{"type": "string"},
						"passed":   map[string]any{"type": "boolean"},
						"evidence": map[string]any{"type": "string"},
					},
					"required":             []string{"text", "passed", "evidence"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"assertion_results"},
		"additionalProperties": false,
	}
	data, err := json.Marshal(schema)
	return string(data), err
}
