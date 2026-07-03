---
type: plan
has_subplans: true
status: approved
updated: 2026-07-03
feature: platform-workflow-improvement
dependencies:
  plans: []
  subplans:
    01-quality-gates: []
    02-skill-references: []
    03-command-distillation: [02-skill-references]
    04-workflow-plumbing: [03-command-distillation]
    05-catalog-dedup: []
    06-sources-eval-backfill: [01-quality-gates]
proposed_subplans:
  - 01-quality-gates
  - 02-skill-references
  - 03-command-distillation
  - 04-workflow-plumbing
  - 05-catalog-dedup
  - 06-sources-eval-backfill
parallel_groups:
  - group: 1
    plans: [01-quality-gates, 02-skill-references, 05-catalog-dedup]
    note: "Gates land first so they protect every subsequent edit (decision 1).
      Skill references must exist before commands are slimmed. The dedup diff
      touches skills disjoint from 02. All three touch disjoint files."
  - group: 2
    plans: [03-command-distillation, 06-sources-eval-backfill]
    depends_on: [1]
    note: "03 moves command prose into the references written by 02.
      06 needs 01's exemption-marker mechanism to know what to backfill."
  - group: 3
    plans: [04-workflow-plumbing]
    depends_on: [2]
    note: "Edits the same command files as 03; sequenced to avoid conflicts."
skills_to_consult:
  - skill-guide
  - command-guide
  - plan-guide
  - moon-guide
  - versioning-guide
  - typescript-guide
  - shell-scripting-guide
  - vitest-guide
research_sources:
  documentation:
    - https://agentskills.io/specification
    - https://code.claude.com/docs/en/skills.md
    - https://code.claude.com/docs/en/plugins-reference.md
    - https://code.claude.com/docs/en/plugin-marketplaces.md
    - https://platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices
  versions:
    agent-skills-spec: "open standard since Dec 2025; description max 1024
      chars; optional license/compatibility/metadata/allowed-tools fields"
    claude-code-commands: "merged into skills Q4 2025; legacy commands/
      format fully supported - staying legacy per decision 2"
    catalog: "74 skill packages + 2 command plugins, lockstep 4.1.0,
      published via .claude-plugin/marketplace.json"
---

# Platform workflow improvement

## Overview

Improve `packages/command` (2 plugins, 40 commands) and `packages/skill`
(74 skill packages), and both senses of "the workflow": the shipped plan/PR
lifecycle (command-workflow plugin + plan/code-review skills) and the
dev/release pipeline that validates and publishes these packages. Research
(2026-07-03) found a structurally healthy catalog whose real gaps are
quality gates that exist but never run, five fat pr-* commands drifting
from the thin-delegation pattern, undocumented pr-review operations, and
missing workflow plumbing.

## Execution Context

Cross-repo ordering, the ecosystem glossary (skill/command package
anatomy, moon commands, lockstep releases, plan-status conventions),
and the per-subplan execution protocol live in `plans/agentic/roadmap.md` —
read it before any subplan. When a subplan cites "decision N", the
Decision Log below is normative. Each subplan is self-contained given
its own Context section, this decision log, and its `skills_to_consult`
— assume no other session context exists.

## Decision Log (2026-07-03, via plan-research + plan-clarify)

1. **Scope**: both product workflow and dev/release workflow, as
   independently mergeable subplans; gates land first.
2. **Command format**: stay on legacy `commands/`; distill fat commands by
   moving orchestration into owning-skill references. Unified-format
   migration deferred to a future plan gated on codex-plugin support.
3. **Quality gates**: tiered - deterministic checks per-PR in a real
   `ci-check`; `skill-audit-sources` as a weekly scheduled report; evals
   report-only on the version-packages PR, promoted to blocking once
   stable.
4. **Backfill**: `SOURCES.md` for the ~20 source-derived skills; explicit
   exemption marker for house-style/principles skills; `eval-queries.json`
   priority-first (process + easily-confused skills).
5. **Dedup**: evidence-gated - diff the 5 `error-handling.md` and 3
   `validation.md` files first; create a general-owner skill only if the
   shared core is substantive, default to cross-links. The 4 `testing.md`
   files link up to `testing-guide` unconditionally.
6. **Plumbing**: full - `plan-worktree-create` auto-sets
   `branch.<branch>.plan`, `pr-create` gains `--plan`, positional arg
   standardized to `[plan-file]`, new `plan-list` command backed by a new
   read-only plan-guide operation.
7. **Lifecycle naming** (2026-07-03, applied immediately, ahead of
   subplan execution): `plan-clarify` + `plan-interrogate` merged into
   `plan-decide` (walks known open decisions, or discovers unknown ones
   by questioning); `plan-refine` renamed `plan-revise`; all plan-*
   commands carry stage-tagged descriptions ("Pre-plan:" / "Draft:" /
   "Execution:") and a lifecycle strip under the title. Breaking by
   semver, but the owner chose to ship it as the 4.1.0 lockstep bump
   (applied 2026-07-03, all 76 plugin packages + manifests +
   marketplace). Subplans 03/04 re-scoped accordingly.

## Goals

- Every PR touching skill or command packages passes deterministic gates:
  structure validation (already live for skills, extended to commands),
  marketplace/version sync, and plugin-dependency registration.
- Audit and eval tooling runs on schedule / at release instead of never;
  "green" means checked, not skipped (explicit exemptions).
- All 40 commands follow the thin-delegation pattern; every delegated
  operation is documented in its owning skill.
- The plan lifecycle is navigable without conversation context: plans are
  discoverable, associated with branches and PRs, and share one argument
  convention.
- Catalog integrity: one owner per concept (or explicit cross-links),
  provenance tracked where a real upstream exists.

## Current State

- moonrepo 2.3.5, npm workspaces; skills/commands are markdown plugins in
  lockstep 4.1.0, published via `.claude-plugin/marketplace.json`.
- Per-PR validation for skill packages already exists: tag-skill.yml's
  `ci-check` aggregates `[build, format-check, skill-validate]`, and
  `skill-validate` covers frontmatter, body limits, reference links, and
  progressive-disclosure triggers. The genuinely dead pieces are
  `skill-audit-sources`, `skill-eval-triggers`, and `skill-eval-outputs`
  (`runInCI: false`) - never run anywhere.
- tag-command.yml's `ci-check` aggregates only `[build, format-check]` -
  command plugins ship with no validation at all.
- Nothing anywhere checks marketplace version sync or plugin-dependency
  registration.
- 5 fat pr-* commands (45-51 lines) inline orchestration; 23 thin commands
  average ~25 lines. 5 command descriptions still exceed ~100 chars
  (the plan-* ones were fixed with decision 7).
- `code-review-guide/references/` lacks `review-post.md` /
  `review-resolve.md`; the `commentId` contract between post and resolve is
  implicit.
- `command-workflow/.claude-plugin/plugin.json` declares 3 unused
  dependencies (bdd, tdd, code-quality) and omits 2 used ones
  (github, gitlab - host-detected).
- 29/74 skills lack `SOURCES.md`; 32/74 lack `eval-queries.json`
  (including skill-plan). 3 descriptions unquoted; 3 near the 1024-char
  spec cap.
- `plan-continue` reads `git config branch.<branch>.plan` but nothing sets
  it; no `--plan` on pr-create; no plan discovery command; four names for
  the same positional argument across plan commands.

## Research Findings

- Upstream spec is stable; no forced migrations. Commands/skills
  unification (Q4 2025) is backward compatible - legacy format has no
  announced removal. Recommended description length for discovery is
  80-200 chars (catalog median 433).
- All command->skill delegations resolve (23 thin commands verified); the
  problem is placement of procedure, not broken references.
- Catalog tiering (general -> language -> framework) has no downward or
  circular references; 202 cross-skill mentions, zero dangling.
- The dedup findings are filename-level; content overlap is unverified -
  hence decision 5's evidence gate.
- Verified during critique: `skill-validate` already gates skill PRs via
  ci-check deps and already covers reference links; command packages have
  no equivalent. The gate work is additive, not remedial.

## Proposed Approach

1. **01-quality-gates** - close the actual gaps: new deterministic checks
   (marketplace version sync, plugin-dependency registration) as a new
   `script-moon-plugin-validate` package inherited by BOTH the skill and
   command tags (command packages currently ship with zero validation);
   fix both plugin.json dependency lists up front (remove
   bdd/tdd/code-quality; declare github/gitlab as host-detected) so the
   new check passes at HEAD; exemption-marker convention in
   `script-moon-skill-audit-sources`; weekly scheduled audit-sources
   report hosted on GitHub Actions (matching release.yml; the GitLab
   mirror documents it as a GitHub-only job); eval report-only hook on
   the version-packages PR with a repo API-key secret and a default
   `--max-budget-usd` cap.
2. **02-skill-references** - write `review-post.md` and
   `review-resolve.md` (including the `commentId` write-back contract) in
   code-review-guide; write the `create` operation reference in
   pull-request-guide (git-guide handoff, host detection, flag routing,
   preview gate); update both SKILL.md progressive-disclosure sections.
3. **03-command-distillation** - distill the 5 pr-* commands to thin
   delegation (~25-30 lines); trim the 5 remaining long descriptions
   (stage-tagging the pr-review pipeline, decision 7 convention); keep
   `.codex-plugin` parity byte-identical (plugin.json dependency lists
   are already corrected in 01).
4. **04-workflow-plumbing** - `plan-worktree-create` sets
   `branch.<branch>.plan`; `pr-create --plan`; standardize `[plan-file]`
   across plan commands; new `plan-list` command + plan-guide operation
   reference (glob `plans/*.md`, report status/phase per plan and
   subplan).
5. **05-catalog-dedup** - diff the 5 error-handling and 3 validation
   references; record the verdict; create general owner(s) only if the
   shared core is substantive, else cross-link; add testing-guide
   cross-links to the 4 testing.md files unconditionally; quote the
   unquoted descriptions (code-review's lands with 02, which already
   edits that file); trim the 3 near-cap descriptions.
6. **06-sources-eval-backfill** - author `SOURCES.md` for the ~20
   source-derived skills; mark house-style/principles skills exempt via
   01's marker; sources that cannot be confidently reconstructed get a
   low-confidence marker or exempt-with-reason instead of invented
   provenance (timeboxed per skill); add `eval-queries.json` to the
   priority set (plan, git, github/gitlab, testing/tdd/bdd, command,
   instruction, reflect).

## Risk Assessment

- **Eval flakiness as a gate**: mitigated by decision 3 - report-only
  until stable; promotion to blocking is a later, separate change.
- **Dual-CI drift**: moon-task gates propagate to GitHub and GitLab
  automatically, but the cron and eval hooks are GitHub-only by decision;
  the GitLab mirror records the exception so it cannot silently diverge.
- **Scheduled reports nobody reads**: the weekly audit report needs a
  visible landing place (issue or PR comment); wiring includes the
  notification, not just the cron.
- **Command edits during parallel work**: 03 and 04 touch the same files -
  sequenced groups prevent conflicts.
- **Dedup could still be wrong at content level**: the evidence gate makes
  the worst case one wasted diff task, not two thin packages.
- **Lockstep version churn**: every subplan merge is releasable, but the
  version bump + marketplace sync happens once at the end via the release
  PR (per packages/command AGENTS.md: bump command plugins, skill plugins,
  and marketplace.json together).
- **Deferred**: unified skill-format migration (needs codex-side support
  verification); eval-queries for the remaining ~25 niche skills.

## Success Criteria

- [ ] A PR introducing marketplace version drift or an unregistered
      plugin dependency fails `ci-check` for skill AND command packages
      (reference links remain covered by skill-validate); current HEAD
      passes, with the plugin.json fixes landing inside 01.
- [ ] Weekly audit-sources workflow produces a report; exemption marker
      documented in skill-guide and honored by the script (missing !=
      exempt).
- [ ] Eval report appears on a version-packages PR (report-only).
- [ ] All 5 pr-* commands are thin (~30 lines, standard Delegation
      wording); their procedure lives in skill references; all 5
      remaining long descriptions <= ~100 chars.
- [ ] code-review-guide documents review-post and review-resolve
      including the commentId contract; pull-request-guide documents
      create.
- [ ] plugin.json dependencies accurate for both plugins;
      `.claude-plugin`/`.codex-plugin` remain byte-identical.
- [ ] All plan commands use `[plan-file]`; `plan-list` works;
      `pr-create --plan` and `plan-worktree-create` set the branch
      config `plan-continue` reads.
- [ ] Dedup verdict recorded with the diff evidence; testing.md
      cross-links added; SOURCES.md present or exempt for all 74 skills;
      priority eval-queries added.
- [ ] `npx moon run :lint :typecheck :test :build` green; release via
      version-packages PR only.

## Estimated Effort

~1.5-2 weeks elapsed: 01 ~3-4 days (new script + CI wiring), 02 ~1-2 days,
03 ~1-2 days, 04 ~2 days, 05 ~2 days (diff + verdict + hygiene),
06 ~2-3 days (20 SOURCES.md files is research-heavy).
