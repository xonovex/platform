---
name: assuring-accessibility
description: "Use when making a website, app, document or service accessible, or checking, gating or claiming its accessibility conformance. Triggers on accessibility, a11y, WCAG version or level, success criteria, inclusive design, keyboard or screen-reader testing, assistive technology, accessibility evidence, release gates, conformance claims, exceptions, or remediation, even when the user doesn't say 'accessibility assurance'."
---

# Accessibility Assurance

Make accessibility a versioned, testable quality and governance concern without requiring one platform, scanner, or conformance claim.

## Essentials

- **Pin the target**: record the standard, version, level, applicable criteria, exact subject revision, environments, and user journeys, see [references/assessment-and-evidence.md](references/assessment-and-evidence.md)
- **Design with affected users**: include perceivable, operable, understandable, and robust outcomes from discovery through content and interaction design, see [references/design-and-implementation.md](references/design-and-implementation.md)
- **Layer the evidence**: combine deterministic inspection, automated checks, keyboard and zoom testing, representative assistive technology, and qualified human review, see [references/assessment-and-evidence.md](references/assessment-and-evidence.md)
- **Publish criterion outcomes**: report pass, fail, not-applicable, and not-tested with origins, limitations, findings, gaps, freshness, and native references, see [references/assessment-and-evidence.md](references/assessment-and-evidence.md)
- **Govern exceptions and drift**: scope, own, expire, review, remediate, reassess, and retire accessibility decisions, see [references/governance-and-operations.md](references/governance-and-operations.md)
- **Keep tools in their lane**: platform skills own implementation APIs; scanners and agents provide evidence but never accountable conformance by themselves

## Gotchas

- An automated scanner covers only what it can observe; a clean report is not proof that every applicable criterion passes.
- `not-applicable` needs a criterion-specific reason and reviewer; it is not a synonym for untested or inconvenient.
- Conformance applies to the declared subject, scope, version, level, and dependencies. A changed UI, content path, component, platform, or assistive-technology environment may invalidate evidence.
- An exception preserves a known gap and remediation duty. It never changes a failed criterion into a pass or establishes conformance.
- Crosswalks between accessibility standards, laws, and design guidance are contextual; they do not establish equivalence, certification, or legal compliance.

## Example

```text
Target: checkout@sha256:abc · WCAG 2.2 AA · authenticated web journey
Evidence: deterministic markup checks + scanner@4.2 + keyboard/zoom + two AT/browser pairs
Outcome: 43 pass · 1 fail · 6 not applicable · 2 not tested
Decision: release blocked by profile; finding owned; scoped alternative access expires in 14 days
Reassess: exact fixed revision plus regression of adjacent journeys
```

## Progressive Disclosure

- Read [references/design-and-implementation.md](references/design-and-implementation.md) - Load when defining inclusive requirements, user journeys, content, interaction, platform ownership, or implementation handoffs
- Read [references/assessment-and-evidence.md](references/assessment-and-evidence.md) - Load when selecting applicable criteria, planning tests, assessing an exact revision, grading outcomes, or publishing accessibility evidence
- Read [references/governance-and-operations.md](references/governance-and-operations.md) - Load when composing profiles, release gates, exceptions, remediation, monitoring, drift, incident response, or retirement
