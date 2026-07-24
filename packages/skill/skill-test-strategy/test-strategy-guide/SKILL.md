---
name: test-strategy-guide
description: "Use when deciding what must be tested, at which levels, with which test types and techniques, and what evidence is sufficient for a product change or release. Triggers on test strategy, test approach, quality risks, coverage model, test levels, test types, entry or exit criteria, environments, test data, automation allocation, defect policy, or residual quality risk, even when the user doesn't say test strategy."
---

# Test Strategy

Design a proportionate, risk-based system of test activities and evidence without
turning a document template, coverage percentage, or test pyramid into the objective.

## Essentials

- **Pin scope and quality goals** — identify the product or change revision, users,
  critical journeys, architecture, environments, constraints, and acceptance basis
- **Model product risks** — connect failure modes to likelihood or exposure,
  consequence, detectability, affected users, and existing controls
- **Select complementary coverage** — allocate checks across static, component,
  integration, contract, system, acceptance, exploratory, accessibility, performance,
  resilience, compatibility, security, recovery, and production verification as needed
- **Choose techniques deliberately** — derive tests from examples, state transitions,
  boundaries, decisions, data combinations, journeys, models, threats, and operational
  scenarios
- **Design trustworthy execution** — define environments, test data, observability,
  oracles, independence, automation, flake policy, defect handling, and evidence
- **Set evidence-based completion** — use risk coverage, result quality, unresolved
  defects, limitations, and residual risk; never infer readiness from test counts alone

## Gotchas

- Risk-based testing prioritizes effort; it does not excuse unexamined high-impact gaps.
- Code coverage reveals execution, not assertion quality or product-risk coverage.
- A lower-level test is preferable only when it can detect the same failure with
  sufficient realism and confidence.
- Automation is valuable for repeatability and feedback, not as a target percentage.
- Test strategy informs quality evidence; release readiness makes the candidate-level
  recommendation.

## Example

```text
Subject: Checkout release candidate rc-18.
Risk: Provider timeout can duplicate a charge after client retry; impact is critical.
Coverage: Component idempotency properties, provider contract tests, system-level
          timeout injection, exploratory recovery charter, and production canary signal.
Environment: Provider sandbox plus deterministic fault proxy; production data excluded.
Completion: All critical scenarios pass at rc-18, no unresolved critical defects,
            flaky results investigated, and sandbox limitation recorded as residual risk.
```

## Progressive Disclosure

- Read [references/scope-and-risks.md](references/scope-and-risks.md) - Load when defining the test basis, scope, quality goals, product risks, priorities, or traceability
- Read [references/coverage-and-techniques.md](references/coverage-and-techniques.md) - Load when selecting test levels, types, techniques, automation, environments, data, or specialist capabilities
- Read [references/evidence-and-completion.md](references/evidence-and-completion.md) - Load when defining entry or exit criteria, evaluating results, handling defects and flakes, or reporting residual risk
