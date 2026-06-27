# workflow-skills: Creating Workflow / Operation Skills

A workflow skill encodes a **procedure** the agent — or a command plugin — delegates to (e.g. plan, git, skill, instruction, command, content, reflect), not coding-style rules for a language or framework (those are [guideline-skills.md](guideline-skills.md)). A skill may **combine both** — a procedure skill that also lays down coding-style rules keeps an `## Essentials` block alongside its `## Operations` (e.g. llmstxt teaches the format _and_ its authoring workflow).

## Skeleton

`Core Principles → Operations → Gotchas → Progressive Disclosure`

- **Core Principles** — the rules that govern every operation (replaces a guideline skill's `Essentials`).
- **Operations** — one row per procedure the skill performs (replaces the single `Example`). Each operation maps **1:1** to a `references/` file and, usually, to a command (`/plan-create` → the `plan-create` operation → `references/plan-create.md`). The Operations list mirrors the Progressive Disclosure list.
- **Gotchas** and **Progressive Disclosure** — as in any skill.

Split Operations into domain groups once there are more than ~7 — e.g. `## Plan Lifecycle` + `## Plan Operations`, or `## Commit Operations` + `## Branch Operations` + `## Worktree Operations`.

For the structural patterns _inside_ an operation — output templates, checklists, validation loops, plan-validate-execute — see [instruction-patterns.md](instruction-patterns.md).

## Template

Scaffold from `assets/workflow-skill-template/`:

- [`SKILL.md.template`](../assets/workflow-skill-template/SKILL.md.template) — Core Principles, Operations, Gotchas, Progressive Disclosure
- [`references/{operation}.md`](../assets/workflow-skill-template/references/{operation}.md) — operation template (Goal + Core Workflow, then the operation-specific sections it needs; common optional closers: Output / Error Handling / Gotchas)
- [`eval-queries.json`](../assets/workflow-skill-template/eval-queries.json) — trigger-eval queries (8 train + 4 validation, mix of should-trigger and near-miss)
- [`SOURCES.md`](../assets/workflow-skill-template/SOURCES.md) — optional; include only if the skill distills an external source (house-process skills cite none)

To scaffold: copy the directory, rename `{operation}.md` files, and fill in `{placeholders}`.

## Command delegation

When a command plugin delegates to an operation, keep the contract one-directional: the command body says "load the `<skill>` skill and perform its **<operation>** operation", and the skill — not the command — owns the procedure, output format, and gotchas. Declare the skill in the command plugin's `dependencies` (both manifests).

## Conventions

- **Operation bullet:** `- **<Operation>** — <what it produces> — see [references/<operation>.md](references/<operation>.md)`
- One operation per reference file; one level deep under `references/`; kebab-case filename matching the operation.
- The `Load when…` trigger for each operation lives in Progressive Disclosure, not in the reference file.
