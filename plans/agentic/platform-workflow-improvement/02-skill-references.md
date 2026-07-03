---
type: plan
has_subplans: false
parent_plan: plans/agentic/platform-workflow-improvement.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/skill-code-review/code-review-guide/**
    - packages/skill/skill-pull-request/pull-request-guide/**
skills_to_consult:
  - skill-guide
  - command-guide
  - code-review-guide
  - pull-request-guide
  - git-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 02 — Skill References: Document the Undocumented pr Operations

## Objective

Write the skill-side documentation that the five fat pr-* commands
currently inline: `review-post` and `review-resolve` operation references
(including the `commentId` write-back contract) in code-review-guide, and
a `create` operation reference in pull-request-guide. After this subplan,
every operation a command delegates to exists in its owning skill — the
precondition for subplan 03's distillation.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03); read cited files before
editing.

1. `packages/skill/skill-code-review/code-review-guide/references/`
   contains `conventional-comments.md`, `findings-schema.md`,
   `review-analyze.md`, `review-refine.md`, `review-structure.md` — but
   NO `review-post.md` or `review-resolve.md`, although
   `pr-review-post.md` and `pr-review-resolve.md` commands delegate
   "review craft" to this skill and inline the procedure themselves.
2. The `commentId` contract is implicit: pr-review-post populates
   `commentId`s on posted findings; pr-review-resolve matches findings to
   threads by them. `findings-schema.md` does not document the field.
3. `pr-create.md:34-40` inlines a three-skill orchestration
   (pull-request-guide description craft, git-guide push/rebase/title,
   host-skill create op) plus "Command-level glue": branch/base routing,
   host detection with stop-if-missing, preview + `--yes`/`--dry-run`
   gating. pull-request-guide has no `create` operation reference to
   receive this.
4. The source prose to relocate lives in the five command files under
   `packages/command/command-workflow/commands/pr-*.md` — treat those
   Delegation sections as the raw material; do NOT edit the commands here
   (that is subplan 03).
5. House rules: skill-guide governs reference structure and progressive
   disclosure; comments/state present behavior only.

## Tasks

1. **`code-review-guide/references/review-post.md`** — the posting
   operation: transform verified findings into a host review (summary +
   line-anchored comments per review-structure.md), additive body edits,
   host delegation (github-guide / gitlab-guide), and the commentId
   capture contract (every posted finding records the created comment's
   id for the resolve stage).
2. **`code-review-guide/references/review-resolve.md`** — the resolve
   operation: match findings to threads via commentId, verdict taxonomy
   (fixed / no-change-needed / disputed / stale), re-check flow against
   the new diff, thread-closing etiquette per host.
3. **Document `commentId`** in
   `code-review-guide/references/findings-schema.md` as an optional
   field populated at post time and required by resolve.
4. **`pull-request-guide/references/create.md`** — the create operation:
   sequence (git-guide handoff for branch/commit/title/push-rebase →
   host detection from the git remote, stop naming the missing host
   skill if none installed → host-skill create op for PR/MR, draft,
   reviewers, work-item and sibling links), flag routing table
   (branch/base/title/description/draft/reviewers/work-item/related),
   and the preview gate (`--yes` skips, `--dry-run` stops after
   preview).
5. **Update both SKILL.md progressive-disclosure sections** to list the
   new references with load-when lines. While in
   code-review-guide/SKILL.md: double-quote its description (one of the
   three unquoted ones — done here instead of subplan 05 to avoid a
   parallel-group file overlap).
6. **Validate**: `npx moon run skill-code-review:skill-validate
   skill-pull-request:skill-validate` (reference-link integrity is
   enforced by the validator).

## Validation Steps

- `npx moon run '#skill:skill-validate'` green for the two packages.
- `npx moon run :fmt:check :lint` green.
- Cross-check: every operation named by the five pr-* command Delegation
  sections now resolves to a reference file in its owning skill.

## Success Criteria

- [ ] review-post.md and review-resolve.md exist, cover the inlined
      command prose, and are listed in SKILL.md progressive disclosure.
- [ ] commentId documented in findings-schema.md (post populates,
      resolve consumes).
- [ ] pull-request-guide create.md covers the pr-create orchestration
      including host detection and preview gating.
- [ ] skill-validate green; no command files modified.

## Files Modified/Created

- Created: `packages/skill/skill-code-review/code-review-guide/references/review-post.md`,
  `.../review-resolve.md`,
  `packages/skill/skill-pull-request/pull-request-guide/references/create.md`
- Modified: both SKILL.md files,
  `code-review-guide/references/findings-schema.md`

## Dependencies

None (group 1). Subplan 03 depends on this subplan.

## Estimated Duration

1–2 days.
