---
type: plan
has_subplans: false
parent_plan: ../symmetric-workflow-commands.md
parallel_group: 4
status: complete
updated: 2026-07-20
completed_date: "2026-07-20"
dependencies:
  plans:
    - subplan-01-command-contract-and-inventory.md
    - subplan-02-plan-skill-decoupling.md
    - subplan-03-role-reference-and-invocation-guides.md
    - subplan-04-external-trigger-boundary.md
  files:
    - packages/command/command-workflow/commands/create.md
    - packages/command/command-workflow/commands/review.md
    - packages/command/command-workflow/docs/role-lenses.md
    - packages/command/command-workflow/docs/references.md
    - packages/command/command-workflow/docs/invocation.md
    - packages/skill/skill-plan/plan-guide/SKILL.md
    - packages/diagram/diagram-agent-workflow/operation-model.dot
    - packages/agent/agent-operator-go/cmd/operator/main.go
skills_to_consult:
  - testing-guide
  - typescript-guide
  - vitest-guide
  - command-guide
  - skill-guide
  - moon-guide
  - git-guide
validation:
  type_check: passed
  lint: passed
  build: passed
  tests: passed
  integration: passed
---

# Subplan 05: Validation and breaking release

## Objective

Make the symmetric model mechanically checkable, prove that removed workflow
and governance surfaces have no active residue, and align package/catalog
metadata for the coordinated next major release. This child closes the refactor
after all behavior and documentation children are complete.

The release is intentionally breaking. Document removals in the established
release notes/changelog path, but do not create aliases, wrappers, deprecated
files, or `MIGRATION.md`.

## Tasks

### 1. Turn command validation into an inventory contract

- Extend
  `packages/command/command-workflow/scripts/validate-documentation.mjs:1-93`
  with pure helpers that read the command directory and validate an explicit
  expected set.
- Require exactly the eight core operations and four workspace utilities, with
  no missing or extra Markdown files.
- Validate required command frontmatter/description structure, unique command
  names, non-empty prompt bodies, and internal documentation links.
- Derive displayed counts from the inventory or reject hard-coded stale counts.
- Update the dependency check for an empty universal dependency set while still
  requiring Claude/Codex manifest parity.

Use an explicit invariant rather than a prefix heuristic:

```js
const expectedCommands = new Set([
  "abandon",
  "create",
  "decide",
  "execute",
  "publish",
  "review",
  "revise",
  "validate",
  "workspace-abandon",
  "workspace-cleanup",
  "workspace-create",
  "workspace-merge",
]);
```

### 2. Add semantic residue checks with focused fixtures

- Add validator checks for `--profile`, former command names, lifecycle/gate
  claims, central provider/reference resolution, workflow-runtime identifiers,
  role-specific command APIs, trigger/executor modes, and enforced A1/A2/A3
  semantics.
- Scope checks to active command/skill/agent/docs/diagram/config surfaces so
  historical plans and legitimate sandbox-security policy do not create false
  positives.
- Prefer small allowlists for unavoidable release-note mentions over broad path
  exclusions.
- Add fixture-driven tests for:
  - missing, extra, and duplicate commands;
  - malformed prompt/frontmatter;
  - stale role/lifecycle terminology;
  - acceptable provider-native reference examples;
  - acceptable execution-security policy wording.

### 3. Align cross-package validation

- Update
  `packages/script/script-moon-skill-validate/src/cross-package-links.ts:304-390`
  so comments and scans describe the current README/docs/commands surface and
  no longer mention a migration file or delegating workflow commands.
- Update `cross-package-links.test.ts:57-186` fixtures to exercise the 12-command
  layout and the retained planning references.
- Preserve responsibility boundaries: command-local link/inventory validation
  stays in `validate-documentation.mjs`; cross-package validation checks only
  links that cross package boundaries.
- Update `command-workflow/moon.yml:15-39` task inputs only where the new docs or
  tests require it.

### 4. Run a tracked-file residue audit

- Search active tracked files under `packages/command`, `packages/skill`,
  `packages/agent`, `packages/shared`, and `packages/diagram` for:
  - removed packages and runtime identifiers;
  - all 53 former command filenames/invocations;
  - `--profile` and maturity-as-policy behavior;
  - lifecycle approval/authority/governed-tail claims;
  - `AgentTrigger`, `AgentSchedule`, trigger receiver/service, and cron residue.
- Classify every policy match. Retain sandbox, pod, network, filesystem,
  process, and tool security policy; remove only delivery-lifecycle governance.
- Ignore generated `.moon/cache` content because it is not a tracked source
  surface.
- Confirm `packages/shared/shared-agent-go` has no workflow-runtime dependency
  and no functional diff from this refactor.

### 5. Synchronize package, plugin, catalog, and release metadata

- Update command-workflow and skill-plan descriptions in their `package.json`,
  `.claude-plugin/plugin.json`, and `.codex-plugin/plugin.json` manifests.
- Update the matching entries in `.claude-plugin/marketplace.json` and
  `.agents/plugins/marketplace.json` to describe composable operations and
  planning/code-research boundaries.
- Apply the repository's lockstep plugin/skill version policy through the
  normal version-packages workflow, targeting the next major because commands
  and Kubernetes APIs are removed without compatibility support.
- Update generated lockfiles and established release notes/changelogs as
  required by that workflow.
- Mention removal of old commands and `AgentTrigger`/`AgentSchedule` in release
  notes, but do not create a migration document or step-by-step compatibility
  path.
- Do not push or publish. Release remains PR-only and occurs only after the
  version-packages PR is reviewed and merged to `main`.

### 6. Execute the complete validation matrix

- Run all affected package CI checks and fix warnings at their source.
- Run TypeScript/Vitest checks for cross-package validation.
- Run operator Go formatting, build, lint, unit, integration, manifest, and
  generated-code checks.
- Render and inspect both updated diagrams.
- Re-run residue searches after formatting/generation because generated outputs
  can restore deleted names.
- Inspect the final diff for accidental historical-plan rewrites or unrelated
  changes.

## Validation steps

1. `npx moon run command-workflow:ci-check`
2. `npx moon run skill-plan:ci-check`
3. `npx moon run diagram-agent-workflow:ci-check`
4. `npx moon run agent-operator-go:ci-check`
5. `npx moon run shared-agent-go:ci-check`
6. Run the `script-moon-skill-validate` build, typecheck, lint, and test targets
   exposed by Moon, including cross-package link tests.
7. Run the repository version-package validation/dry-run target without
   publishing.
8. Run final tracked-file residue searches and `git diff --check`.
9. Confirm package, harness manifests, marketplace catalogs, and lockfiles agree
   on descriptions, dependencies, and versions.

## Success criteria

- [x] Validation derives and enforces the exact 12-command inventory.
- [x] Tests fail for asymmetric inventory, malformed prompts, stale lifecycle
      concepts, and broken cross-package links.
- [x] Residue checks distinguish retained execution security from removed
      workflow governance.
- [x] No active source/config/documentation/eval surface references removed
      commands, runtimes, profiles, gates, triggers, or schedules.
- [x] `shared-agent-go` remains functionally unchanged and passes its checks.
- [x] All affected Moon, TypeScript, Vitest, Go, Kubernetes, diagram, formatting,
      and release-dry-run checks pass without warnings.
- [x] Plugin/package/catalog metadata is synchronized for the next lockstep
      major release.
- [x] No branch is created, nothing is pushed/published, and no migration or
      compatibility artifact is added.

## Files modified/created

- Modify: `packages/command/command-workflow/scripts/validate-documentation.mjs`.
- Create or modify focused validator fixtures/tests under
  `packages/command/command-workflow/scripts/`.
- Modify as needed: `packages/command/command-workflow/moon.yml`.
- Modify: `packages/script/script-moon-skill-validate/src/cross-package-links.ts`.
- Modify: `packages/script/script-moon-skill-validate/src/cross-package-links.test.ts`.
- Modify: affected package and plugin manifests for command-workflow and
  skill-plan.
- Modify: `.claude-plugin/marketplace.json`.
- Modify: `.agents/plugins/marketplace.json`.
- Modify: root/package lockfiles and established release metadata generated by
  the lockstep version process.
- Verify only: active files in command, skill, agent, shared, and diagram
  packages not otherwise owned above.

## Dependencies

- Requires subplans 01-04 to be complete and validated at their package
  boundaries.
- Runs alone because it intentionally touches manifests, Moon tasks, validators,
  and metadata that overlap the preceding outputs.
- Completion makes the parent plan ready for implementation validation, not
  automatic publishing.

## Validation Results

- All five affected package `ci-check` targets pass: command workflow, plan
  skill, workflow diagram, agent operator, and shared agent policy.
- The skill validator passes build, typecheck, lint, formatting, coverage, and
  all 24 tests; cross-package validation resolves 4/4 links, 248/248 handoffs,
  and 90 dependency pairs.
- The release validator passes 987 checks across 92 lockstep command and skill
  packages at version `7.0.0`, including package, Claude, Codex, marketplace,
  dependency, description, and lockfile parity.
- Command validation passes nine focused fixture tests and 11 validation groups
  over the exact 12-command inventory. Final semantic-residue and tracked-file
  checks find no active removed surface, and `shared-agent-go` has no diff.
- Operator unit, coverage, formatting, lint, build, envtest integration,
  manifest, and generated-code checks pass. Both updated diagrams render and
  were visually inspected.
- `npm audit --audit-level=moderate` reports zero vulnerabilities, and
  `git diff --check` passes. npm 11.16's `dedupe` command emits an upstream
  advisory for `esbuild@0.28.1` because that command path does not load the
  root `allowScripts` policy; `npm install` and `npm approve-scripts` confirm
  the pinned approval is effective with no unreviewed install scripts.
- No branch, push, publish, `MIGRATION.md`, alias, wrapper, or compatibility
  artifact was created.

## Estimated duration

One to two focused implementation sessions. Inventory tests are compact, while
the cross-package residue audit and lockstep release dry run span the monorepo.
