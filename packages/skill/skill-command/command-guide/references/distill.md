# distill: Distill a Fat Command into a Skill Delegator

Refactor a self-contained ("fat") command — one that inlines its whole Goal / Workflow / Output / Gotchas — into a thin delegator that keeps only its argument contract and loads a guideline skill at run time. The procedure moves into the skill's `references/{operation}.md`, leaving the skill as the single source of truth and the command as a stable public interface.

## Goal

- Move a command's procedure, output format, and gotchas into a guideline-skill reference
- Reduce the command to its frontmatter, argument contract, and a delegation block (~15-25 lines, down from 70-160)
- Wire the command plugin to depend on the skill plugin so the skill is present at run time
- Preserve behavior: the distilled command, when invoked, loads the skill and produces the same output

## Arguments

- `command-file` (required) — path to the command file to distill
- `--skill <plugin>` (optional) — the guideline-skill plugin that should own the procedure (auto-detect from the command's domain if omitted)
- `--operation <name>` (optional) — the operation/reference name within the skill (default: the command's verb)
- `--dry-run` (optional) — preview the thin command, the skill reference, and manifest changes without writing

## What Stays vs What Moves

### Stays in the command (its public contract)

- Frontmatter — `description`, `allowed-tools` (with `Skill` added), `argument-hint`
- Arguments — the flag/default contract (command-unique; absent from the skill)
- A `## Delegation` block naming the exact skill + operation

### Moves to the skill reference (the single source of truth)

- Goal, Core Workflow, Output Format, Examples, Error Handling, Gotchas — deleted from the command, authored once in `references/{operation}.md`

## Core Workflow

1. **Read the command** — identify its verb, argument contract, and the procedure body to relocate
2. **Find the owning skill** — match the command's domain to an existing `*-guide` skill; if the operation already has an owner, reuse it (never duplicate a concept — cross-reference)
3. **Author or update the skill reference** — write the procedure into `references/{operation}.md` in the skill's reference style (Goal / Workflow / Output / Error Handling / Gotchas)
4. **Register the operation** — add it to the skill's `SKILL.md` under Operations and Progressive Disclosure with an explicit load-when trigger
5. **Slim the command** — replace the body with the thin shape: frontmatter + Arguments + a Delegation block
6. **Add the `Skill` tool** — ensure `allowed-tools` includes `Skill`
7. **Wire the dependency** — add the skill plugin to the command plugin's `dependencies` in BOTH `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` (bare-string name; once per plugin, not per command)
8. **Preview or apply** — show the three artifacts (`--dry-run`) or write them
9. **Report** — command line-count reduction, the skill/operation it now targets, manifests touched

## The Thin-Command Shape

```
---
description: <one line>
allowed-tools: [..., Skill]
argument-hint: "<contract>"
---

# /<plugin>:<command> — <Title>

## Arguments

- <the flag/default contract>

## Delegation

Load the `<skill>` skill (plugin `<plugin>`) and perform its **<operation>**
operation with these arguments. The skill is the source of truth for the
procedure, output format, and gotchas — do not restate them.
```

## Install ≠ Load

The `dependencies` entry guarantees the skill is **installed**, not that its text is in context at run time. The command must **explicitly** load the skill via the `Skill` tool — reliable precisely because the dependency guarantees presence. Never rely on implicit ambient auto-trigger of another plugin's skill; it is not a designed contract.

## Error Handling

- Command already thin (has a Delegation block, no inlined procedure) → report and skip
- No owning skill exists → author the `*-guide` skill first (skill-guide owns guideline-skill authoring), or flag it for the user
- Operation already owned by another skill → reuse it; do not duplicate the concept
- Skill reference link doesn't resolve after registration → fix the `SKILL.md` → `references/` path

## Safety

Commit to git first; use `--dry-run`; confirm the skill reference resolves and the manifests are valid JSON before relying on the result; verify the command still allows the `Skill` tool.

## Gotchas

- Distilling the body but forgetting the `dependencies` entry leaves the command unable to guarantee its skill at run time — wire both manifests
- Adding `Skill` to `allowed-tools` is easy to miss; without it the delegation can't load the skill
- Copying the procedure into the skill instead of moving it leaves two diverging copies — the command body must be deleted, not duplicated
- A delegation block that names the skill but not the exact operation makes the command ambiguous — name both
- Leaving the argument contract out of the thin command strands callers — the flags/defaults stay command-side
