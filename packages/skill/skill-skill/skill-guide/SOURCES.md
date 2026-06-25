# Sources

## Agent Skills: Best Practices for Skill Creators

- **URL:** https://agentskills.io/skill-creation/best-practices.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Core Principles, Gotchas
  - `references/create.md` → Content Rules
  - `references/instruction-patterns.md` → all sections
  - `references/guideline-skills.md` → patterns
- **Aspects extracted:**
  - "Start from real expertise" — extracting from hands-on tasks and project artifacts → influenced `create.md` workflow and `extract-from-codebase.md`
  - "Spending context wisely" → "Add only what the agent lacks; omit what it knows" rule in Content Rules
  - "Aim for moderate detail" → 3-7 essentials guideline in `create.md` output structure
  - "Match specificity to fragility" → Content Rules
  - "Provide defaults, not menus" → Defaults-over-menus rule in `SKILL.md` Core Principles
  - "Favor procedures over declarations" → Content Rules
  - Gotchas section pattern → required Gotchas in skill template (Output Structure)
  - Templates for output format → `instruction-patterns.md` Templates section
  - Checklists for multi-step workflows → `instruction-patterns.md` Checklists section
  - Validation loops → `instruction-patterns.md` Validation Loops section
  - Plan-validate-execute → `instruction-patterns.md` Plan-Validate-Execute section
  - Bundling reusable scripts → `using-scripts.md`

## Agent Skills: Optimizing Skill Descriptions

- **URL:** https://agentskills.io/skill-creation/optimizing-descriptions.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `references/writing-descriptions.md` → Writing Principles, Anatomy, Before/After, Gotchas
  - `references/evaluating-triggers.md` → entire file
- **Aspects extracted:**
  - Imperative phrasing ("Use when...") → Writing Principles
  - User-intent-over-implementation rule → Writing Principles
  - "Be pushy on triggers" + "even when the user doesn't say X" pattern → Writing Principles
  - 1024-char limit → Spec Constraints (multiple files)
  - Trigger eval set design (~20 queries, 8-10 positive + 8-10 negative) → `evaluating-triggers.md` Trigger Eval Set
  - Near-miss negative queries → Strongest should-not-trigger queries
  - Realistic query content (file paths, personal context, casual language) → Making Eval Queries Realistic
  - Multi-run trigger rate measurement → Measuring Trigger Rate
  - Train / validation split methodology → Train / Validation Split
  - Optimization loop (5 iterations, pick best validation pass rate) → Optimization Loop

## Agent Skills: Using Scripts in Skills

- **URL:** https://agentskills.io/skill-creation/using-scripts.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `references/using-scripts.md` → entire file
- **Aspects extracted:**
  - One-off commands vs bundled scripts decision rule
  - Version pinning examples (npx@version, uvx@version, etc.)
  - `compatibility` frontmatter for runtime requirements
  - Self-contained scripts with inline deps: PEP 723 (Python), Deno `npm:`/`jsr:`, Bun auto-install, Ruby `bundler/inline`
  - Designing for agentic use: no interactive prompts, `--help`, helpful error messages, structured output (JSON/CSV), stdout vs stderr separation
  - Idempotency, closed-set inputs, `--dry-run`, meaningful exit codes, safe defaults, bounded output size

## Agent Skills: Evaluating Skill Output Quality

- **URL:** https://agentskills.io/skill-creation/evaluating-skills.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `references/evaluating-outputs.md` → entire file (incl. Automated Runner section)
  - `scripts/eval-outputs.py` → with-skill/without-skill arms, LLM-as-judge grading, `benchmark.json`
- **Aspects extracted:**
  - Test case shape (prompt, expected output, optional input files) → Test Case Shape
  - Reference-guided binary LLM-as-judge with evidence; with-skill vs without-skill arms in isolated runs → `scripts/eval-outputs.py` (method also informed by public SkillsBench / skill-creator output-eval practice)
  - `evals/evals.json` storage convention
  - Workspace layout (`iteration-N/`, `with_skill/`, `without_skill/`) → Workspace Layout
  - Run pair (with skill + baseline) → Running Eval Pairs
  - Fresh context per run → Workflow
  - Timing capture (`total_tokens`, `duration_ms`) → Running Eval Pairs
  - Good vs weak assertions, evidence requirement → Writing Assertions
  - Grading via script for mechanical checks, LLM for the rest → Grading
  - `benchmark.json` aggregation → Aggregating
  - Pattern analysis: always-pass, always-fail, with-vs-without delta, variance → Pattern Analysis
  - Human review + actionable feedback → Human Review
  - Iteration signals (failed assertions, feedback, transcripts) → Iteration Signals
  - "Stay lean, explain why, bundle repeated work" iteration principles

## Agent Skills: Specification

- **URL:** https://agentskills.io/specification.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Spec Constraints, Skill Structure
  - `references/validate.md` → Spec Constraints, Frontmatter Checks, Body Checks, Reference Checks
- **Aspects extracted:**
  - Frontmatter fields: required `name` (≤64 chars, kebab-case regex, matches parent dir), required `description` (≤1024 chars), optional `license` / `compatibility` (≤500 chars) / `metadata` / `allowed-tools`
  - Body recommended <500 lines / ~5000 tokens
  - Directory layout: `SKILL.md` + optional `scripts/`, `references/`, `assets/`
  - Progressive disclosure: metadata at startup, body on activation, resources on demand
  - File references one level deep, relative paths from skill root
  - Reference filename kebab-case convention

## Agent Skills: Equipping Agents for the Real World

- **URL:** https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `SKILL.md` → Core Principles (`Design to coexist`, `Routing-first descriptions`), Spec Constraints (progressive-disclosure budget)
  - `references/composability.md` → entire file
  - `references/catalog-audit.md` → entire file
- **Aspects extracted:**
  - "Composable by design" — a skill should work well alongside others, not assume it is the only capability → Design-to-coexist principle and one-owner-per-concept rule
  - Progressive disclosure as token budget (metadata ~100 tokens → SKILL.md → references on demand) → progressive-disclosure budget note and routing-first descriptions
  - General → language → framework tiering as the composition model → `composability.md` tiered model and `catalog-audit.md` owner-selection

## Agent Skills: Security & Least Privilege (synthesized)

- **URL:** https://agentskills.io/specification.md
- **Last reviewed:** 2026-05-27
- **Used for:**
  - `references/security.md` → all sections
  - `SKILL.md` → Core Principle (Treat skills as software), Gotchas
- **Aspects extracted:**
  - `allowed-tools` frontmatter (experimental, space-separated allowlist) → least-privilege section; the "reduces blast radius, not injection likelihood" framing also informed by the security-considerations section of https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview
  - `compatibility` field for declaring network/runtime needs → network/runtime section
  - Treat-skills-as-software, audit untrusted scripts/URLs, fetched-content-as-untrusted-data, no hardcoded secrets → repo-original synthesis (not a transcription of an upstream "Security" page)
