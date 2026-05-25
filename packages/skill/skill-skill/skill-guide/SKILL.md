---
name: skill-guide
description: "Use when authoring, reviewing, extracting, merging, simplifying, or validating Agent Skills (SKILL.md plus references / scripts / assets), or when auditing, splitting, de-duplicating, or tiering a whole set of skills. Triggers on edits under a skills directory, on prompts about creating a new skill, progressive disclosure, reference files, pattern extraction, merging or assimilating skills, simplification to bullet format, validation against the Agent Skills spec, description tuning, evaluating trigger rate / output quality, or making a catalog composable (one owner per concept, cross-references, general→language→framework tiers) — even when the user doesn't say 'skill'."
---

# Skill Guidelines Management

Author, extract, merge, simplify, and validate Agent Skills following the Agent Skills spec and authoring best practices.

## Spec Constraints

- `name`: 1-64 chars, lowercase kebab-case; no consecutive/leading/trailing hyphens; not the reserved words `anthropic`/`claude`; no XML tags; must match parent dir
- `description`: 1-1024 chars; imperative "Use when…" + explicit triggers; third person (no "I can…"/"You can…")
- Body: <500 lines / ~5000 tokens; push detail to `references/`
- Reference files: one level deep under `references/`, kebab-case filenames
- **Optional frontmatter:** `license` (string), `compatibility` (≤500 chars; declare network/runtime needs), `metadata` (string→string map), `allowed-tools` (experimental, space-separated allowlist, e.g. `Bash(git:*) Read`)
- **Progressive-disclosure budget** — discovery sees only name+description (~100 tokens), so the description alone decides routing; `SKILL.md` loads on activation (keep ≤ ~500 lines / 5k tokens); anything needed <~20% of the time belongs in `references/`, loaded on demand

## Core Principles

- **Progressive Disclosure** — SKILL.md contains essentials; `references/*` contains depth, loaded on demand
- **Project Independence** — remove project-specific paths, names, domains
- **Composable split** — one concept has one owner skill; prefer small mix-and-match skills, cross-reference others by name instead of duplicating, and generalize anything not inherently language/API-specific into a general skill that specific skills link to for the "why", see [references/composability.md](references/composability.md)
- **Design to coexist** — a skill is one capability among many loaded together; it must work alongside others, never assume it is the only one, and link up the general→language→framework tiers without the general tier depending on a specific one
- **Routing-first descriptions** — the description is the router (discovery sees only name+description); tune trigger words and Skip clauses, and debug mis-routes by asking "which skill did you use?", see [references/writing-descriptions.md](references/writing-descriptions.md)
- **Sources in SOURCES.md** — cite provenance only in `SOURCES.md`; never name authors, companies, talks, books, or blogs inside `SKILL.md` or `references/*` (tool/API/standard names are fine)
- **Bullet Format** — `- **Rule** - Brief 5-10 word how-to, see [references/<topic>.md](references/<topic>.md)`
- **Style Consistency** — match existing skill patterns in structure and voice
- **Add what the agent lacks; omit what it knows** — no general-knowledge filler
- **Defaults over menus** — one default, alternatives mentioned briefly
- **Procedures over declarations** — teach the approach/steps, not a one-off answer

## Skill Structure

- **SKILL.md** — frontmatter, essentials (3-7 bullets), Gotchas, one example, progressive disclosure links with load-when triggers
- **references/\*.md** — statement, rationale, how to apply, examples, counter-examples — one topic per file
- **scripts/** (optional) — bundled executables for repeated work
- **assets/** (optional) — templates and data files

## Gotchas

- Reference files are **not** SKILL.md files — they don't need frontmatter or their own Gotchas section
- Load-when triggers live in the parent SKILL.md's progressive-disclosure list, not at the top of each reference
- Skill name must equal the parent directory name exactly — renaming a skill means renaming the dir too
- `description` is a hard 1024-char limit; it tends to grow during iteration, so re-check after each edit

## Scripts

PEP 723 self-contained Python scripts (run with `uv run <script>` — `uv` creates an isolated env on first run, no manual install step):

- `scripts/validate.py <skill-dir>` — spec / quality / harness-neutrality audit (read-only; exits non-zero on errors)
- `scripts/eval-triggers.py <queries.json> <skill-name>` — run trigger-eval queries against a skill (Claude Code reference implementation; requires `claude` CLI in PATH)
- `scripts/eval-outputs.py <evals.json> <skill-name>` — run output-quality evals with-skill vs without-skill; writes per-arm pass rate / tokens / duration + `benchmark.json` (requires `claude` CLI in PATH)
- `scripts/update-skill-from-source.py <skill-dir>` — audit a skill's `SOURCES.md` for drift: staleness vs `Last reviewed`, dangling provenance, source→reference mapping; `--fetch` to check URLs, `--mark-reviewed` to stamp the date after review (read-only by default)

Cross-platform (macOS / Linux / Windows wherever `uv` is installed). Install `uv` with `brew install uv` or `curl -LsSf https://astral.sh/uv/install.sh | sh`.

## Progressive Disclosure

Single index of every reference; each entry names the operation/concept and when to load it.

- Read [references/create.md](references/create.md) - Load when creating a new skill from a document, URL, or task description
- Read [references/extract-from-codebase.md](references/extract-from-codebase.md) - Load when extracting patterns from this codebase into a skill
- Read [references/merge.md](references/merge.md) - Load when porting elements from one skill into another
- Read [references/simplify-skill.md](references/simplify-skill.md) - Load when condensing a verbose SKILL.md to bullet format
- Read [references/simplify-reference.md](references/simplify-reference.md) - Load when condensing or merging verbose reference files
- Read [references/validate.md](references/validate.md) - Load when auditing a SKILL.md against the spec
- Read [references/composability.md](references/composability.md) - Load when deciding skill boundaries, owners, tiers, or whether to generalize vs link a concept
- Read [references/catalog-audit.md](references/catalog-audit.md) - Load when auditing, splitting, or de-duplicating a set of skills onto the composable split
- Read [references/guideline-skills.md](references/guideline-skills.md) - Load when creating a coding-guideline / style-rule skill (topic categories, do/don't patterns, reference shape)
- Read [references/writing-descriptions.md](references/writing-descriptions.md) - Load when authoring or rewriting a `description` field (writing principles, before/after)
- Read [references/evaluating-triggers.md](references/evaluating-triggers.md) - Load when verifying or iterating on trigger rate (eval queries, train/validation split, optimization loop)
- Read [references/evaluating-outputs.md](references/evaluating-outputs.md) - Load when verifying or iterating on output quality (test cases, assertions, grading)
- Read [references/using-scripts.md](references/using-scripts.md) - Load when the skill needs an executable component (one-off commands, self-contained scripts)
- Read [references/instruction-patterns.md](references/instruction-patterns.md) - Load when designing the body of a workflow-heavy skill (templates, checklists, validation loops, plan-validate-execute)
