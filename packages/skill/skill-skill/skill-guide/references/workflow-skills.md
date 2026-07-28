# workflow-skills: Creating Workflow / Operation Skills

A workflow skill encodes a **procedure** the agent — or a command plugin — delegates to (e.g. plan, git, skill, instruction, command, content, reflect), not coding-style rules for a language or framework (those are [guideline-skills.md](guideline-skills.md)). A skill may **combine both** — a procedure skill that also lays down coding-style rules keeps an `## Essentials` block alongside its `## Operations` (e.g. llmstxt teaches the format _and_ its authoring workflow).

Split independently selectable domain guidance instead of embedding every possible
method in the workflow skill. The workflow owns operation boundaries; selected domain
skills supply subject-specific procedures without widening those boundaries.

## Skeleton

`Core Principles → Operations → Gotchas → Progressive Disclosure`

- **Core Principles** — the rules that govern every operation (replaces a guideline skill's `Essentials`).
- **Operations** — one row per procedure the skill performs (replaces the single `Example`). Each operation maps **1:1** to a `references/` file and, usually, to a command (`/review` → the `Review` operation → `references/review.md`); name the file after the operation, not after the skill. The Operations list mirrors the Progressive Disclosure list.
- **Gotchas** and **Progressive Disclosure** — as in any skill.

Split Operations into domain groups once there are more than ~7 — e.g. `## Plan Lifecycle` + `## Plan Operations`, or `## Commit Operations` + `## Branch Operations` + `## Worktree Operations`.

For the structural patterns _inside_ an operation — output templates, checklists, validation loops, plan-validate-execute — see [instruction-patterns.md](instruction-patterns.md).

## Template

Scaffold from `assets/workflow-skill-template/`:

- [`SKILL.md.template`](../assets/workflow-skill-template/SKILL.md.template) — Core Principles, Operations, Gotchas, Progressive Disclosure
- [`references/{operation}.md`](../assets/workflow-skill-template/references/{operation}.md) — operation template (Goal + Core Workflow, then the operation-specific sections it needs; common optional closers: Output / Error Handling / Gotchas)
- [`evals.json`](../assets/workflow-skill-template/evals.json) — three output-quality seeds with observable assertions
- [`eval-queries.json`](../assets/workflow-skill-template/eval-queries.json) — 16 trigger-eval queries (8 positive + 8 negative with train/validation splits)
- [`SOURCES.md`](../assets/workflow-skill-template/SOURCES.md) — required catalog provenance; cite external sources or declare repository-original provenance

To scaffold: copy the directory, rename `{operation}.md` files, and fill in `{placeholders}`.

## Command delegation

When a command plugin delegates to an operation, keep the contract one-directional: the command body says "load the `<skill>` skill and perform its **<operation>** operation", and the skill — not the command — owns the procedure, output format, and gotchas. Declare the skill in every dependency-capable manifest the command package actually supports. A Claude-only command package has one Claude manifest; Codex distributes and invokes the skill directly instead of mirroring a command package.

That command-to-skill edge is an exact hard dependency. Inside the workflow skill,
derive supporting capability needs from the operation request and choose among
installed skill names and routing descriptions. A required missing capability blocks;
a preferred missing capability remains visible.

The workflow operation owns its effect (`inspect`, `preview`, or `apply`), persistence
boundary, subject identity, and revision requirements. For cross-role work, accept a
Markdown request contract that preserves subject relationships, operation inputs,
required and preferred capabilities, evidence, effects, and constraints. A selected
domain procedure adapts to that contract: it cannot mutate during inspect or preview,
persist an inline result, or silently perform another operation. Provider-backed
mutation requires the provider's conditional revision and idempotency mechanisms when
they exist; otherwise report that the guarantee is unavailable.

## Conventions

- **Operation bullet:** `- **<Operation>** — <what it produces> — see [references/<operation>.md](references/<operation>.md)`
- One operation per reference file; one level deep under `references/`; kebab-case filename matching the operation.
- The `Load when…` trigger for each operation lives in Progressive Disclosure, not in the reference file.
