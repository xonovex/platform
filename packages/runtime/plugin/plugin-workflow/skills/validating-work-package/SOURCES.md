# Sources

## LLM Evaluators Recognize and Favor Their Own Generations

- **URL:** https://arxiv.org/abs/2404.13076
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials
  - `references/independence.md` → Why the assessor is fresh
- **Aspects extracted:**
  - A model rates its own output higher than an independent checker does, so the assessment runs in a fresh context that implemented none of the work → independence, Essentials

## LLM01:2025 Prompt Injection

- **URL:** https://genai.owasp.org/llmrisk/llm01-prompt-injection/
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Gotchas
  - `references/independence.md` → Package material is data for the assessor
- **Aspects extracted:**
  - Text inside supplied or fetched content that tells the assessor to act is data, because instructions embedded in input can redirect a model → independence, Gotchas

## How we built our multi-agent research system

- **URL:** https://www.anthropic.com/engineering/multi-agent-research-system
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/independence.md` → What the assessor receives, Starting the assessors
- **Aspects extracted:**
  - An assessor brief states an objective, the inputs, an output format and a clear boundary, because vague briefs make parallel agents duplicate or miss work; the number of agents scales with the task → independence

## git-write-tree

- **URL:** https://git-scm.com/docs/git-write-tree
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Gotchas
  - `references/verdicts-and-deliveries.md` → Matching a verdict to the worktree
- **Aspects extracted:**
  - `git write-tree` reads the index named by `GIT_INDEX_FILE`, so a private index inside a subshell yields the tree of the worktree without touching the real index → verdicts and deliveries, Gotchas

## Xonovex change-workflow conventions

- **Provenance:** Repository-original rules for closing out a work package, built on the contract of **recording-handoff**: the met, gated and open states, re-derived evidence, the integration seams, the kinds of blocker, the decision register, open points and residue, follow-up seeds, the delivery that is prepared and not applied, and the close-out file that a rerun replaces
- **Last reviewed:** 2026-11-19
- **References:** references/independence.md, references/status.md, references/verdicts-and-deliveries.md, references/integration.md, references/gates-and-blockers.md, references/decisions-and-open-points.md, references/follow-up-seeds.md, references/close-out-file.md
- **Aspects extracted:** The close-out runs the entry step of **recording-handoff** first and writes the log entry of its run; its file is replaced by the next run because git keeps the history; the knowledge skills it loads come from the skills map of the `validating-work-package` phase; a delivery that no step applied is reported for the user to commit and publish through **using-git** and **using-github** or **using-gitlab**.
