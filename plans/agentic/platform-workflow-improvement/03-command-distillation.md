---
type: plan
has_subplans: false
parent_plan: plans/agentic/platform-workflow-improvement.md
parallel_group: 2
status: pending
dependencies:
  plans:
    - plans/agentic/platform-workflow-improvement/02-skill-references.md
  files:
    - packages/command/command-workflow/commands/pr-create.md
    - packages/command/command-workflow/commands/pr-review-analyze.md
    - packages/command/command-workflow/commands/pr-review-post.md
    - packages/command/command-workflow/commands/pr-review-refine.md
    - packages/command/command-workflow/commands/pr-review-resolve.md
    - packages/command/command-utility/commands/slashcommand-distill.md
    - packages/command/command-utility/commands/version-bump.md
skills_to_consult:
  - command-guide
  - skill-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 03 — Command Distillation: Thin the Fat pr-* Commands

## Objective

Restore all 40 commands to the thin-delegation pattern: the five pr-*
commands drop their inlined orchestration (now documented skill-side by
subplan 02) and shrink to ~25–30 lines with the standard Delegation
wording; the five remaining over-long descriptions come down to ~1
line, and the pr-review pipeline gets stage-tagged descriptions.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); read cited files before
editing.

1. Fat commands (lines): `pr-review-analyze.md` (51),
   `pr-review-refine.md` (51), `pr-create.md` (49), `pr-review-post.md`
   (47), `pr-review-resolve.md` (45). Each inlines "Command-level
   orchestration"/"glue" prose in its Delegation section (e.g.
   `pr-create.md:34-40`). The 23 thin commands share this exact wording —
   use it verbatim:

   > Load the `<skill>` skill (plugin `<plugin>`) and perform its
   > **<operation>** operation with these arguments. The skill is the
   > source of truth for the procedure, output format, and gotchas — do
   > not restate them.

2. Subplan 02 created the receiving operations: pull-request-guide
   `create`, code-review-guide `review-post` / `review-resolve` (analyze
   and refine references already existed). Distill AGAINST those — if a
   sentence in a command's Delegation block is not in the skill
   reference, move it there first (that is an 02 fix, coordinate before
   deleting).
3. Over-long descriptions (chars): slashcommand-distill 198
   (command-utility), pr-create 187, pr-review-resolve 166,
   pr-review-post 164, version-bump 163 (command-utility). Target
   <= ~100 chars, per the command-guide distill rule (~1 line).
   NOTE (2026-07-03, parent decision 7): the plan-* lifecycle commands
   were already renamed/merged (plan-decide, plan-revise), trimmed, and
   stage-tagged ("Pre-plan:"/"Draft:"/"Execution:" prefixes + a
   lifecycle strip under each H1) ahead of this subplan — only pr-* and
   command-utility descriptions remain here. Follow the same stage-tag
   convention for the pr-review pipeline (e.g. "Review 1/4: ...").
4. `allowed-tools`: pr-create.md already lists `Skill` (verified); the
   audit flagged pr-review-analyze as missing it — verify each pr-*
   command and add `Skill` where absent.
5. Do NOT touch plugin.json dependency lists (done in 01) and do NOT
   rename arguments (that is subplan 04, which edits some of the same
   files — this ordering is why 04 is group 3).

## Tasks

1. **Distill `pr-create.md`**: keep frontmatter (shortened description),
   Arguments, Examples; replace the three-skill Delegation prose with the
   standard wording targeting pull-request-guide **create**; delete the
   "Command-level glue" paragraph (now in create.md).
2. **Distill the four `pr-review-*.md` commands** the same way:
   analyze → code-review-guide **review-analyze**; refine →
   **review-refine**; post → **review-post**; resolve →
   **review-resolve**. Verify/add `Skill` in allowed-tools on each.
3. **Trim the 5 descriptions** listed in Context 3 to <= ~100 chars,
   preserving the imperative one-line summary style of the thin
   commands; prefix the four pr-review-* descriptions with their
   pipeline stage ("Review 1/4: analyze", "2/4: refine", "3/4: post",
   "4/4: resolve"); keep `argument-hint` untouched.
4. **Consistency pass** across all 40 commands: confirm none besides the
   five carries orchestration prose in Delegation (grep for
   "Command-level"); confirm delegation wording matches the standard
   block.
5. **Format + validate**: `npx moon run '#command:format'`, then the
   01-installed `plugin-validate` and full ci-check.

## Validation Steps

- Each pr-* command <= ~30 lines (`wc -l`), standard Delegation block,
  `Skill` in allowed-tools.
- All 5 remaining descriptions <= ~100 chars
  (`awk`/frontmatter check); pr-review-* carry pipeline-stage tags.
- `npx moon ci :ci-check` green (includes plugin-validate from 01).
- Manual smoke: run `/xonovex-workflow:pr-create --dry-run` in a test
  session — behavior unchanged (same flags, same preview gate, procedure
  served from the skill).

## Success Criteria

- [ ] 5 pr-* commands thin, delegating to the 02 operations; no
      orchestration prose remains command-side.
- [ ] 5 remaining descriptions <= ~100 chars, pr-review pipeline
      stage-tagged.
- [ ] `Skill` present in allowed-tools wherever a command loads skills.
- [ ] ci-check green; `--dry-run` smoke test behaves as before.

## Files Modified/Created

- Modified: the 7 command files listed in frontmatter `dependencies.files`.
- Created: none.

## Dependencies

Requires 02 (skill references must exist before command prose is
deleted). Subplan 04 edits overlapping files and must wait for this.

## Estimated Duration

1–2 days.
