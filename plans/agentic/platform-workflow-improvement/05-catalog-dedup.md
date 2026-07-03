---
type: plan
has_subplans: false
parent_plan: plans/agentic/platform-workflow-improvement.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - packages/skill/skill-c99/c99-guide/references/error-handling.md
    - packages/skill/skill-expressjs/expressjs-guide/references/error-handling.md
    - packages/skill/skill-hono/hono-guide/references/error-handling.md
    - packages/skill/skill-lua/lua-guide/references/error-handling.md
    - packages/skill/skill-shell-scripting/shell-scripting-guide/references/error-handling.md
    - packages/skill/skill-expressjs/expressjs-guide/references/validation.md
    - packages/skill/skill-kubernetes/kubernetes-guide/references/validation.md
    - packages/skill/skill-shell-scripting/shell-scripting-guide/references/validation.md
    - packages/skill/skill-android-analytics/android-analytics-guide/**
    - packages/skill/skill-android-wcag/android-wcag-guide/**
    - packages/skill/skill-cmake/cmake-guide/references/testing.md
    - packages/skill/skill-github/github-guide/SKILL.md
    - packages/skill/skill-gitlab/gitlab-guide/SKILL.md
    - packages/skill/skill-user-stories/user-stories-guide/SKILL.md
skills_to_consult:
  - skill-guide
  - code-quality-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 05 — Catalog Dedup: Evidence First, Then Cross-Links or an Owner

## Objective

Settle the concept-ownership findings with evidence (parent decision 5):
diff the five `error-handling.md` and three `validation.md` references;
create a general-owner skill ONLY if the shared core is substantive,
otherwise cross-link. Unconditionally: link the four `testing.md` files
up to `testing-guide`, and fix the description hygiene items.

## Context (read this first — no other context is assumed)

Anchors as of `main` @ `2b276a7f` (2026-07-03).

1. The duplication finding is FILENAME-level, not content-verified. The
   catalog's own counter-example: `create.md` appears 5 times and is
   correct domain-parallelism. Do not scaffold new packages before the
   diff verdict.
2. Error-handling references (5): c99, expressjs, hono, lua,
   shell-scripting. Validation references (3): expressjs, kubernetes,
   shell-scripting. Testing references not linking `testing-guide` (4):
   android-analytics, android-wcag, cmake, expressjs.
3. Creating a new skill package costs: scaffold (package.json,
   moon.yml, .claude-plugin + .codex-plugin, SKILL.md),
   `.claude-plugin/marketplace.json` registration, lockstep version,
   and a description competing in trigger space with five language
   guides that all mention "error handling".
4. Cross-reference house style: bold skill names (`**testing-guide**`)
   in prose; upward-only (framework → language → general); verified 202
   existing cross-references, zero dangling.
5. Hygiene items (from audit): unquoted YAML descriptions in
   android-analytics and android-wcag SKILL.md (code-review's is quoted
   by subplan 02, which already edits that file); near-cap descriptions
   github (720 chars), gitlab (762), user-stories (883) against the
   1024 spec limit.

## Tasks

1. **Diff the five error-handling.md files**: identify prose/principles
   present in >= 3 of them (normalize for language-specific syntax).
   Record the verdict in this subplan's Results section when executing:
   substantive shared core (>~40% overlapping guidance) vs
   language-specific parallel files.
2. **Same for the three validation.md files.**
3. **Act on the verdict** —
   Branch A (substantive): create `packages/skill/skill-error-handling`
   (and/or `skill-validation`) per skill-guide scaffolding: general
   principles only, description tuned to NOT steal language-guide
   triggers ("Use when designing error propagation strategy…" with
   explicit skip-clause routing to language guides); register in
   marketplace.json; shrink the language files to language-specific
   deltas with an upward cross-link.
   Branch B (default, parallel files): add a short sibling-awareness
   line to each file and NO new package.
4. **Unconditional testing links**: in the four testing.md references,
   add the upward cross-link to **testing-guide** for shared taxonomy
   (keep framework-specific content in place).
5. **Hygiene**: double-quote the android-analytics and android-wcag
   descriptions; trim github/gitlab/user-stories descriptions toward
   ~500–600 chars preserving trigger + skip clauses (skill-guide
   description rules).
6. **Validate**: `#skill:skill-validate` for every touched package; if
   Branch A ran, `plugin-validate` (01) confirms marketplace
   registration and lockstep version.

## Validation Steps

- Diff artifacts (overlap summary per concept) recorded in Results.
- `npx moon run '#skill:skill-validate'` green for touched packages.
- If Branch A: new package passes skill-validate + plugin-validate;
  trigger sanity — the 3 near-cap and any new descriptions re-checked
  under 1024 chars.
- `npx moon ci :ci-check` green.

## Success Criteria

- [ ] Written overlap verdict for error-handling and validation, with
      the diff evidence.
- [ ] Branch A xor B executed per verdict; no thin platitude skill
      created without evidence.
- [ ] 4 testing.md files cross-link testing-guide.
- [ ] 3 descriptions quoted; 3 near-cap descriptions trimmed.
- [ ] skill-validate + ci-check green.

## Files Modified/Created

- Modified: the reference files listed in frontmatter; 6 SKILL.md
  frontmatter blocks (3 quoting, 3 trimming).
- Created (Branch A only): `packages/skill/skill-error-handling/**`
  (and/or `skill-validation/**`) + marketplace.json entry.

## Dependencies

None (group 1); files disjoint from 02. If Branch A creates packages,
the 01 checks validate them (01 lands in the same group — merge 01
first within the group if Branch A triggers).

## Estimated Duration

2 days (diff + links + hygiene); +1–2 days if Branch A triggers.
