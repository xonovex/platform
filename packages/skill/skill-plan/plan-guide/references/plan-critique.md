# plan-critique: Independently Stress-Test a Plan

Review an inline plan or an exact provider-native plan reference without changing it. Use the optional native revision to pin mutable provider state, run in fresh context, and return separate findings inline.

## Lenses

- **Red-team** — attack essential assumptions and unproven dependencies.
- **Pre-mortem** — assume delivery failed and trace likely causes to planning choices.
- **Falsify** — find concrete inputs, scales, sequences, environments, or recovery cases the plan misses.
- **Steelman** — identify the smallest change that would strengthen a weak section.

## Core Workflow

1. Resolve the explicit plan, supporting references, review criteria, and optional native revision. Require an exact revision when the selected provider otherwise exposes mutable state.
2. Start fresh independent context that did not author the plan. Reconstruct provider-native inputs instead of relying on conversation memory.
3. Ground material claims in source evidence and current environment facts.
4. Apply the selected lenses and report findings with severity, confidence, affected section, failure mode, evidence, and suggested direction.
5. Return the critique as a separate inline result. Use a later Publish operation if it must be persisted.

## Gotchas

- Critique never revises the subject or changes its descriptive status.
- A passing critique is not validation and does not authorize implementation.
- Vague findings without a failure mode or supporting evidence are not actionable.
