---
type: plan
has_subplans: false
parent_plan: plans/agentic/platform-workflow-improvement.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - .moon/tasks/tag-skill.yml
    - .moon/tasks/tag-command.yml
    - .github/workflows/**
    - .gitlab/ci.yml
    - .claude-plugin/marketplace.json
    - packages/script/script-moon-plugin-validate/**
    - packages/script/script-moon-skill-audit-sources/**
    - packages/command/command-workflow/.claude-plugin/plugin.json
    - packages/command/command-workflow/.codex-plugin/plugin.json
    - packages/command/command-workflow/moon.yml
skills_to_consult:
  - moon-guide
  - typescript-guide
  - shell-scripting-guide
  - vitest-guide
  - skill-guide
  - versioning-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# 01 — Quality Gates: Close the Real Validation Gaps

## Objective

Give the plugin catalog the gates it actually lacks: marketplace/version
sync and plugin-dependency registration checks on every PR (for BOTH the
skill and command tags), an explicit exemption convention for
`skill-audit-sources`, a weekly scheduled audit report, and a report-only
eval hook on version-packages PRs. Fix the known `plugin.json` dependency
defects first so the new checks pass at HEAD.

## Context (read this first — no other context is assumed)

File:line references are anchors as of `main` @ `2b276a7f` (2026-07-03)
and will drift — always read the cited file before editing; if a line
reference doesn't match, locate the named construct instead.

Verified current state:

1. `.moon/tasks/tag-skill.yml:33-35` — `ci-check` aggregates
   `[build, format-check, skill-validate]`, so skill packages ALREADY get
   per-PR structural validation (`moon-skill-validate` covers
   frontmatter, body limits, reference links, progressive-disclosure
   triggers per `packages/script/script-moon-skill-validate/README.md`).
   Do not rebuild any of that.
2. `.moon/tasks/tag-command.yml` — `ci-check` aggregates only
   `[build, format-check]`. Command packages ship with zero validation.
3. Dead tasks (`runInCI: false`): `skill-audit-sources`
   (tag-skill.yml:58), `skill-eval-triggers` (:69), `skill-eval-outputs`
   (:80). Never run anywhere; keep them out of per-PR CI (parent
   decision 3) but give them scheduled/release homes.
4. Nothing checks that `.claude-plugin/marketplace.json` versions match
   package versions, or that every plugin-declared dependency is a
   registered marketplace plugin.
5. Known violation the new check must not trip over:
   `packages/command/command-workflow/.claude-plugin/plugin.json`
   declares `xonovex-skill-bdd`, `xonovex-skill-tdd`,
   `xonovex-skill-code-quality` (unused by any command) and omits
   `xonovex-skill-github` / `xonovex-skill-gitlab` (loaded host-detected
   by pr-create / pr-review-post / pr-review-resolve).
6. `.gitlab/ci.yml` mirrors GitHub; moon-task gates propagate to both
   automatically. Cron + eval hooks are GitHub-only by parent decision 1;
   the GitLab mirror documents that exception.
7. `moon-skill-eval-triggers` requires the `claude` CLI on PATH and
   supports `--max-budget-usd` (see its README) — the eval hook needs an
   `ANTHROPIC_API_KEY` repo secret and a budget default.

## Tasks

1. **Fix plugin.json dependency lists** (both manifests byte-identical):
   in `packages/command/command-workflow/.claude-plugin/plugin.json` and
   `.codex-plugin/plugin.json`, remove `xonovex-skill-bdd`,
   `xonovex-skill-tdd`, `xonovex-skill-code-quality`; add
   `xonovex-skill-github` and `xonovex-skill-gitlab`. Sync
   `packages/command/command-workflow/moon.yml` `dependsOn` to match.
   Verify command-utility's manifests are already accurate (audit said
   6/6 used) before touching them.
2. **New package `packages/script/script-moon-plugin-validate`** —
   scaffold mirroring `script-moon-skill-validate` (package.json with
   bin, moon.yml with `script`/`npm` tags, README). Checks, each with a
   clear failure message:
   a. every workspace plugin package version ===
      `.claude-plugin/marketplace.json` entry version === marketplace
      top-level version (lockstep);
   b. every `dependencies` entry in any `plugin.json` names a plugin
      registered in marketplace.json;
   c. `.claude-plugin/plugin.json` and `.codex-plugin/plugin.json` are
      byte-identical per package;
   d. (advisory, warn-only) dependencies declared by a command plugin
      that no command file in it ever references — surfaced in output,
      never a failure.
   Unit tests with fixture manifests (violation per check).
3. **Wire the check into both tags**: add a `plugin-validate` task to
   `.moon/tasks/tag-skill.yml` AND `.moon/tasks/tag-command.yml`
   (same shape as `skill-validate`: npx bin + install dep), and append it
   to each `ci-check` deps list.
4. **Exemption convention in `script-moon-skill-audit-sources`**: support
   a SOURCES.md whose frontmatter declares
   `exempt: "<reason>"` (house-style/principles skills). Audit output
   must distinguish three states — tracked, exempt-with-reason, MISSING —
   and exit non-zero only on schema errors, not on missing (missing is
   report content until 06 backfills). Document the marker in the README.
5. **Weekly audit workflow** `.github/workflows/skill-audit.yml`: cron
   (weekly), runs `npx moon run '#skill:skill-audit-sources'` across the
   catalog, uploads the report artifact and creates-or-updates a pinned
   issue with the summary. Add a comment in `.gitlab/ci.yml` recording
   the GitHub-only exception (parent decision 1).
6. **Eval report-only hook**: extend the release path so version-packages
   PRs get an eval report comment — a GitHub workflow triggered on PRs
   whose branch matches the version-packages convention, running
   `moon-skill-eval-triggers` for skills that have `eval-queries.json`,
   with `--max-budget-usd` defaulted (start at 5), using the
   `ANTHROPIC_API_KEY` secret; job is `continue-on-error` (report-only).

## Validation Steps

- `npx moon run script-moon-plugin-validate:test` (fixtures) green.
- `npx moon ci :ci-check` green at HEAD after task 1 (proves check + fix
  ordering).
- Introduce a deliberate marketplace version mismatch locally → ci-check
  fails with the task-2a message; revert.
- `act`-style dry run (`npm run ci:github`) parses the new workflows.
- `npx moon run :lint :typecheck :build :test` green.

## Success Criteria

- [ ] plugin.json dependency lists accurate; `.claude-plugin` ==
      `.codex-plugin` per package.
- [ ] `plugin-validate` runs in ci-check for skill AND command tags;
      HEAD passes; seeded violations fail with actionable messages.
- [ ] audit-sources distinguishes tracked / exempt / missing; exemption
      marker documented.
- [ ] Weekly skill-audit workflow exists, GitHub-hosted; GitLab mirror
      documents the exception.
- [ ] Eval hook posts a report-only comment on a version-packages PR,
      budget-capped, secret-gated.

## Files Modified/Created

- Modified: `.moon/tasks/tag-skill.yml`, `.moon/tasks/tag-command.yml`,
  `packages/command/command-workflow/{.claude-plugin,.codex-plugin}/plugin.json`,
  `packages/command/command-workflow/moon.yml`,
  `packages/script/script-moon-skill-audit-sources/*`, `.gitlab/ci.yml`
- Created: `packages/script/script-moon-plugin-validate/**`,
  `.github/workflows/skill-audit.yml`, eval-hook workflow (or release.yml
  extension)

## Dependencies

None (group 1). Subplan 06 depends on this subplan's exemption marker.

## Estimated Duration

3–4 days.
