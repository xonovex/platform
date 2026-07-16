---
type: plan
has_subplans: true
status: approved
updated: '2026-07-16'
dependencies:
  plans:
  - composable-workflow-implementations-merge
  subplans:
  - e2e-remediation-and-ci-preparation
  - runtime-probe-expansion
  - dependency-advisory-remediation
proposed_subplans:
- e2e-remediation-and-ci-preparation
- runtime-probe-expansion
- dependency-advisory-remediation
parallel_groups:
- group: 1
  plans:
  - e2e-remediation-and-ci-preparation
  - runtime-probe-expansion
  - dependency-advisory-remediation
skills_to_consult:
- github-guide
- git-guide
- npm-guide
- testing-guide
- moon-guide
- shell-scripting-guide
research_sources:
  retrieved: '2026-07-16'
  documentation:
  - plans/composable-workflow-phases/VALIDATION.txt (standing caveats)
  - plans/composable-workflow-implementations-merge/BASELINE.txt (environment exclusions)
  - packages/skill/skill-claude-code/code-harness-guide/references/capabilities.md (probe precedent)
  - independent plan critique 2026-07-16 (11 findings; facts verified against remote and CI config)
  versions:
    local-main: 6.0.1 at cf42bce1
    remote-main-verified: 166c4f26 (5.0.0; the 6.0.0 release was never pushed — 24 local commits)
    claude-code-probed: 2.1.211
    kernel: 7.0.6-gentoo (Kind/runc setns failure host)
    npm-audit: 6 advisories (1 low, 2 moderate, 3 high), all fixAvailable true
---

# Environment Hardening (Local-First)

## Overview

Close out the environment caveats the merge validation left open, entirely locally: fix the operator test wiring and prepare (but do not activate) CI coverage for the Kind e2e suites, expand runtime probes where honestly possible, and clear the six npm-audit advisories. By owner decision there are **no PRs and no pushes** in this plan: remote `main` stays at `166c4f26`, the 24 local commits (spanning the unpublished 6.0.0 and 6.0.1 releases) remain local, the version stays 6.0.1, and publishing is deferred by necessity. The independent-review Definition-of-Done item is consciously deferred as an owner decision and recorded here rather than silently dropped.

## Goals

- `go-test-integration` runs green from a bare `moon run` with no manually exported environment (`KUBEBUILDER_ASSETS` wired in the task itself, `script:` form).
- The fourth Kind suite (`e2e-coco`, currently invisible to moon) gets a moon task alongside the existing three.
- A CI workflow for the e2e suites is authored, validated locally as far as possible, and committed **dormant** — ready to run the day the owner pushes; kata/coco runs must report skipped-vs-executed counts so a pass-by-skip is visible.
- A best-effort investigation of the host Kind/runc `setns` failure is time-boxed and documented, whatever the outcome.
- Every harness whose guard contract is exercisable with credentials already on this host gets a runtime probe recorded per the Claude Code precedent; every other candidate gets an executable probe runbook; version-only observations claim exactly that and nothing more.
- `npm audit` reports zero advisories via plain transitive updates.
- VALIDATION.txt reflects reality after each change: candidate counts, the e2e disposition, the advisory note, and which caveats are permanent (org-specific reviews, crosswalk non-certification) with a named owner status.
- The deferred items (push, review, publish) are recorded in the plan and VALIDATION.txt as owner decisions with their reactivation conditions.

## Non-goals

- Pushing anything to the remote, opening PRs, or triggering `release.yml` — owner-deferred; the trigger facts (PR title containing "version packages", or `workflow_dispatch` with `dry_run=false`) are recorded for that day.
- Any version change — the tree stays at 6.0.1.
- Live enterprise-tenant probes (Azure DevOps, Bitbucket, Bitrise, AWS, Datadog) — credential-gated, runbook-only.
- Purchasing subscriptions or creating accounts to enable harness probes (Copilot, Codex sign-ins etc. only if already present).
- Replacing the host OS/container runtime.

## Current State (verified, post-critique)

- Remote `main` is `166c4f26` (5.0.0). The entire 6.0.0 release (`e18d54f5`) and the 6.0.1 merge (12 further commits to `cf42bce1`) exist only locally — 24 unpushed commits. Local `main` and the merge worktree branch both sit at `cf42bce1`; the fable worktree holds 12 commits unreachable from `cf42bce1` (superseded reference implementation).
- Branch protection on remote `main`: one approving review required, `required_linear_history: true`, **no** required status checks, **no** CODEOWNERS file, `enforce_admins: false`. Recorded for whenever pushing resumes; irrelevant while local.
- `ci.yml` runs only `:ci-check`, which excludes `go-test-integration` and all e2e suites. There are **four** Kind-based suites (`e2e`, `e2e-gvisor`, `e2e-kata`, `e2e-coco`); only three have moon tasks. Kind cluster creation fails on this host (docker/runc `setns` error, kernel 7.0.6-gentoo). `go-test-integration` works only when the caller exports `KUBEBUILDER_ASSETS` manually; `setup-envtest`, `kind`, and `kubectl` are already in the nix devshell.
- Capability matrices: Claude Code probed (2.1.211); Codex, Kiro, Copilot, Pi, OpenCode and the five enterprise platforms are documentation-derived candidates. All five harnesses are free-ish to install but credential-gated at runtime (Codex: OpenAI sign-in; Copilot: paid subscription; Kiro: AWS Builder ID, desktop-oriented; Pi/OpenCode: model API key) — install-ability is not the operative boundary.
- `npm audit`: six advisories (@babel/core, form-data, js-yaml, linkify-it, markdown-it, undici), all with plain `fixAvailable: true`; the release validator checks lockfile versions only for workspace plugin packages, so transitive bumps don't fight it.
- The kata suite's VM-isolation test self-skips in unprivileged Kind; a "green" run may prove little without skip-count reporting.

## Key Decisions

1. **No PRs — the work ships as direct commits, version stays 6.0.1** (owner decision, 2026-07-16). This conflicts with the repo's release-via-PR policy in AGENTS.md and is recorded as the owner's explicit prerogative; the independent-review DoD item is deferred with this plan as its record, to be satisfied whenever pushing/review resumes.
2. **No push** (owner decision, 2026-07-16). Everything in this plan lands as local commits on the existing branch/main pair. Publishing is thereby deferred by necessity. Reactivation facts recorded in Non-goals.
3. **E2e strategy: fix locally what is fixable, prepare CI dormant.** The task wiring (`KUBEBUILDER_ASSETS` via `setup-envtest` in `script:` form) and the missing `e2e-coco` moon task are real local fixes. The CI workflow for integration + e2e suites is authored and committed but cannot run until a push happens — it is validated by local lint/dry-parse only, and this limitation is stated, not hidden. Host runc investigation is time-boxed (half a day) and documented either way. Consequence accepted: the release remains e2e-unexercised until the owner pushes; VALIDATION.txt keeps saying so.
4. **Probe boundary: "guard contract exercisable with credentials already present on this host."** Expected outcomes enumerated up front: each of the five harnesses gets exactly one of — full probe (version + contract, only if a working credentialed install already exists), version-only observation (recorded as "CLI version observed; contract not exercised — remains documentation-derived for hook semantics"), or runbook. No credential acquisition. The qualified result line counts only full probes as non-candidates.
5. **Audit remediation via plain transitive updates** (`npm audit fix` / targeted `npm update`), no `overrides`, no waiver machinery — all six advisories have plain fixes available. A waiver remains a one-line contingency if an update regresses the gate.
6. **Fable worktree dispositioned as superseded.** Its 12 unreachable commits are the reference implementation whose valuable parts were ported (per the port manifest); the worktree is retained or abandoned at the owner's convenience via `plan-worktree-abandon`/`cleanup` — recorded, not required.

## Proposed Approach

1. **Operator test wiring and CI preparation.** Convert `go-test-integration` to a `script:` task that wires `KUBEBUILDER_ASSETS=$(setup-envtest use -i -p path)`; add a `go-test-e2e-coco` moon task mirroring its siblings; author `.github/workflows/e2e.yml` running integration + the four e2e suites on a hosted runner (nix devshell, manual/nightly trigger, skip-vs-run counts surfaced per suite); time-boxed host runc investigation, outcome documented in the operator README; update VALIDATION.txt's exclusion note to reference the prepared-but-dormant coverage and the no-push deferral.
   *Verify:* bare `moon run agent-operator-go:go-test-integration` green with no exported env; `moon run agent-operator-go:go-test-e2e-coco` exists and fails only on the known host runtime error; workflow file passes actionlint/parse; VALIDATION.txt updated; full `:ci-check` green.
2. **Runtime probe expansion.** Inventory the five harnesses against Decision 4's boundary on this host; execute full probes where a credentialed install already exists; record version-only observations with their exact non-claims; write probe runbooks (commands, expected evidence, matrix fields) into each remaining candidate's guide plus the five enterprise skills; update matrices and the qualified result line to match; declare the permanent org-review caveats (AWS trust policies, Datadog data collection, crosswalk non-certification) as standing with named ownership in VALIDATION.txt.
   *Verify:* every matrix row's evidence status matches what actually ran (spot-audit); candidate count in the result line equals the matrices; release validator green.
3. **Dependency advisory remediation.** `npm audit --json` inventory; targeted transitive updates; audit clean; full gate green; VALIDATION.txt advisory note replaced with the clean state and date.
   *Verify:* `npm audit` exits 0; plugin versions untouched at 6.0.1 (machine sweep); `:ci-check` green.

Each workstream commits locally on the existing branch, fast-forwarded into local `main` as before.

## Risk Assessment

| Risk | Impact | Mitigation |
| --- | --- | --- |
| The release stays e2e-unexercised while unpushed | A real operator defect stays hidden until a future push runs the dormant CI | Accepted explicitly (Decision 3); VALIDATION.txt states it; host investigation may close it early |
| Dormant CI workflow is wrong on first real run | Broken gate discovered on the eventual push day | actionlint + parse validation now; keep the workflow minimal; first-run expectations documented in the workflow header |
| Version-only probes read as more than they are | Overclaiming — the exact failure the merge fixed elsewhere | Decision 4 defines the claim text verbatim; release validator's honesty guards still active |
| Transitive updates regress behavior | A "chore" breaks the gate | Full `:ci-check` after each update; waiver contingency; isolated commit for easy revert |
| Deferred review/push is forgotten | DoD item silently rots | Recorded in this plan, VALIDATION.txt, and the parent plan trail with reactivation conditions |

## Proposed Child Plans

| Group | Child plan | Purpose | Depends on |
| --- | --- | --- | --- |
| 1 | `e2e-remediation-and-ci-preparation` | Task wiring fix, e2e-coco moon task, dormant CI workflow, host investigation, VALIDATION update | None |
| 1 | `runtime-probe-expansion` | Probes/version-observations/runbooks per Decision 4, matrix + result-line reconciliation, caveat ownership | None |
| 1 | `dependency-advisory-remediation` | Clear the six advisories via plain updates, isolated commit | None |

All three are mutually independent and may execute in any order; their files are disjoint except `plans/composable-workflow-phases/VALIDATION.txt`, where each subplan edits only its own section (single-worktree execution serializes the writes).

## Success Criteria

- `moon run agent-operator-go:go-test-integration` passes with no manually exported environment.
- All four e2e suites have moon tasks; the three runnable-on-a-working-host suites still fail here only with the documented runc error; skip-vs-run reporting exists for kata/coco.
- `.github/workflows/e2e.yml` exists, parses clean, and states its dormant status and first-run expectations; nothing was pushed.
- Every harness matrix row carries evidence matching exactly what ran (full probe / version-only / runbook), and the qualified result line's candidate count equals the matrices.
- `npm audit` exits clean; all packages remain at 6.0.1 (machine sweep).
- VALIDATION.txt records: the no-push/no-PR/no-publish owner deferrals with reactivation conditions, the e2e disposition, the probe outcomes, the clean audit, and the permanent-vs-closable caveat split.
- Full `:ci-check` green after each workstream; local `main` and the branch stay in sync; remote untouched at `166c4f26`.

## Estimated Effort

**Small.** Three independent local workstreams: the operator wiring and workflow authoring are a few files; probes likely resolve to version-only observations plus runbooks (no credential negotiation); the audit fix is mechanical. The largest single unknown is the time-boxed host runc investigation, which is capped by design.
