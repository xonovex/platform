---
type: plan
has_subplans: false
parent_plan: composable-workflow-implementations-merge
parallel_group: 3
status: complete
dependencies:
  plans:
  - merge-plugin-plans-decoupling
  files:
  - plans/composable-workflow-phases/traceability/decision-source-matrix.md
  - plans/composable-workflow-phases/traceability/subplan-traceability.md
  - packages/script/**
skills_to_consult:
- plan-guide
- typescript-guide
- testing-guide
validation:
  type_check: pass
  lint: pass
  build: pass
  tests: pass
  integration: pass
---

# Subplan 05: Traceability Repairs

## Objective

Repair the shared traceability defects both implementations inherited — broken table rows, uncontrolled classification vocabulary, blanket per-task ID blocks — and add automated shape checks to the repo-level validator so they cannot recur.

## Tasks

1. **Fix D-036–D-039.** In `plans/composable-workflow-phases/traceability/decision-source-matrix.md`, repair the four rows with 6 cells against the 5-column header (stray pipe splitting the rationale) and restore the `,` source-separator convention used by rows D-001–D-035.
2. **Constrain the classification vocabulary.** Inventory all classification values in the matrix (~15 ad-hoc variants against a 4-class legend); map each variant to a legend class, or deliberately extend the legend where a variant carries real meaning. Every row value must afterwards appear in the legend.
3. **Repair blanket ID blocks.** In `subplan-traceability.md`, replace the copy-pasted identical decision/control/source ID sets on the worst offenders (subplans 01, 05, 10) with per-task-specific IDs derived from what each task actually implements.
4. **Add shape checks** to the repo-level traceability validator (created by subplan 04): every table row's cell count equals its header's; every classification value ∈ legend; flag any subplan whose task rows all carry an identical ID block.
5. **Prove the checks bite.** Seed one deliberate regression per check (extra cell, off-legend value, blanket block), confirm the validator fails, remove the seeds, confirm green. Do not commit the seeds.
6. **Validate and commit** as `fix(plan): repair traceability table shape, vocabulary, and per-task mappings`.

## Validation Steps

- Repo-level traceability validator green, including the three new shape checks.
- Demonstrated red→green cycle for each new check (recorded in the subplan on completion).
- `decision-source-matrix.md`: 39 data rows, all 5 cells; all classifications legend-valid.
- Full repo validation green.

## Success Criteria

- [x] D-036–D-039 render as 5-column rows with comma separators (39/39 rows shape-valid)
- [x] Classification vocabulary fully legend-constrained: 23 ad-hoc variants mapped to the 4 legend classes per-row (synthesis never reclassified as derived, honoring validation-policy rule 6)
- [x] Per-task ID sets: subplans 10 and 11 fully repaired (were 15/15 blanket), plus the nine-row repeat groups in subplans 01 and 05; every narrowed ID drawn from its subplan's recorded union so all references still resolve; max identical-cell repeat now ≤ 3 in the repaired subplans
- [x] Three new shape checks (row cell-count vs header across all five traceability artifacts; legend-constrained classifications; identical-ID-block flag) with vitest-tested helpers, each proven to fail on a live seeded regression and pass after restore. Note: the validator's check count dropped 8,800 → 7,249 because ~1,550 duplicated ID references were removed from the repaired artifact — the check set is a superset of the split baseline; parent-plan criterion reconciled
- [x] One conventional commit (plus this bookkeeping)

## Files Modified/Created

- Modified: `plans/composable-workflow-phases/traceability/{decision-source-matrix.md,subplan-traceability.md}`, repo-level validator under `packages/script/`

## Dependencies

Subplan 04 (the repo-level validator it extends). Disjoint from subplan 03; parallel-safe in group 3.

## Estimated Duration

0.5 day.
