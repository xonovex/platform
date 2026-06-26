---
type: plan
has_subplans: false
parent_plan: plans/command-skill-distillation.md
parallel_group: 3
status: pending
feature: command-skill-distillation
dependencies:
  plans:
    - plans/command-skill-distillation/03-git-host-skill-create.md
    - plans/command-skill-distillation/02-workflow-plan-git-distill.md
  files:
    - packages/command/command-workflow/commands/pr-create.md
    - packages/command/command-workflow/commands/pr-review-analyze.md
    - packages/command/command-workflow/commands/pr-review-refine.md
    - packages/command/command-workflow/commands/pr-review-post.md
    - packages/command/command-workflow/commands/pr-review-resolve.md
    - packages/command/command-workflow/.claude-plugin/plugin.json
    - packages/command/command-workflow/.codex-plugin/plugin.json
skills_to_consult: [command-guide, skill-guide, git-guide, pull-request-guide, code-review-guide]
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 04 — pr-command-distill

## Objective

De-duplicate the restated craft inside the five `pr-*` commands and repoint their
orchestration at the operations that subplan `03` created — `git-host-guide`
(`references/{create,review-post,review-resolve}.md` + the Host Mapping tables) and
`code-review-guide` (`references/{findings-schema,review-analyze,review-refine}.md`) —
plus the existing `pull-request-guide` and `git-guide`. Finish the `xonovex-workflow`
`dependencies` array in both manifests by merging the three craft/host skills onto the
`plan` + `git` entries that subplan `02` added.

These commands are tier-2: they stay **thicker** than the thin tier-1 delegators. Each one
keeps its frontmatter (with `Skill` in `allowed-tools`), its `## Arguments` flag contract,
its workflow ordering, its error handling, and its per-stage least-privilege `allowed-tools`
(e.g. `pr-review-refine` is `Read/Write/Edit/Grep/AskUserQuestion/Skill` with no `Bash`;
`pr-review-post` adds `Bash` for the host calls). What gets **deleted** is duplicated craft
that now has exactly one owner: the inlined Description Template, the two divergent copies of
the Findings Schema, the host-driving prose / Host Mapping tables, and the restated label
vocabulary and blocking rules. Each command names the exact skill(s) + operation(s) it loads
in a `## Delegation` block. Behavior when invoked must be unchanged; the single live drift
(the Findings Schema in `pr-review-analyze` carries `status: new|recurring` while the copy in
`pr-review-post` dropped it) collapses to one owner.

## Tasks

1. **Distill `pr-create.md`** — `packages/command/command-workflow/commands/pr-create.md`.
   - Delete the `## Description Template` block (lines ~74-102, the `## What / ## Why / ## Changes / ## Testing / ## Tradeoffs / ## Related` markdown) — `pull-request-guide` owns it.
   - Delete the inlined `## Host Mapping` table (lines ~46-59); replace with a one-line pointer that the host create/link/draft mapping is owned by `git-host-guide`'s `create` operation. (Scope: keep a host-mapping *reference*, not the table.)
   - Trim `## Goal`, `## Implementation Details`, and `## Gotchas` down to orchestration-only points (host detection ordering, preview/confirm, coordinates-from-remote each run); drop the bullets that restate description craft now owned by `pull-request-guide`.
   - Keep: frontmatter, `## Arguments` (verbatim flag contract), `## Core Workflow` ordering, `## Error Handling`, preview/`--dry-run`/`--yes` confirm gate, `## Examples`.
   - Rewrite the intro skill bullets (lines ~18-21) into a `## Delegation` block naming the exact owners:

   ```markdown
   ## Delegation

   Load and follow these skills via the `Skill` tool; they are the source of truth — do not restate them:
   - `pull-request-guide` (plugin `xonovex-skill-pull-request`) — the **description craft** (what/why/how, sizing depth to change, the lean template, the self-review gate).
   - `git-host-guide` (plugin `xonovex-skill-git-host`) — the **create** operation: host detection, coordinates-from-remote, auth, branch push, PR create, draft, reviewers, work-item / sibling links, and the Host Mapping.
   - `git-guide` (plugin `xonovex-skill-git`) — branch, commit, the conventional-commit title, and the rebase-onto-base that keeps the diff clean.
   ```

2. **Distill `pr-review-analyze.md`** — `packages/command/command-workflow/commands/pr-review-analyze.md`.
   - Delete the `## Findings Schema` block (lines ~50-71) — `code-review-guide`'s `findings-schema` operation is now the single owner of the JSON contract (including `status` and `commentId`).
   - Trim `## Implementation Details` / `## Gotchas` to drop restated review craft (label vocabulary, decorations, verify-before-asserting) now owned by `code-review-guide`'s `review-analyze`; keep the analyze-specific mechanics that the command orchestrates (new-file line tracking from hunk headers, the `--since` diff-against-prior tagging).
   - Keep: frontmatter, `## Arguments` including `--since`, `--effort`, `--out`, `--diff`; `## Core Workflow` ordering; `## Error Handling`; `## Examples`.
   - Rewrite the intro craft paragraph (line ~18) into a `## Delegation` block: load `code-review-guide` (plugin `xonovex-skill-code-review`) and perform its **`review-analyze`** operation; the findings shape comes from its **`findings-schema`** operation. Note to also load relevant domain skills (design-system, accessibility, language) for grounded findings.

3. **Distill `pr-review-refine.md`** — `packages/command/command-workflow/commands/pr-review-refine.md`.
   - This command has no inlined schema block to delete, but `## Per-Finding Operations` (keep/reword/relabel/re-anchor/merge/split/drop), `## Feedback Sources`, and the label-discipline bullets restate `code-review-guide` review craft — trim the craft restatement, keep the per-finding arg flow and the STOP-after-each-pass control loop that this command uniquely orchestrates.
   - Keep: frontmatter (note the narrower `allowed-tools`: `Read/Write/Edit/Grep/AskUserQuestion/Skill`, no `Bash`); `## Arguments` (`findings-file`, `--walk`, `--final`); `## Prerequisites`; `## Core Workflow` ordering; `## Error Handling`; `## Examples`.
   - Rewrite the intro craft paragraph (line ~21) into a `## Delegation` block: load `code-review-guide` (plugin `xonovex-skill-code-review`) and perform its **`review-refine`** operation, with the contract from its **`findings-schema`** operation.

4. **Distill `pr-review-post.md`** — `packages/command/command-workflow/commands/pr-review-post.md`.
   - Delete the `## Findings Schema` block (lines ~64-83) — owned by `code-review-guide`'s `findings-schema`. This removes the second, divergent copy (the one missing `status`), resolving the live drift.
   - Delete the inlined `## Host Mapping` table (lines ~39-48); replace with a one-line pointer that the anchored-comment / blocking-task / deep-link mapping is owned by `git-host-guide`'s `review-post` operation.
   - Trim `## Goal` / `## Implementation Details` / `## Gotchas` to drop restated label/decoration/cross-link craft (owned by `code-review-guide`) and host-REST prose (owned by `git-host-guide`); keep the posting orchestration ordering (validate → preview → post inline → promote blockers → post+cross-link summary → write back ids → verify).
   - Keep: frontmatter, `## Arguments` (`--pr`, `--findings`, `--no-tasks`, `--yes`, `--dry-run`), `## Core Workflow` ordering, `## Error Handling`, preview/confirm gate, `## Examples`.
   - Rewrite the intro skill bullets (lines ~17-18) into a `## Delegation` block: load `code-review-guide` (plugin `xonovex-skill-code-review`) for the **review craft** + **`findings-schema`** contract, and `git-host-guide` (plugin `xonovex-skill-git-host`) for the **`review-post`** operation (host posting, blocking mechanism, deep-links, Host Mapping).

5. **Distill `pr-review-resolve.md`** — `packages/command/command-workflow/commands/pr-review-resolve.md`.
   - Delete the inlined `## Host Mapping` table (lines ~40-47); replace with a one-line pointer that the read-threads / resolve-thread / reply mapping is owned by `git-host-guide`'s `review-resolve` operation.
   - Trim `## Implementation Details` / `## Gotchas` to drop restated review craft; keep the resolve-specific orchestration (match by `commentId` first, fall back to anchor; verdict-needs-evidence; never blanket-resolve).
   - Keep: frontmatter, `## Arguments` (`--pr`, `--findings`, `--reply`, `--yes`, `--dry-run`), `## Core Workflow` ordering, `## Error Handling`, verify/confirm gate, `## Examples`.
   - Rewrite the intro skill bullets (lines ~18-19) into a `## Delegation` block: load `code-review-guide` (plugin `xonovex-skill-code-review`) for the **verify-addressed judgment** (is the finding genuinely fixed, not just moved), and `git-host-guide` (plugin `xonovex-skill-git-host`) for the **`review-resolve`** operation (read threads, resolve, reply, Host Mapping).

6. **Merge the craft/host deps into both manifests** —
   `packages/command/command-workflow/.claude-plugin/plugin.json` and
   `packages/command/command-workflow/.codex-plugin/plugin.json`.
   - Add `xonovex-skill-pull-request`, `xonovex-skill-code-review`, `xonovex-skill-git-host` to the existing `dependencies` array. Subplan `02` already added `xonovex-skill-plan` and `xonovex-skill-git` — **merge, do not clobber**. Both files currently have no `dependencies` key (it lands via `02`); if `02` has not yet written it when this lands, create the array with all five entries. Final array (bare strings, same marketplace, version optional):

   ```json
   "dependencies": [
     "xonovex-skill-plan",
     "xonovex-skill-git",
     "xonovex-skill-pull-request",
     "xonovex-skill-code-review",
     "xonovex-skill-git-host"
   ]
   ```

   - Apply the identical array to both `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` so Claude Code and Codex stay in parity. `xonovex-skill-git-host` is the new plugin registered by subplan `03`.

7. **Validate** — run the steps in `## Validation Steps`: `fmt-check` and `build` on `command-workflow`, parse both manifests, then run the `analyze → post → resolve` pipeline once against a throwaway PR and confirm the findings schema has exactly one owner, the named skills load via the `Skill` tool, and the output is unchanged from the pre-distillation behavior.

## Validation Steps

- **type_check / tests**: no app code (markdown + JSON only). The "type" check is JSON validity — `node -e "require('./packages/command/command-workflow/.claude-plugin/plugin.json'); require('./packages/command/command-workflow/.codex-plugin/plugin.json')"` parses both manifests. No unit tests exist for command markdown.
- **lint**: `npx moon run command-workflow:fmt-check` — prettier `fmt:check` over the touched package (`**/*.md` + manifests).
- **build**: `npx moon run command-workflow:build` — moon build for the only touched project.
- **integration**:
  - Confirm each of the five commands still has `Skill` in `allowed-tools` and a `## Delegation` block naming the exact skill plugin(s) + operation(s).
  - Grep the five command bodies to confirm zero remaining inlined `Findings Schema`, `Description Template`, or `Host Mapping` tables — exactly one owner each (schema in `code-review-guide/findings-schema`, template in `pull-request-guide`, Host Mapping in `git-host-guide`).
  - Invoke `pr-review-analyze` → `pr-review-post` (`--dry-run`) → `pr-review-resolve` (`--dry-run`) once against a throwaway branch/PR and confirm: each command loads its named skill at run time, the findings schema is consistent across stages (the `status` drift is gone), and the rendered preview/output matches the pre-distillation behavior.
  - Confirm `git-host-guide`'s Host Mapping drives the create/post/resolve host calls identically to the deleted inlined tables.

## Success Criteria

- [ ] All five `pr-*` commands keep their frontmatter (`Skill` in `allowed-tools`), `## Arguments` flag contract, workflow ordering, error handling, and per-stage least-privilege `allowed-tools` — they are not flattened to thin delegators.
- [ ] `pr-create.md` no longer inlines the Description Template or the Host Mapping table; description craft delegates to `pull-request-guide`, host create to `git-host-guide/create`, branch/commit/rebase to `git-guide`.
- [ ] `pr-review-analyze.md` no longer inlines the Findings Schema; methodology delegates to `code-review-guide/review-analyze`; `--since` and `--effort` flow preserved.
- [ ] `pr-review-refine.md` delegates to `code-review-guide/review-refine` + `findings-schema`; the per-finding arg flow and STOP-each-pass loop are preserved.
- [ ] `pr-review-post.md` no longer inlines the Findings Schema or the Host Mapping; craft delegates to `code-review-guide`, host posting to `git-host-guide/review-post`.
- [ ] `pr-review-resolve.md` delegates verify-addressed judgment to `code-review-guide` and host resolve to `git-host-guide/review-resolve`.
- [ ] The Findings Schema has exactly one owner (`code-review-guide/findings-schema`); the `status: new|recurring` drift between analyze and post is gone.
- [ ] Both `command-workflow` manifests declare `dependencies` containing all five entries (`plan`, `git`, `pull-request`, `code-review`, `git-host`), merged not clobbered, identical across `.claude-plugin` and `.codex-plugin`.
- [ ] `fmt-check` and `build` are green on `command-workflow`; both manifests parse as valid JSON.
- [ ] The analyze→post→resolve pipeline runs once with each command loading its named skill and producing unchanged output.

## Files Modified / Created

- `packages/command/command-workflow/commands/pr-create.md` (modified)
- `packages/command/command-workflow/commands/pr-review-analyze.md` (modified)
- `packages/command/command-workflow/commands/pr-review-refine.md` (modified)
- `packages/command/command-workflow/commands/pr-review-post.md` (modified)
- `packages/command/command-workflow/commands/pr-review-resolve.md` (modified)
- `packages/command/command-workflow/.claude-plugin/plugin.json` (modified — merge deps)
- `packages/command/command-workflow/.codex-plugin/plugin.json` (modified — merge deps)

## Dependencies

- **`03-git-host-skill-create.md` must land first.** This subplan repoints all host orchestration and the de-duped findings contract at operations that `03` authors: `git-host-guide` (`references/{create,review-post,review-resolve}.md` + Host Mapping tables) and `code-review-guide`'s new `references/{findings-schema,review-analyze,review-refine}.md`. The `xonovex-skill-git-host` plugin name added in the manifests is registered by `03`. Delegating to operations that do not exist yet would strand the commands.
- **`02-workflow-plan-git-distill.md` must land first (shared manifest).** Subplans `02` and `04` both edit `command-workflow/.claude-plugin/plugin.json` and `.codex-plugin/plugin.json`. `02` adds `xonovex-skill-plan` + `xonovex-skill-git` to `dependencies`; this subplan merges the three craft/host skills onto that array. Landing after `02` avoids a clobber/merge conflict on the same `dependencies` key (per the parent plan's command-workflow contention risk).

## Estimated Duration

~1 day — five careful rewrites that must preserve the stateful pipeline invariants (anchors, `commentId` write-back, the analyze→refine→post→resolve handoff) while stripping craft, plus the two-manifest merge and the live pipeline integration check.
