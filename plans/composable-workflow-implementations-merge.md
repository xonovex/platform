---
type: plan
has_subplans: true
status: pending-approval
updated: '2026-07-16'
dependencies:
  plans:
  - composable-workflow-phases
proposed_subplans:
- merge-baseline-and-layout-normalization
- merge-walking-skeleton-and-evidence-honesty
- merge-adversarial-fixture-corpus-and-runner
- merge-plugin-plans-decoupling
- merge-traceability-repairs
- merge-closure-release-and-validation
parallel_groups:
- group: 1
  plans:
  - merge-baseline-and-layout-normalization
- group: 2
  plans:
  - merge-walking-skeleton-and-evidence-honesty
  - merge-adversarial-fixture-corpus-and-runner
  - merge-plugin-plans-decoupling
- group: 3
  plans:
  - merge-traceability-repairs
- group: 4
  plans:
  - merge-closure-release-and-validation
skills_to_consult:
- skill-guide
- command-guide
- git-guide
- versioning-guide
- shell-scripting-guide
- typescript-guide
- testing-guide
research_sources:
  retrieved: '2026-07-16'
  documentation:
  - /home/mvierssen/Projects/xonovex/xonovex-platform (merge base, branch main)
  - /home/mvierssen/Projects/xonovex/xonovex-platform-fable (graft source, pinned)
  - plans/composable-workflow-phases.md (both repos; shared parent plan)
  versions:
    xonovex-platform: 6.0.0 at e18d54f5
    xonovex-platform-fable: 5.0.0 at 82303137
    common-base: 166c4f26
---

# Composable Workflow Implementations Merge

## Overview

Consolidate the two independent implementations of the `composable-workflow-phases` plan into one canonical tree. `xonovex-platform` is the merge base (keeps its 6.0.0 release, Go webhook hardening, command quartets, migration docs, and CI-wired validators); `xonovex-platform-fable` at commit `82303137` is the pinned graft source for its executable walking skeleton, adversarial fixture corpus, and layout discipline. This is a file-level port, **not** a `git merge` — the two trees use divergent paths (nested vs. flat skill dirs, `code-harness-guide` vs. `cc-harness-guide`, different command names), so a branch merge would silently union parallel trees instead of surfacing conflicts. All implementation work happens in a dedicated git worktree (`../xonovex-platform-merge`) and lands on `main` via merge-back on completion.

## Goals

- Graft fable's executable walking skeleton (`guard.sh`, `run-skeleton.sh`, fixtures) into `skill-agent-governance` so the skeleton is proven by execution, not fixture replay.
- Port fable's adversarial fixture corpus (64 scenarios across workflow, governance, harness, and enterprise suites) and give it a machine runner in the platform's `.mjs` validator style, wired into moon test tasks.
- Normalize the 9 nested skill packages to the repo's flat `<topic>-guide/` layout that `packages/skill/AGENTS.md` mandates.
- Decouple the shipped `command-workflow` plugin from monorepo `plans/` paths.
- Repair the shared traceability defects (D-036–D-039 table shape, uncontrolled classification vocabulary) and add automated shape checks so they cannot recur.
- Make evidence claims honest: subplan-05 status must match what was actually executed; `VALIDATION.txt` must be consistent with `validation-policy.md` rule 5.
- Close the parent `composable-workflow-phases` plan bookkeeping once validation is green.
- Every ported or changed area carries a concrete verification check, and a port manifest proves nothing slated from fable was silently dropped.

## Non-goals

- Porting fable's command surface, prose references, or single-run-per-phase command names — platform's quartet surface is canonical.
- Re-mapping all 150 subplan traceability rows to per-task precision — only defect repair plus automated guards.
- Running live probes against enterprise platforms (Azure DevOps, Bitbucket, Bitrise, AWS, Datadog) — matrices stay candidate-until-probed.
- Reverting or re-cutting the published 6.0.0 release.

## Current State

Both repos diverged from common base `166c4f26` ("bump plugin catalog to 5.0.0"):

- **xonovex-platform** (`e18d54f5`, released 6.0.0): +26k/−1.7k across 680 files. Real machinery — 8,862-check `validate-documentation.mjs`, ~4,300 lines of fixture validators with mutation guards, fail-closed Go webhook hardening with tests, command deprecations with `MIGRATION.md`, reconciled diagrams, complete `evals.json`/`SOURCES.md`. Defects: walking skeleton is hand-authored fixture JSON (`"runtimeVersion": "fixture-runtime-1.0.0"`) yet subplan-05 claims `integration: passed`; 9/17 new skill packages use `skills/<topic>-guide/` nesting with `"skills": "./skills"` Codex manifests that `packages/skill/AGENTS.md` warns are unreliable; `command-workflow` docs, `moon.yml`, and validator hardcode `plans/composable-workflow-phases/` paths; `VALIDATION.txt` records PASS despite zero runtime probes, which validation-policy rule 5 forbids; parent plan still `status: approved`; a few untracked empty dirs.
- **xonovex-platform-fable** (`82303137`, additive at 5.0.0): +15.4k/−79 across 399 files. Executable walking skeleton verified by review (17/17 checks, advisory-first stop without `--yes`, shellcheck-clean, no side effects); uniform flat layout; candidate-until-probed matrices including one live Claude Code probe (2.1.211); 64 adversarial YAML fixtures with `must_not` prohibitions, zero orphans. Its fixtures and skeleton reference fable-only identifiers (`decision-record`, `acceptance-run`, `workflow-conformance-verify`, `cc-harness-guide`) that must be re-pointed during port.
- **Shared defects** (inherited from common planning artifacts): `decision-source-matrix.md` rows D-036–D-039 have 6 cells against a 5-column header with inconsistent separators; subplan traceability rows are blanket copy-pasted ID blocks.

## Research Findings

Three independent reviews (deep review per repo plus a head-to-head structural comparison) established:

- Fable's unique strengths are **self-contained**: the walking-skeleton assets are one directory; the fixture corpus is additive files indexed by suite; the layout/bookkeeping wins are mechanical edits. All graft cheaply onto platform.
- Platform's unique strengths are **entangled** with its content shape (validators written against its JSON schemas and `docs/` tree; deprecations cascading through 17 `plan-*` commands and `skill-plan` references), which is why platform is the base and not the other way around.
- The Go webhook hardening exists only in platform; fable's operator work was documentation-only and its pre-existing webhook still fails open. Nothing to port there — keep platform's.
- Fable proved a local Claude Code runtime probe is feasible (recorded version 2.1.211), which platform can reuse to satisfy validation-policy rule 5 for at least one harness.
- Fixture formats differ: fable ships YAML with agent-performed conformance; platform's validators are JSON-based Node scripts with no YAML dependency.

## Key Decisions

1. **Platform is the merge base; fable is pinned at `82303137`** for all ports so upstream drift cannot change what "correctly merged" means. Settled by prior analysis.
2. **File-level porting, never `git merge`** — divergent paths would union silently. Settled.
3. **Ported fixtures are converted YAML → JSON** to match the platform's validator infrastructure and avoid adding a YAML dependency to skill packages. Settled 2026-07-16: conversion is verified lossless (the fable corpus contains zero YAML comments; content is plain scalar/list/map data). The port manifest pairs each destination one-to-one with its YAML source at `82303137`; a 10-fixture spot-audit against the originals confirms fidelity. Alternative (keep YAML, add `yaml` dev-dependency) rejected: two fixture formats in one guide invites drift.
4. **Validation-policy rule 5 is satisfied the hybrid way** (settled 2026-07-16): run the local Claude Code probe (fable demonstrated feasibility at 2.1.211) and record it as dated evidence with the observed runtime version and an explicit rerun-on-version-change drift note; amend the policy/`VALIDATION.txt` so unprobed platforms carry explicit `candidate` status and the result line is qualified (e.g., `PASS (candidate matrices: 5 harnesses, 5 platforms unprobed)`) instead of an unqualified PASS.
5. **The 9-package layout move ships as a patch (6.0.1)**, lockstep across all 93 packages (settled 2026-07-16). Grounded in verified facts: both marketplace catalogs resolve only package roots, guide paths live solely in each plugin's own manifests, and the plugin cache snapshots whole packages per version — so the move plus manifest updates ship atomically and no consumer resolves the old paths. The release notes explicitly name the 9 relocated packages. Alternative (major 7.0.0 for the literal "published paths changed" reading) rejected: no evidenced path-scripting consumer exists, and a second breaking release for a defect repair erodes what "major" means.
6. **Both optional fable extras are skipped as files** (settled 2026-07-16): `two-plane-architecture.{dot,png}` (platform's rewritten `target-architecture.dot` already depicts both planes) and `platform-skill-convention.md` (platform's `enterprise-platforms.md` covers the same ground). Recorded skip-with-reason in the port manifest. The three rules fable states that platform does not are grafted into `enterprise-platforms.md`: cross-platform flows are owned by the initiating platform; OIDC-preferred/no-static-keys-by-default as a shared convention clause; `SOURCES.md` pinning with retrieval dates as a release precondition. Nuance: if the standalone two-plane diagram is later preferred, improve `target-architecture.dot` — never ship two overlapping diagrams.
7. **All implementation work executes in a dedicated git worktree** at `../xonovex-platform-merge` on branch `composable-workflow-implementations-merge`, created before the baseline gate and merged back to `main` on completion (then removed). The primary checkout stays untouched for the merge's scope.

## Proposed Approach

Each workstream ends with its own verification checks; the final gate re-runs everything.

1. **Worktree + baseline gate.** Create the dedicated worktree at `../xonovex-platform-merge` on branch `composable-workflow-implementations-merge`; all subsequent workstreams execute there. Record a green pre-merge baseline in the worktree at `e18d54f5`: full moon validation (format, lint, typecheck, build, test), `validate-documentation.mjs` (8,862 checks), `go test ./internal/webhook/`. Also snapshot the fable source list: every file under the port scope at `82303137`, forming the initial **port manifest** (source path → planned destination → disposition: port / adapt / skip-with-reason).
   *Verify:* `git worktree list` shows the new worktree on its branch; all baseline commands exit 0 in the worktree and results are recorded in the plan dir; the manifest enumerates 100% of fable's walking-skeleton and fixture files.
2. **Layout normalization.** `git mv` the 9 nested packages (skill-aws, -azure-devops, -bitbucket, -bitrise, -datadog, -accessibility, -ai-governance, -reliability, -security-assurance) from `skills/<topic>-guide/` to `<topic>-guide/`; point each `.codex-plugin/plugin.json` at the guide dir; delete the leftover empty dirs.
   *Verify:* `find packages/skill -path '*/skills/*' -name SKILL.md` returns nothing; no `"skills": "./skills"` remains in any plugin manifest; both marketplace catalogs still list 93 plugins whose paths all resolve; skill validators pass for all 9 moved packages.
3. **Walking-skeleton graft.** Port fable's `agent-governance-guide/assets/walking-skeleton/` into platform's `skill-agent-governance`, re-pointing fable-only identifiers to platform equivalents; rewrite `references/walking-skeleton.md` to distinguish the recorded fixture scenario (kept, still validator-checked) from the executable run (new); wire the skeleton run into the package's moon test task.
   *Verify:* `run-skeleton.sh --yes` exits 0 with 17/17 checks; without `--yes` it stops after preview (consent gate is executable); `shellcheck` clean; `git status` clean after a run; existing fixture validators and mutation guards still pass; `grep -r` for fable-only identifiers (`cc-harness-guide`, `decision-record`, `acceptance-run`, `workflow-conformance-verify`, `workflow-onboarding-advise`, `workflow-drift-detect`, `workflow-module-manage`) under `packages/` returns nothing.
4. **Evidence honesty.** Update subplan-05: uncheck criteria the fixture replay never demonstrated, record the executable skeleton run (with date and check count) as the integration evidence, keep `status: complete` only if the re-read supports it. Run the local Claude Code probe and record the observed runtime version in the capability matrix; amend `validation-policy.md`/`VALIDATION.txt` per Key Decision 4.
   *Verify:* no subplan claim exceeds recorded evidence (spot-audit of every checked criterion in subplan-05); `VALIDATION.txt` result line is derivable from validation-policy rule 5 as amended; the Claude Code matrix row carries a dated probe with observed runtime version and a rerun-on-version-change drift note; the other five remain explicitly candidate.
5. **Adversarial fixture corpus + runner.** Convert fable's 64 fixtures to JSON, dedupe against platform's existing fixture sets (dispositions recorded in the manifest), place them in the owning guides' assets, index every fixture from a suite reference, and write a `.mjs` runner in the established validator style (schema check, given/when/expect + `must_not` consistency, mutation guards) wired into moon test.
   *Verify:* runner exits 0 over the full corpus; fixture count in the suite indexes equals manifest count (ports minus recorded dedupes); zero orphan fixtures (every fixture referenced from an index, machine-checked by the runner); deliberately tampering with one fixture makes the runner fail (mutation guard demonstrated); 10-fixture spot-audit against the fable YAML originals recorded in the manifest; moon test green for both contract skill packages.
6. **Plugin `plans/` decoupling.** Split `validate-documentation.mjs` into package-scoped checks (shipped, no `plans/` knowledge) and a repo-level plan-traceability task (moon task outside the package, or in a repo script package); fix `docs/validation-and-traceability.md` links; remove `/plans/**` from the package `moon.yml` inputs.
   *Verify:* `grep -r 'plans/composable-workflow-phases' packages/command/command-workflow/` returns nothing; the package test task passes when run against the package tree alone; combined check count across the split tasks ≥ the 8,862 baseline (no silent coverage loss); repo-level task still validates plan traceability.
7. **Traceability repairs.** Fix D-036–D-039 cell counts and separator style; constrain classification values to the legend (extend the legend where a variant is genuinely needed); add automated shape checks to the repo-level validator: every table row's cell count matches its header, every classification value is in the legend, and any subplan whose task rows all share an identical ID block is flagged. Repair the worst blanket blocks (subplans 01, 05, 10) to per-task IDs.
   *Verify:* new shape checks pass on the repaired files and fail on a seeded regression (demonstrated once, then removed); re-run of the full doc validator green.
8. **Closure, release, and end-to-end gate.** Record both fable extras as skip-with-reason in the port manifest and graft the three convention rules (initiating-platform ownership of cross-platform flows; OIDC-preferred/no-static-keys-by-default; `SOURCES.md` pinning with retrieval dates as a release precondition) into `enterprise-platforms.md`. Apply the patch bump to 6.0.1 lockstep across all 93 packages with release notes naming the 9 relocated packages. Flip `composable-workflow-phases` parent to `status: complete` with `completed_date`. Regenerate `VALIDATION.txt`. Merge the worktree branch back to `main` and remove the worktree.
   *Verify:* the full success-criteria checklist below.

## Risk Assessment

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Layout moves break published-plugin consumers | Installed 6.0.0 plugins resolve stale paths | Settled: patch 6.0.1 with manifest-verified resolution (catalogs reference package roots only; cache snapshots per version); release notes name the 9 relocated packages |
| `main` advances while the merge worktree lives | Merge-back conflicts; stale baseline | Short-lived execution groups; rebase the worktree branch on `main` before merge-back; the primary checkout takes no merge-scope changes |
| Ported fixtures carry fable-only semantics | Fixtures reference commands/operations/guides that don't exist in platform | Identifier re-point during port + repo-wide grep check (workstream 3/5); runner validates every referenced operation resolves |
| YAML→JSON conversion changes fixture meaning | Adversarial scenarios silently weakened | Conversion is mechanical (structure-preserving); port manifest pairs each source/destination; spot-audit 10 fixtures against fable originals in review |
| Two skeleton representations confuse users | Recorded fixtures mistaken for execution evidence again | Single `walking-skeleton.md` narrative naming both artifacts and what each proves; subplan-05 cites only the executable run as integration evidence |
| Decoupling drops validation coverage | Checks silently lost in the package/repo split | Assert combined check count ≥ 8,862 baseline; both tasks wired into CI |
| Traceability repair scope creep | 150-row re-mapping stalls the merge | Non-goal: only defect repair + automated guards + three worst subplans |
| Fable repo drifts during the merge | "Correctly merged" becomes unverifiable | All ports pinned to `82303137`; manifest records the pin |

## Proposed Child Plans

| Group | Child plan | Purpose | Depends on |
| --- | --- | --- | --- |
| 1 | `merge-baseline-and-layout-normalization` | Baseline gate, port manifest, 9-package flat-layout move, manifest/catalog fixes | None |
| 2 | `merge-walking-skeleton-and-evidence-honesty` | Skeleton graft, subplan-05 evidence repair, Claude Code probe, validation-policy amendment | Group 1 |
| 2 | `merge-adversarial-fixture-corpus-and-runner` | Fixture conversion, dedupe, suite indexes, runner + mutation guards, moon wiring | Group 1 |
| 2 | `merge-plugin-plans-decoupling` | Package/repo validator split, docs link fixes, moon input cleanup | Group 1 |
| 3 | `merge-traceability-repairs` | D-036–D-039 fix, vocabulary legend, shape checks, worst-offender ID blocks | Group 2 (validator split) |
| 4 | `merge-closure-release-and-validation` | Optional extras, semver application, parent-plan closure, VALIDATION regeneration, E2E gate | Group 3 |

## Success Criteria

Merge-correctness (the "everything landed" proof):

- The port manifest shows a disposition for 100% of in-scope fable files at `82303137`, with zero unexplained skips; a script (or recorded audit) confirms every `port`/`adapt` destination file exists.
- `grep -r` for all fable-only identifiers under `packages/` returns nothing.
- Fixture suite indexes account for every ported fixture; runner-enforced zero-orphan check passes.

Executable checks (all must exit 0, wired into CI where noted):

- `run-skeleton.sh --yes`: 17/17 checks; consent stop without `--yes`; shellcheck clean; workspace and repo clean afterwards (CI: moon test).
- Fixture runner over the full corpus, including a demonstrated mutation-guard failure (CI: moon test).
- Existing platform validators: `validate-walking-skeleton-fixtures.mjs` and siblings still pass (CI).
- Split documentation validators: package-scoped + repo-level, combined checks ≥ 8,862 (CI).
- `go test ./internal/webhook/` (CI).
- Full moon gate: format, lint, typecheck, build, test across all affected projects.

Structure and bookkeeping:

- Zero `skills/`-nested guide dirs; zero `"skills": "./skills"` manifests; 93/93 marketplace parity with all paths resolving; skill validators pass for all 17 new packages.
- No `plans/` path references inside `packages/command/command-workflow/`.
- All traceability tables shape-valid (machine-checked); classification values legend-constrained; D-036–D-039 render as 5-column rows.
- Subplan-05 claims match recorded evidence; Claude Code matrix carries a dated probe with observed runtime version and rerun-on-version-change drift note; `VALIDATION.txt` result is qualified per amended rule 5.
- `enterprise-platforms.md` contains the three grafted convention rules; the port manifest records both fable extras as skip-with-reason.
- `composable-workflow-phases` parent: `status: complete` with `completed_date`; this plan's own status updated by `plan-update` as groups land.
- Patch release 6.0.1 applied lockstep-consistent across all 93 packages (machine-checked version sweep); release notes name the 9 relocated packages.
- All merge work committed on the `composable-workflow-implementations-merge` worktree branch, merged back to `main`, and the `xonovex-platform-merge` worktree removed (`git worktree list` clean); the primary checkout carries no uncommitted merge-scope changes at any point.

## Estimated Effort

**Medium.** Six subplans, mostly mechanical porting and repair with two design-sensitive spots: the package/repo validator split (workstream 6) and the fixture YAML→JSON conversion at volume (workstream 5). The walking-skeleton graft and layout moves are low-risk; the end-to-end gate re-uses existing CI machinery. All decisions are settled (see Key Decisions); the plan is ready for `plan-critique` / `plan-accept`.
