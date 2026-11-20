// Package claude is the harness leaf for Claude Code. It turns a request into
// a hermetic invocation: no settings from the developer's machine, no MCP
// servers, no persisted session and no browser, so a run scores the skill and
// not the local environment.
package claude

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

// Name is the harness name.
const Name = "claude"

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

	instructionFile = "CLAUDE.md"
)

// Harness is the Claude Code leaf.
type Harness struct{}

// New returns the leaf.
func New() Harness { return Harness{} }

// Name returns the harness name.
func (Harness) Name() string { return Name }

// Binary returns the program name.
func (Harness) Binary() string { return "claude" }

// isolated lists the flags that keep a run hermetic.
func isolated(format string) []string {
	return []string{
		"-p", "--output-format", format,
		"--setting-sources", "",
		"--strict-mcp-config", "--mcp-config", `{"mcpServers":{}}`,
		"--no-session-persistence",
		"--no-chrome",
	}
}

// stream lists the flags of a run that streams its events.
func stream() []string {
	return append(isolated("stream-json"), "--verbose", "--include-partial-messages")
}

func limits(req harness.Request, turns int) []string {
	args := []string{}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.Effort != "" {
		args = append(args, "--effort", string(req.Effort))
	}
	if req.Caps.BudgetUSD > 0 {
		args = append(args, "--max-budget-usd", strconv.FormatFloat(req.Caps.BudgetUSD, 'f', -1, 64))
	}
	return append(args, "--max-turns", strconv.Itoa(turns))
}

// stagePlugins stages each offered plugin in the workspace with only the
// offered skills, and returns the stage steps and the plugin directories.
func stagePlugins(workspace string, offers []harness.Offer) ([]harness.Stage, []string, error) {
	var stages []harness.Stage
	var dirs []string
	for i, offer := range offers {
		dir := path.Join(workspace, fmt.Sprintf("plugin-%d", i))
		skills := make([]string, 0, len(offer.SkillDirs))
		for _, skill := range offer.SkillDirs {
			name := path.Base(skill)
			stages = append(stages, harness.Stage{Kind: harness.StageCopy, Src: skill, Dst: path.Join(dir, "skills", name)})
			skills = append(skills, "./skills/"+name)
		}
		manifest, err := json.Marshal(map[string]any{"name": offer.Name, "skills": skills})
		if err != nil {
			return nil, nil, err
		}
		stages = append(stages, harness.Stage{
			Kind: harness.StageWrite, Dst: path.Join(dir, ".claude-plugin", "plugin.json"), Content: string(manifest),
		})
		dirs = append(dirs, dir)
	}
	return stages, dirs, nil
}

func pluginFlags(dirs []string) []string {
	var args []string
	for _, dir := range dirs {
		args = append(args, "--plugin-dir", dir)
	}
	return args
}

// Plan derives the invocation of a request.
func (Harness) Plan(req harness.Request, workspace string) (harness.Plan, error) {
	plan := harness.Plan{Dir: workspace, Prompt: req.Prompt}
	switch req.Role {
	case harness.RoleTrigger:
		stages, dirs, err := stagePlugins(workspace, req.Offer)
		if err != nil {
			return plan, err
		}
		plan.Stage = stages
		plan.Args = append(stream(), limits(req, triggerTurns)...)
		plan.Args = append(plan.Args, "--system-prompt", triggerSystem, "--tools", "Skill")
		plan.Args = append(plan.Args, pluginFlags(dirs)...)
	case harness.RoleGenerate:
		turns := req.Caps.MaxTurns
		if turns == 0 {
			turns = defaultTurns
		}
		plan.Args = append(stream(), limits(req, turns)...)
		plan.Args = append(plan.Args, "--system-prompt", generateSystem)
		if req.WithSkill {
			stages, dirs, err := stagePlugins(workspace, req.Offer)
			if err != nil {
				return plan, err
			}
			plan.Stage = stages
			plan.Args = append(plan.Args, "--tools", "Skill,Read")
			plan.Args = append(plan.Args, pluginFlags(dirs)...)
			plan.Prompt = "/" + req.Target + " " + req.Prompt
		} else {
			plan.Args = append(plan.Args, "--tools", "Read", "--disallowedTools=Skill")
		}
	case harness.RoleJudge:
		schema, err := judgeSchema(req.Assertions)
		if err != nil {
			return plan, err
		}
		plan.Args = append(isolated("json"), "--tools", "")
		if req.Caps.BudgetUSD > 0 {
			plan.Args = append(plan.Args, "--max-budget-usd", strconv.FormatFloat(req.Caps.BudgetUSD, 'f', -1, 64))
		}
		plan.Args = append(plan.Args, "--max-turns", "1", "--system-prompt", judgeSystem, "--json-schema", schema)
		if req.Model != "" {
			plan.Args = append(plan.Args, "--model", req.Model)
		}
	case harness.RoleAgent:
		turns := req.Caps.MaxTurns
		if turns == 0 {
			turns = defaultTurns
		}
		plan.Args = append(stream(), limits(req, turns)...)
		// Bash is on offer but not allowed, so the run asks for it and the
		// harness denies it: the attempt shows in the stream without a command
		// ever running on the host.
		plan.Args = append(plan.Args,
			"--tools", "Read,Write,Edit,Glob,Grep,Bash",
			"--allowedTools", "Read,Write,Edit,Glob,Grep")
		if req.Instructions != "" {
			plan.Stage = append(plan.Stage, harness.Stage{
				Kind: harness.StageWrite, Dst: path.Join(workspace, instructionFile), Content: req.Instructions,
			})
		}
	default:
		return plan, fmt.Errorf("claude has no invocation for role %q", req.Role)
	}
	return plan, nil
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
