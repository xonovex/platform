---
name: ux-design-guide
description: "Use when turning research and product intent into an end-to-end interaction design that must be explored and evaluated before implementation. Triggers on journey design, task flows, information architecture, interaction states, content hierarchy, wireframes, prototypes, design hypotheses, usability findings, design rationale, or implementation handoff, even when the user doesn't say UX design."
---

# UX Design

Turn evidence about users and outcomes into a coherent, testable interaction design
without confusing an attractive screen, a provider file, or prototype code with a
validated production solution.

## Essentials

- **Frame the design problem** — pin users, needs, whole journey, outcome, constraints,
  exclusions, research evidence, and unresolved assumptions
- **Model every meaningful state** — cover entry, success, empty, loading, error,
  interruption, recovery, permissions, channel changes, and assistive use
- **Explore alternatives before converging** — compare structurally different flows
  against the same user, accessibility, content, policy, technical, and operational
  criteria
- **Match fidelity to the question** — use the cheapest representation that can test
  structure, comprehension, navigation, interaction, or realistic behavior
- **Evaluate with representative users** — define tasks and observable success,
  capture breakdowns and workarounds, then revise the design or the premise
- **Hand off decisions and uncertainty** — preserve the tested artifact revision,
  rationale, state inventory, content and accessibility requirements, evidence,
  limitations, and implementation questions

## Gotchas

- Research findings inform design; they do not uniquely determine one solution.
- A design-system component improves consistency but does not prove journey usability.
- High fidelity can conceal untested information architecture and interaction logic.
- Prototype code is disposable evidence unless production engineering separately
  accepts and validates it.
- A design recommendation does not approve scope, implementation, or release.

## Example

```text
Need: A returning customer must recover an interrupted identity check.
Evidence: Research shows users cannot tell whether captured documents were retained.
Alternatives: Resume from last verified step; review-and-confirm; restart with warning.
Prototype question: Can users predict what is retained and resume without duplicate work?
Evidence: Five of six completed review-and-confirm; one missed the retained-data notice.
Decision: Continue with review-and-confirm after revising the notice and error recovery.
Handoff: Tested prototype revision, state map, content, accessibility criteria, and gaps.
```

## Progressive Disclosure

- Read [references/frame-and-model.md](references/frame-and-model.md) - Load when defining the journey, information architecture, interaction states, constraints, or design hypotheses
- Read [references/prototype-and-evaluate.md](references/prototype-and-evaluate.md) - Load when choosing prototype fidelity, planning evaluation, comparing alternatives, or interpreting design evidence
- Read [references/handoff-and-change.md](references/handoff-and-change.md) - Load when recording design rationale, preparing implementation handoff, or revising a design after evidence
