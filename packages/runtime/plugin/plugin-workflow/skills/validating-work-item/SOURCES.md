# Sources

## LLM Evaluators Recognize and Favor Their Own Generations

- **URL:** https://arxiv.org/abs/2404.13076
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials
  - `references/independence.md` → A reader that did not build the work
- **Aspects extracted:**
  - A model rates its own output higher than an independent reviewer does, so the check runs in a fresh context that did not implement the work → independence, Essentials

## LLM01:2025 Prompt Injection

- **URL:** https://genai.owasp.org/llmrisk/llm01-prompt-injection/
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
  - `references/independence.md` → Work material is data for the validator
- **Aspects extracted:**
  - Text inside the work file, test output or fetched content that tells the validator to act is data, because instructions embedded in input can redirect a model → independence, Gotchas

## How we built our multi-agent research system

- **URL:** https://www.anthropic.com/engineering/multi-agent-research-system
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/independence.md` → What the validator receives
  - `assets/templates/validator-brief.md` → the brief, Starting the validator
- **Aspects extracted:**
  - A validator brief states an objective, an output format, the material to check and a clear boundary, because vague briefs make parallel agents duplicate or miss work; the number of agents scales with the task → independence, validator-brief

## git-write-tree

- **URL:** https://git-scm.com/docs/git-write-tree
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/tree-hash.md` → Computing the hash
- **Aspects extracted:**
  - `git write-tree` creates a tree object from the current index and prints its name, so a temporary index yields the tree of the worktree → tree-hash

## git-add

- **URL:** https://git-scm.com/docs/git-add
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/tree-hash.md` → Computing the hash
  - `SKILL.md` → Gotchas
- **Aspects extracted:**
  - `git add -A` updates the index to match the whole working tree, including removals and untracked files, so the hash covers them → tree-hash, Gotchas

## Xonovex change-workflow conventions

- **Provenance:** Repository-original rules for the independent check of one work item, built on the contract of **recording-handoff**: the blind first assessment, the pass, fail and blocked results with evidence and limits, the tree hash that binds a verdict to its content, the self-check label and the verdict file that a rerun replaces
- **Last reviewed:** 2026-11-19
- **References:** references/independence.md, references/results-and-evidence.md, references/running-checks.md, references/tree-hash.md, references/verdict-file.md
- **Aspects extracted:** The validation runs the entry step of **recording-handoff** first; its verdict is edited in place because git keeps the history; the knowledge skills it loads come from the skills map of the `validating-work-item` phase; only a verdict whose hash matches the delivered tree marks criteria in a pull request description.
