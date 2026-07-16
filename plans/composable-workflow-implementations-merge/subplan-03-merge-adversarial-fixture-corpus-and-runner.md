---
type: plan
has_subplans: false
parent_plan: composable-workflow-implementations-merge
parallel_group: 3
status: pending
dependencies:
  plans:
  - merge-baseline-and-layout-normalization
  - merge-walking-skeleton-and-evidence-honesty
  files:
  - packages/skill/skill-agent-governance/agent-governance-guide/assets/fixtures/**
  - packages/skill/skill-agent-governance/agent-governance-guide/scripts/**
  - packages/skill/skill-workflow/workflow-guide/assets/fixtures/**
  - packages/skill/skill-workflow/workflow-guide/scripts/**
  - packages/skill/skill-agent-governance/moon.yml
  - packages/skill/skill-workflow/moon.yml
  - plans/composable-workflow-implementations-merge/port-manifest.md
skills_to_consult:
- typescript-guide
- testing-guide
- skill-guide
- moon-guide
validation:
  type_check: pending
  lint: pending
  build: pending
  tests: pending
  integration: pending
---

# Subplan 03: Adversarial Fixture Corpus and Runner

## Objective

Port fable's 64 adversarial fixtures as JSON into the owning platform guides, index every fixture into a suite, and give the corpus a machine runner with mutation guards wired into moon test — closing the "no runner for the fixtures" gap.

## Tasks

1. **Extract and convert.** Read each fixture from the pin (26 governance, 20 workflow, ~18 harness/enterprise) via `git -C ../xonovex-platform-fable show 82303137:<path>`; convert YAML → JSON structure-preserving (folded scalars become strings). Place under the owning guide's `assets/fixtures/` (governance and workflow in their contract skills; harness/enterprise fixtures follow the platform's centralized-fixture convention in `skill-agent-governance`). Pair every destination with its source in the port manifest.
2. **Re-point identifiers and dedupe.** Replace fable-only command/operation/guide names with platform equivalents. Compare against platform's existing fixture sets (`harness-conformance-fixtures.json`, `enterprise-platform-fixtures.json`, `external-enforcement-fixtures.json`, walking-skeleton fixtures); where a fable scenario duplicates an existing case, record `skip-with-reason: duplicate of <file>#<case>` in the manifest instead of porting.
3. **Index into suites.** Add or extend suite reference files (e.g., `references/conformance.md` fixture tables) so every ported fixture is referenced exactly once with its owning contract; no orphans.
4. **Write the runner.** New `scripts/validate-conformance-fixtures.mjs` per contract skill, in the established validator style: JSON schema check (id, contract, scenario, given/when/expect, `must_not` non-empty), referential checks (referenced operations/contracts exist as reference files), orphan detection (fixture ↔ suite index bijection), aggregate-and-throw failures, plus mutation guards verifying the runner detects a tampered fixture.
5. **Wire into CI.** Add the runners to both packages' `package.json` test scripts / moon test tasks alongside the existing validators.
6. **Spot-audit.** Compare 10 randomly chosen converted fixtures field-by-field against their fable YAML originals; record the audit (files, date, result) in the port manifest.
7. **Validate and commit** as `feat(skill): port adversarial conformance fixtures with machine runner`.

## Validation Steps

- Runner exits 0 over the full corpus in both packages; mutation guard demonstrably fails on a tampered fixture.
- Suite-index fixture count == manifest `ported` count (ports minus recorded dedupes); orphan check green.
- `grep -rn` for fable-only identifiers under `packages/` → empty.
- Existing validators unaffected; moon test green for both contract skills; repo lint/format green.

## Success Criteria

- [ ] Every fable fixture has a manifest disposition: ported (JSON) or skip-with-reason
- [ ] Zero orphan fixtures, machine-checked; suites reference every ported fixture once
- [ ] Runner + mutation guards wired into moon test for both contract skills
- [ ] 10-fixture spot-audit recorded, zero semantic deviations
- [ ] Full validation green; one conventional commit

## Files Modified/Created

- Created: `~64` fixture JSONs under the owning guides' `assets/fixtures/`; `scripts/validate-conformance-fixtures.mjs` × 2
- Modified: suite reference files, both packages' `moon.yml`/`package.json`, port manifest

## Dependencies

Subplans 01 (manifest) and 02 (shares `skill-agent-governance` moon.yml and reference index — must land first).

## Estimated Duration

1–2 days (largest subplan by volume).
