// Package codex is the harness leaf for Codex. It turns a request into an
// ephemeral exec that reads neither the user configuration nor user rules and
// runs with a throwaway home, so a run scores the skill and not the machine.
package codex

import (
	"errors"
	"fmt"
	"path"

	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/agents"
	"github.com/xonovex/platform/packages/library/shared-agent-go/pkg/types"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

// Name is the harness name.
const Name = string(types.AgentCodex)

const (
	// TriggerSignal is the marker the staged target skill tells the model to
	// reply with when the skill applies. Codex signals applicability in its
	// answer rather than by invoking a skill.
	TriggerSignal = "XONOVEX_SKILL_TRIGGERED"
	// UseSignal opens the final response of a run that used the staged skill.
	UseSignal = "XONOVEX_SKILL_USED"

	triggerAppendix = "\n## Evaluation signal\n\nWhen this skill applies, reply with exactly " + TriggerSignal + ".\n"
	useAppendix     = "\n## Evaluation signal\n\nBegin the final response with " + UseSignal + " on its own line, then fulfill the request.\n"

	// streamOutput is the output format of every run.
	streamOutput = "stream"
	// The sandbox options of the facts: a read-only run and a run that may
	// write the workspace.
	readOnly       = "readOnly"
	workspaceWrite = "workspaceWrite"
)

// Harness is the Codex leaf. Its facts name the program, its flags, the
// configuration home and the skills layout; the leaf holds the signals of each
// role and the stream dialect.
type Harness struct {
	facts *types.AgentConfig
}

// New returns the leaf over the facts of the harness. It fails when the facts
// lack the configuration home, the skills directory or the instruction file the
// leaf stages.
func New(facts *types.AgentConfig) (Harness, error) {
	if facts.ConfigHome == nil || facts.SkillsDir == "" || facts.InstructionFile == "" {
		return Harness{}, errors.New("the codex facts need a configuration home, a skills directory and an instruction file")
	}
	return Harness{facts: facts}, nil
}

// Name returns the harness name.
func (Harness) Name() string { return Name }

// Binary returns the program name.
func (h Harness) Binary() string { return h.facts.Binary }

// command returns the arguments of a hermetic run in the sandbox option, with
// the model and the effort of the request.
func (h Harness) command(req harness.Request, sandbox string) ([]string, error) {
	args, err := agents.HeadlessArgs(h.facts, streamOutput)
	if err != nil {
		return nil, err
	}
	settings := []agents.Setting{{Name: sandbox}}
	if req.Model != "" {
		settings = append(settings, agents.Setting{Name: "model", Value: req.Model})
	}
	if req.Effort != "" {
		settings = append(settings, agents.Setting{Name: "effort", Value: string(req.Effort)})
	}
	options, err := agents.Options(h.facts, settings...)
	if err != nil {
		return nil, err
	}
	return append(args, options...), nil
}

// home stages a throwaway HOME and configuration home below the workspace.
func (h Harness) home(workspace string) ([]harness.Stage, map[string]string) {
	home := path.Join(workspace, "home")
	config := path.Join(workspace, h.facts.ConfigHome.Dir)
	return []harness.Stage{
		{Kind: harness.StageDir, Dst: home},
		{Kind: harness.StageDir, Dst: config},
	}, map[string]string{
		"HOME":                 home,
		h.facts.ConfigHome.Env: config,
	}
}

// stageSkills copies the offered skills below root and appends the signal to
// the skill named target.
func (h Harness) stageSkills(root string, offers []harness.Offer, target, appendix string) []harness.Stage {
	var stages []harness.Stage
	for _, offer := range offers {
		for _, dir := range offer.SkillDirs {
			name := path.Base(dir)
			dst := path.Join(root, h.facts.SkillsDir, name)
			stages = append(stages, harness.Stage{Kind: harness.StageCopy, Src: dir, Dst: dst})
			if name == target && appendix != "" {
				stages = append(stages, harness.Stage{Kind: harness.StageAppend, Dst: path.Join(dst, "SKILL.md"), Content: appendix})
			}
		}
	}
	return stages
}

// Plan derives the invocation of a request.
func (h Harness) Plan(req harness.Request, workspace string) (harness.Plan, error) {
	stages, env := h.home(workspace)
	plan := harness.Plan{Dir: workspace, Env: env, Prompt: req.Prompt, Stage: stages}
	var err error
	switch req.Role {
	case harness.RoleTrigger:
		plan.Args, err = h.command(req, readOnly)
		plan.Stage = append(plan.Stage, h.stageSkills(workspace, req.Offer, req.Target, triggerAppendix)...)
		plan.Prompt = req.Prompt + "\n\nDecide only whether an available skill applies. " +
			"Do not perform the task. If no skill applies, reply exactly NOT_APPLICABLE."
	case harness.RoleGenerate:
		plan.Args, err = h.command(req, readOnly)
		if req.WithSkill && err == nil {
			plan.Stage = append(plan.Stage, h.stageSkills(env["HOME"], req.Offer, req.Target, useAppendix)...)
			plan.Prompt, err = agents.SkillPrompt(h.facts, req.Target, req.Prompt)
		}
	case harness.RoleJudge:
		plan.Args, err = h.command(req, readOnly)
	case harness.RoleAgent:
		plan.Args, err = h.command(req, workspaceWrite)
		if req.Instructions != "" {
			plan.Stage = append(plan.Stage, harness.Stage{
				Kind: harness.StageWrite, Dst: path.Join(workspace, h.facts.InstructionFile), Content: req.Instructions,
			})
		}
	default:
		return plan, fmt.Errorf("codex has no invocation for role %q", req.Role)
	}
	return plan, err
}
