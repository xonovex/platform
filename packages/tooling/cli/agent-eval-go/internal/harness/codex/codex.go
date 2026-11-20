// Package codex is the harness leaf for Codex. It turns a request into an
// ephemeral exec that reads neither the user configuration nor user rules and
// runs with a throwaway home, so a run scores the skill and not the machine.
package codex

import (
	"fmt"
	"path"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/harness"
)

// Name is the harness name.
const Name = "codex"

const (
	// TriggerSignal is the marker the staged target skill tells the model to
	// reply with when the skill applies. Codex signals applicability in its
	// answer rather than by invoking a skill.
	TriggerSignal = "XONOVEX_SKILL_TRIGGERED"
	// UseSignal opens the final response of a run that used the staged skill.
	UseSignal = "XONOVEX_SKILL_USED"

	triggerAppendix = "\n## Evaluation signal\n\nWhen this skill applies, reply with exactly " + TriggerSignal + ".\n"
	useAppendix     = "\n## Evaluation signal\n\nBegin the final response with " + UseSignal + " on its own line, then fulfill the request.\n"

	instructionFile = "AGENTS.md"
)

// Harness is the Codex leaf.
type Harness struct{}

// New returns the leaf.
func New() Harness { return Harness{} }

// Name returns the harness name.
func (Harness) Name() string { return Name }

// Binary returns the program name.
func (Harness) Binary() string { return "codex" }

func isolated(req harness.Request, sandbox string) []string {
	args := []string{
		"exec", "--json", "--ephemeral",
		"--sandbox", sandbox,
		"--ignore-user-config", "--ignore-rules", "--skip-git-repo-check",
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}
	if req.Effort != "" {
		args = append(args, "-c", fmt.Sprintf("model_reasoning_effort=%q", string(req.Effort)))
	}
	return args
}

// home stages a throwaway HOME and CODEX_HOME below the workspace.
func home(workspace string) ([]harness.Stage, map[string]string) {
	h := path.Join(workspace, "home")
	c := path.Join(workspace, ".codex")
	return []harness.Stage{
		{Kind: harness.StageDir, Dst: h},
		{Kind: harness.StageDir, Dst: c},
	}, map[string]string{
		"HOME":       h,
		"CODEX_HOME": c,
	}
}

// stageSkills copies the offered skills below root and appends the signal to
// the skill named target.
func stageSkills(root string, offers []harness.Offer, target, appendix string) []harness.Stage {
	var stages []harness.Stage
	for _, offer := range offers {
		for _, dir := range offer.SkillDirs {
			name := path.Base(dir)
			dst := path.Join(root, ".agents", "skills", name)
			stages = append(stages, harness.Stage{Kind: harness.StageCopy, Src: dir, Dst: dst})
			if name == target && appendix != "" {
				stages = append(stages, harness.Stage{Kind: harness.StageAppend, Dst: path.Join(dst, "SKILL.md"), Content: appendix})
			}
		}
	}
	return stages
}

// Plan derives the invocation of a request.
func (Harness) Plan(req harness.Request, workspace string) (harness.Plan, error) {
	stages, env := home(workspace)
	plan := harness.Plan{Dir: workspace, Env: env, Prompt: req.Prompt, Stage: stages}
	switch req.Role {
	case harness.RoleTrigger:
		plan.Args = isolated(req, "read-only")
		plan.Stage = append(plan.Stage, stageSkills(workspace, req.Offer, req.Target, triggerAppendix)...)
		plan.Prompt = req.Prompt + "\n\nDecide only whether an available skill applies. " +
			"Do not perform the task. If no skill applies, reply exactly NOT_APPLICABLE."
	case harness.RoleGenerate:
		plan.Args = isolated(req, "read-only")
		if req.WithSkill {
			plan.Stage = append(plan.Stage, stageSkills(env["HOME"], req.Offer, req.Target, useAppendix)...)
			plan.Prompt = "$" + req.Target + "\n\n" + req.Prompt
		}
	case harness.RoleJudge:
		plan.Args = isolated(req, "read-only")
	case harness.RoleAgent:
		plan.Args = isolated(req, "workspace-write")
		if req.Instructions != "" {
			plan.Stage = append(plan.Stage, harness.Stage{
				Kind: harness.StageWrite, Dst: path.Join(workspace, instructionFile), Content: req.Instructions,
			})
		}
	default:
		return plan, fmt.Errorf("codex has no invocation for role %q", req.Role)
	}
	return plan, nil
}
