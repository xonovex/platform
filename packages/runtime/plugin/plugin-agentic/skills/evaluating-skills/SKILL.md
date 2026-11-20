---
name: evaluating-skills
description: "Use when testing by model runs if a skill triggers or helps: trigger evals, output evals with and without it, or a trim ablation. Triggers on eval-trigger.json, eval-output.json, or trigger rate, even when the user doesn't say 'eval'."
compatibility: "The agent-eval command. Live trigger and output evals require a credentialed Claude or Codex CLI and network access."
allowed-tools: "Read Bash(agent-eval:*) Bash(claude:*) Bash(codex:*)"
---

# Skill Evaluation

Measure a skill with model runs: whether its description routes, whether it improves the answer, and whether a trim lost a fact the model needs. The writing of the skill belongs to **authoring-skills**.

## Inputs

- The skill directory with `SKILL.md`, `eval-trigger.json` and `eval-output.json`
- A credentialed `claude` or `codex` CLI for live runs, and a spend limit for each run

## Essentials

- **Evals before docs** - Build the trigger and output evals for the gap first, then write the minimum that passes them; iterate observe, revise, rerun
- **Trigger rate** - Realistic positives, near-miss negatives, a train/validation split and several runs per query, see [references/evaluating-triggers.md](references/evaluating-triggers.md)
- **Output quality** - Binary assertions graded with evidence, matched runs with and without the skill, one results document per run, see [references/evaluating-outputs.md](references/evaluating-outputs.md)
- **Trim by delta** - Cut what the weakest target model already knows, then prove with an ablation that no essential fact was lost, see [references/optimize.md](references/optimize.md)
- **Measure the repository copy** - Pass the local plugin directory, so the with-skill arm reads the files under edit and not an installed release

## Commands

- `agent-eval trigger <skill>`: run the trigger queries against one skill in isolation (needs a `claude` or `codex` CLI in PATH)
- `agent-eval routing [plugin]`: run the catalog queries with the competing skills loaded together
- `agent-eval output <skill>`: run the output evals with and without the skill; writes the pass rate, tokens, and duration of each arm and the tier verdict to the results directory

Each command takes `--harness claude|codex`; Claude is the default. `--dry-run` plans the work without a model run, `--resume` continues an incomplete run, and `--max-model-runs` and `--max-cost-usd` cap one invocation.

## Gotchas

- Each live run spends money: count the runs of a batch and set the spend limit before it starts
- An infrastructure failure invalidates the whole run; never report a partial result as a pass
- Tune a description on the train split only: a validation query that shaped the description no longer measures generalization

## Progressive Disclosure

- Read [references/evaluating-triggers.md](references/evaluating-triggers.md) - Load when writing trigger queries or measuring and tuning the trigger rate
- Read [references/evaluating-outputs.md](references/evaluating-outputs.md) - Load when writing output evals and assertions, or grading and comparing runs
- Read [references/optimize.md](references/optimize.md) - Load when trimming a skill to its delta over baseline knowledge and checking the trim with an ablation
