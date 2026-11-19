---
name: evaluating-skills
description: "Use when measuring an Agent Skill with model runs: trigger evals, output evals with and without the skill, or an ablation after a trim. Triggers on eval-queries.json, evals.json, trigger rate, benchmark.json, even when the user doesn't say 'eval'."
compatibility: "Python 3.11+ with uv runs the bundled evaluation scripts. Live trigger and output evals require a credentialed Claude or Codex CLI and network access."
allowed-tools: "Read Bash(uv:*) Bash(claude:*) Bash(codex:*)"
---

# Skill Evaluation

Measure a skill with model runs: whether its description routes, whether it improves the answer, and whether a trim lost a fact the model needs. The writing of the skill belongs to **authoring-skills**.

## Inputs

- The skill directory with `SKILL.md`, `eval-queries.json` and `evals.json`
- A credentialed `claude` or `codex` CLI for live runs, and a spend limit for each run

## Essentials

- **Evals before docs** - Build the trigger and output evals for the gap first, then write the minimum that passes them; iterate observe, revise, rerun
- **Trigger rate** - Realistic positives, near-miss negatives, a train/validation split and several runs per query, see [references/evaluating-triggers.md](references/evaluating-triggers.md)
- **Output quality** - Binary assertions graded with evidence, matched runs with and without the skill, one `benchmark.json` per iteration, see [references/evaluating-outputs.md](references/evaluating-outputs.md)
- **Trim by delta** - Cut what the weakest target model already knows, then prove with an ablation that no essential fact was lost, see [references/optimize.md](references/optimize.md)
- **Measure the repository copy** - Pass the local plugin directory, so the with-skill arm reads the files under edit and not an installed release

## Scripts

The scripts use PEP 723 and run with `uv run <script>`:

- `scripts/eval-triggers.py <queries.json> <skill-name>`: run trigger-eval queries against a skill (Claude Code reference implementation; requires `claude` CLI in PATH)
- `scripts/eval-outputs.py <evals.json> <skill-name>`: run output-quality evals with and without the skill; writes the pass rate, tokens and duration of each arm and `benchmark.json` (requires `claude` CLI in PATH)

The repository's `moon-skill-eval-triggers` and `moon-skill-eval-outputs` commands accept `--harness claude|codex`. Claude remains the default; Codex runs with an isolated home, ignored user configuration/rules, a read-only sandbox, ephemeral sessions, and only the target skill in the activated arm.

## Gotchas

- Each live run spends money: count the runs of a batch and set the spend limit before it starts
- An infrastructure failure invalidates the whole run; never report a partial result as a pass
- Tune a description on the train split only: a validation query that shaped the description no longer measures generalization

## Progressive Disclosure

- Read [references/evaluating-triggers.md](references/evaluating-triggers.md) - Load when writing trigger queries or measuring and tuning the trigger rate
- Read [references/evaluating-outputs.md](references/evaluating-outputs.md) - Load when writing output evals and assertions, or grading and comparing runs
- Read [references/optimize.md](references/optimize.md) - Load when trimming a skill to its delta over baseline knowledge and checking the trim with an ablation
