# Subplan Example

One child plan, then the ordering:

```markdown
---
type: plan
feature: reader-pull
has_subplans: false
parent_plan: plans/order-import-backpressure.md
parallel_group: 1
status: pending
dependencies:
  plans: []
  files:
    - src/import/order_reader.ts # the seam this child converts
skills_to_consult: [writing-typescript, authoring-tests]
validation:
  build: pending
  tests: pending
---

# Subplan 01: Reader Pull

## Objective

One paragraph: what this child alone delivers.

## Tasks

1. Five to seven, each naming its files and the proof that it landed.

## Validation Steps

1. The project's real task-runner commands, not raw tool invocations.

## Success Criteria

- [ ] Checkable, owned by this child alone

## Files Modified/Created

## Dependencies

## Estimated Duration
```

```text
Ordering: group 1 reader-pull; group 2 batcher-rework after it. Both
edit src/queue/batcher.ts, so they are sequenced, never concurrent.
```
