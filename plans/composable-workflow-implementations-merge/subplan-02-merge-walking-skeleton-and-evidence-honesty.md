---
type: plan
has_subplans: false
parent_plan: composable-workflow-implementations-merge
parallel_group: 2
status: complete
dependencies:
  plans:
  - merge-baseline-and-layout-normalization
  files:
  - packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/**
  - packages/skill/skill-agent-governance/agent-governance-guide/references/walking-skeleton.md
  - packages/skill/skill-agent-governance/moon.yml
  - packages/skill/skill-claude-code/code-harness-guide/references/**
  - plans/composable-workflow-phases/subplan-05*.md
  - plans/composable-workflow-phases/traceability/validation-policy.md
  - plans/composable-workflow-implementations-merge/port-manifest.md
skills_to_consult:
- shell-scripting-guide
- skill-guide
- testing-guide
- moon-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 02: Walking-Skeleton Graft and Evidence Honesty

## Objective

Port fable's executable walking skeleton into `skill-agent-governance`, run it as real evidence, execute the local Claude Code probe, and repair the overstated subplan-05 and validation-policy claims.

## Tasks

1. **Port the skeleton assets.** Extract from the pin into `packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/`:
   ```bash
   git -C ../xonovex-platform-fable show 82303137:packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/guard.sh
   git -C ../xonovex-platform-fable show 82303137:packages/skill/skill-agent-governance/agent-governance-guide/assets/walking-skeleton/run-skeleton.sh
   ```
   plus any fixture files the scripts read. Re-point fable-only identifiers to platform equivalents (`cc-harness-guide` → `code-harness-guide`; `decision-record`/`acceptance-run`/`workflow-conformance-verify`/`workflow-onboarding-advise` → platform command/operation names). Update port-manifest rows to `ported`.
2. **Prove execution.** `bash run-skeleton.sh --yes` → 17/17 checks, exit 0; without `--yes` → stops after preview (consent gate); `shellcheck guard.sh run-skeleton.sh` clean; `git status` clean after runs (temp-workspace only).
3. **Reconcile the skeleton narrative.** Rewrite `references/walking-skeleton.md` to distinguish the retained recorded fixture scenario (validated by the existing `.mjs` validators) from the executable run (what it actually demonstrates: advisory-first consent, idempotent apply, drift detect/remediate, checksum module trust, evidence dedup, rollback). Update the SKILL.md reference index if the entry description changes.
4. **Wire into CI.** Add the skeleton run (`run-skeleton.sh --yes`) to the `skill-agent-governance` moon test task so every CI run executes it.
5. **Run the Claude Code probe.** Execute the skeleton/probe against the locally installed Claude Code; record in `packages/skill/skill-claude-code/code-harness-guide/references/` capability matrix: observed runtime version, probe date, and a rerun-on-version-change drift note. The other five harness matrices stay explicitly `candidate`.
6. **Repair subplan-05.** In `plans/composable-workflow-phases/subplan-05*.md`: uncheck criteria the fixture replay never demonstrated (e.g., "discovers the actual environment"), cite the executable run (date, 17/17) as integration evidence, and keep `integration: passed` only if the re-read supports it.
7. **Amend validation-policy rule 5.** Unprobed capabilities carry explicit `candidate` status; the overall result line must be qualified (e.g., `PASS (candidate matrices: 5 harnesses, 5 platforms unprobed)`). `VALIDATION.txt` regeneration itself happens in subplan 06. Commit as `feat(skill): graft executable walking skeleton and honest evidence`.

## Validation Steps

- Skeleton: 17/17 with `--yes`, preview-stop without, shellcheck clean, workspace clean.
- Existing validators still green: `node packages/skill/skill-agent-governance/agent-governance-guide/scripts/validate-walking-skeleton-fixtures.mjs`.
- `grep -rn 'cc-harness-guide\|decision-record\|acceptance-run\|workflow-conformance-verify\|workflow-onboarding-advise\|workflow-drift-detect\|workflow-module-manage' packages/` → empty.
- moon test for `skill-agent-governance` runs the skeleton and passes.
- Spot-audit every checked criterion in subplan-05 against recorded evidence.

## Success Criteria

- [x] Skeleton executes in CI (moon test) with 17/17 checks; consent gate proven (stops after preview without `--yes`); shellcheck clean; repo untouched after runs
- [x] No fable-only identifiers anywhere under `packages/` (scripts were self-contained; grep sweep clean)
- [x] `walking-skeleton.md` distinguishes recorded scenario from executable run
- [x] Claude Code matrix records `2.1.211 (Claude Code)` probed 2026-07-16 with rerun-on-version-change note; hook-level rows explicitly stay documentation-verified; five other harnesses unchanged (candidate)
- [x] Subplan-05 has an Evidence section separating executed from fixture-recorded claims, per-criterion labels; validation-policy rule 5 amended (candidate status + qualified result line mandatory)
- [x] Port manifest updated (skeleton rows done, probe recorded); one conventional commit

## Files Modified/Created

- Created: `.../agent-governance-guide/assets/walking-skeleton/{guard.sh,run-skeleton.sh,...}`
- Modified: `.../references/walking-skeleton.md`, `skill-agent-governance/moon.yml`, `skill-claude-code/code-harness-guide/references/*`, `plans/composable-workflow-phases/subplan-05*.md`, `.../traceability/validation-policy.md`, port manifest

## Dependencies

Subplan 01 (baseline + manifest). Shares `skill-agent-governance` package files with subplan 03 — runs before it (group 2 vs 3).

## Estimated Duration

0.5–1 day.
