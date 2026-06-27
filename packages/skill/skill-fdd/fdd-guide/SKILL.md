---
name: fdd-guide
description: "Use when running Feature-Driven Development its process scaffolding — the five processes (Develop an Overall Model, Build a Features List, Plan by Feature, Design by Feature, Build by Feature), naming features as '<action> the <result> by/for/of/to a(n) <object>', decomposing into feature sets and major feature sets, tracking the six weighted build milestones, and organising Chief Architect / Chief Programmer / Class Owner roles and per-feature-set teams. Triggers on FDD, feature-driven development, feature list, design by feature, build by feature, feature-naming template, milestone percent-complete, class owner, chief programmer — even when the user doesn not say 'FDD'. Skip defining the domain object model itself (see **ddd-guide**) and the mechanics of client value / INVEST-style slicing (see **user-stories-guide**)."
---

# Feature-Driven Development

FDD is a model-driven, short-iteration delivery rhythm built from five processes, a strict feature-naming template, a feature-set hierarchy, and six weighted milestones that report exact percent-complete per feature. It owns the choreography and tracking; the domain model and the notion of client value are owned elsewhere.

## Essentials

- **Run the five processes in order** - Develop an Overall Model, Build a Features List, Plan by Feature, then iterate Design by Feature and Build by Feature, see [references/fdd-processes.md](references/fdd-processes.md)
- **Name every feature `<action> the <result> by/for/of/to a(n) <object>`** - the connector is any of by/for/of/to, see [references/fdd-processes.md](references/fdd-processes.md)
- **Decompose features into feature sets into major feature sets** - features group by business activity, feature sets group into subject areas, see [references/fdd-processes.md](references/fdd-processes.md)
- **Size each feature at two weeks maximum** - most take 1-3 days; split any business step larger than two weeks, see [references/fdd-processes.md](references/fdd-processes.md)
- **Track build with six weighted milestones summing to 100%** - 1/40/3/45/10/1 from walkthrough to promote, see [references/fdd-processes.md](references/fdd-processes.md)
- **Staff per-feature-set teams under one Chief Programmer with per-class Class Owners** - feature teams form dynamically, see [references/fdd-processes.md](references/fdd-processes.md)

## Gotchas

- The two-week limit is a feature's MAX SIZE, not a Scrum-style sprint timebox — FDD predates and does not use sprints; most features take 1-3 days.
- The feature template connector is by/for/of/to, not always "of"; writing only `<action> the <result> of a(n) <object>` narrows the pattern.
- Process 1 (Develop an Overall Model) is domain modelling owned by **ddd-guide** — FDD contributes the process placement, not a new modelling theory.
- A "feature team" is not a "feature" and is not a named role — it is a transient team formed per feature set under a Chief Programmer.
- The six milestone weights are exact (Code 45% and Design 40% dominate); an even split is folklore. They sum to 100%.
- FDD uses single-owner-per-class (Class Owner) and design/code inspections — it is not interchangeable with collective code ownership or XP/Scrum practices.

## Example

```
Major feature set: Order Management
  Feature set: Processing an order
    Feature: Calculate the total of an order
    Feature: Apply the discount to an order
    Feature: Validate the shipping address of an order
    Feature: Reserve the inventory for an order line
    Feature: Authorize the payment for an order        <- in build
        Domain Walkthrough  1%   [x]
        Design             40%   [x]
        Design Inspection   3%   [x]
        Code               45%   [ ]   <- 44% complete
        Code Inspection    10%   [ ]
        Promote to Build    1%   [ ]
```

## Progressive Disclosure

- Read [references/fdd-processes.md](references/fdd-processes.md) - Load when running the five processes, naming features, sizing/decomposing the feature list, tracking the six milestones, or staffing roles and feature teams
