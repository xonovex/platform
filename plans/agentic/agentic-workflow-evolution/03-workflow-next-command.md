---
type: plan
has_subplans: false
parent_plan: plans/agentic/agentic-workflow-evolution.md
parallel_group: 3
status: pending
dependencies:
  plans:
    - plans/agentic/agentic-workflow-evolution/02-workflow-definitions.md
  files:
    - packages/command/command-workflow/commands/workflow-next.md
    - packages/command/command-workflow/.claude-plugin/plugin.json
    - packages/command/command-workflow/.codex-plugin/plugin.json
    - packages/command/command-workflow/moon.yml
    - packages/skill/skill-workflow/workflow-guide/**
skills_to_consult:
  - command-guide
  - skill-guide
  - plan-guide
  - code-review-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 03 — workflow-next: The L1 Interpreter

## Objective

The assisted level: a thin `workflow-next` command backed by a
workflow-guide operation that reads the definition (from the trusted
ref), the plan document, and the run journal, computes the next step,
and either suggests it (default) or executes at most ONE step before
handing off to a fresh session (`--auto`, parent critique resolution).
Assisted gates get their reviewer wired in.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); executes after the
improvement plan and subplan 02 — re-read cited files.

1. Interpreter semantics are NORMATIVE in workflow-guide
   (`definition-format.md`, `gates.md` from 02). This subplan implements
   them as a prompt operation — any rule discovered ambiguous here gets
   fixed in workflow-guide FIRST (02's files), then implemented; never
   fork the semantics.
2. One-step rule (parent decision + critique): `--auto` executes a
   single step in the current session, records the result, prints the
   hand-off (what to run in a fresh session). Session-per-step is the
   invariant that keeps L1 behavior identical to L2.
3. Gate handling at L1: `manual` → present the gate, stop; `assisted` →
   run the reviewer operation (e.g. plan-critique for merge-review),
   record its recommendation as a gate verdict in the journal (01's
   schema), present for the human's call, stop; `auto` → record verdict
   and proceed (subject to policy read from the trusted ref — decision
   10: read `workflows/*.yml` from main, not the working branch).
4. Plugin dependency: command-workflow must declare
   `xonovex-skill-workflow` — the improvement plan's plugin-validate
   enforces marketplace registration (02 registered it).
5. Discovery: `plan-list` (improvement plan subplan 04) provides plan
   enumeration; workflow-next resolves the active plan from
   `git config branch.<branch>.plan` first, then most-recent, same as
   other plan commands.

## Tasks

1. **workflow-guide reference `workflow-next.md`** — the operation:
   inputs (workflow name or default `feature`, plan file auto-detect);
   load definition from trusted ref (`git show main:workflows/<name>.yml`
   — state the exact mechanism); load plan frontmatter + journal;
   compute next per definition-format.md rules (including foreach over
   pending subplans and loop-cap exhaustion → `blocked` result);
   suggest vs `--auto` one-step semantics; gate handling per Context 3;
   result emission per 01's contract. Add to SKILL.md progressive
   disclosure.
2. **Thin command**
   `packages/command/command-workflow/commands/workflow-next.md` —
   standard Delegation wording to workflow-guide **workflow-next**;
   `argument-hint: [workflow] [--plan-file <path>] [--auto]`;
   allowed-tools: Read, Glob, Grep, Bash, Skill.
3. **Declare the dependency** — add `xonovex-skill-workflow` to
   command-workflow's `.claude-plugin/plugin.json` +
   `.codex-plugin/plugin.json` (byte-identical) + moon.yml `dependsOn`.
4. **Reviewer wiring** — in workflow-guide `gates.md`, the concrete
   assisted-gate procedure: invoke the reviewer operation via the Skill
   tool, map its findings to a gate verdict (`approve | reject |
   needs-human`), journal it; workflow-next follows it.
5. **Fixture walkthrough** — a scratch plan + journal exercising: mid
   foreach (names the right subplan's continue), loop-cap hit (emits
   blocked), manual gate (stops), assisted gate (reviewer verdict
   journaled), `--auto` (one step + hand-off message).

## Validation Steps

- `#skill:skill-validate` (workflow-guide) + `#command:format` +
  plugin-validate green; `npx moon ci :ci-check` green.
- Live smoke on this repo: with the improvement plan mid-flight,
  `/xonovex-workflow:workflow-next` names the correct next step and
  honors the gate policy (parent success criterion 3).
- `--auto` smoke: executes exactly one step, writes one journal entry,
  prints the fresh-session hand-off.

## Success Criteria

- [ ] workflow-next suggests correctly against a real plan; policy read
      from main, not the branch.
- [ ] `--auto` = one step + hand-off; no multi-step in-session chains.
- [ ] Assisted gate runs its reviewer and journals the verdict.
- [ ] Dependency declared; all catalog checks green.

## Files Modified/Created

- Created: `workflow-guide/references/workflow-next.md`,
  `command-workflow/commands/workflow-next.md`
- Modified: workflow-guide SKILL.md + gates.md, command-workflow
  plugin manifests + moon.yml

## Dependencies

Requires 02. Parallel with 04 (disjoint files: commands/skill vs Go).
07 waits for this.

## Estimated Duration

2–3 days.
