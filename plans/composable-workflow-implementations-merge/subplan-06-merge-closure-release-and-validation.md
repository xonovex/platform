---
type: plan
has_subplans: false
parent_plan: composable-workflow-implementations-merge
parallel_group: 4
status: in_progress
dependencies:
  plans:
  - merge-walking-skeleton-and-evidence-honesty
  - merge-adversarial-fixture-corpus-and-runner
  - merge-plugin-plans-decoupling
  - merge-traceability-repairs
  files:
  - packages/skill/skill-agent-governance/agent-governance-guide/references/enterprise-platforms.md
  - plans/composable-workflow-phases.md
  - plans/composable-workflow-phases/VALIDATION.txt
  - plans/composable-workflow-implementations-merge/port-manifest.md
  - packages/**/package.json
skills_to_consult:
- versioning-guide
- git-guide
- skill-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 06: Closure, Release, and End-to-End Validation

## Objective

Graft the three convention rules, complete and audit the port manifest, ship the 6.0.1 lockstep patch, close the parent plans, regenerate honest validation evidence, and merge the worktree back to `main`.

## Tasks

1. **Graft the convention rules.** Add to `references/enterprise-platforms.md` (as edits in its existing sections, not a new file): cross-platform flows are owned by the initiating platform and cross-reference the receiving platform's skill; OIDC/workload federation preferred with long-lived static keys never a default (shared clause, not just per-skill); `SOURCES.md` pinning with retrieval dates as a release precondition for platform claims.
2. **Complete the manifest audit.** Record `two-plane-architecture.{dot,png}` and `platform-skill-convention.md` as skip-with-reason; verify every manifest row has a final status; script-check that every `ported`/`adapted` destination exists on disk and every `skip` has a reason.
3. **Apply the 6.0.1 lockstep bump** using the repo's version-bump moon action across all 93 packages; write release notes explicitly naming the 9 relocated packages; machine-check the version sweep (no package left at 6.0.0).
4. **Close the parent plans.** Flip `plans/composable-workflow-phases.md` to `status: complete` with `completed_date`; regenerate `plans/composable-workflow-phases/VALIDATION.txt` with the qualified result line per amended rule 5, the executable-skeleton evidence, the fixture-runner counts, and the Claude Code probe record. Update this merge plan's own status via `plan-update`.
5. **End-to-end gate.** Full moon validation (format, lint, typecheck, build, test — includes skeleton run, fixture runners, split doc validators), `go test ./internal/webhook/`, and the three grep sweeps (fable-only identifiers; nested `skills/` layouts; `plans/` references in the shipped package). Compare against `BASELINE.txt` — nothing green at baseline may be red now.
6. **Merge back.** Rebase `composable-workflow-implementations-merge` on `main`, re-run the gate if the rebase changed anything, then merge to `main` via the worktree-merge flow and remove the worktree. **This is a mandatory-human gate — do not merge without explicit user confirmation.** Never push; release goes via PR per repo policy.

## Validation Steps

- Manifest audit script/recorded audit: all rows final, all destinations exist, all skips reasoned.
- Version sweep: 93/93 manifests and package.json at 6.0.1; marketplace parity intact.
- `VALIDATION.txt` result line derivable from amended rule 5; no claim exceeds recorded evidence.
- E2E gate green and ≥ baseline; grep sweeps empty.
- After merge-back: `git worktree list` shows no `xonovex-platform-merge`; `main` contains all merge commits; primary checkout clean.

## Success Criteria

- [x] Three convention rules present in `enterprise-platforms.md` (initiating-platform ownership; platform-neutral static-credential prohibition; SOURCES.md pinning + probe precondition); both extras recorded as reasoned skips
- [x] Port manifest 100% final and audit-clean (63 rows, zero pending, all destinations exist)
- [x] 6.0.1 lockstep across 93 packages, both marketplaces, 186 manifests, lockfile; release notes name the 9 relocated packages (bump commit body + MIGRATION.md 6.0.1 section)
- [x] `composable-workflow-phases` complete with `completed_date: 2026-07-16`; `VALIDATION.txt` regenerated with the qualified result line per amended rule 5
- [x] E2E gate green (`:ci-check` + envtest integration = 773 tasks, exit 0; ≥ baseline with e2e/kind excluded identically); all three grep sweeps empty
- [ ] Worktree merged back to `main` (user-confirmed) and removed — **mandatory-human gate, awaiting confirmation**

## Files Modified/Created

- Modified: `enterprise-platforms.md`, `plans/composable-workflow-phases.md`, `VALIDATION.txt`, port manifest, 93 × `package.json`/plugin manifests, release notes/changelog

## Dependencies

All prior subplans (group 4 — final).

## Estimated Duration

0.5–1 day.
