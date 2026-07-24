---
name: security-testing-guide
description: "Use when planning, conducting, reviewing, or reporting an authorized technical security assessment of software and its supporting environment. Triggers on security verification, penetration testing, SAST, DAST, dependency or secret scanning, abuse tests, security controls, ASVS, WSTG, assessment rules of engagement, vulnerability findings, exploitability, remediation verification, or security retest, even when the user doesn't say security testing."
---

# Security Testing

Produce authorized, risk-based, reproducible security evidence using complementary
techniques without treating a scanner, checklist, or one-time penetration test as
proof of security.

## Essentials

- **Establish authorization and safety** — pin owners, targets, revisions,
  environments, time window, allowed techniques, prohibited effects, data handling,
  stop conditions, contacts, and recovery
- **Derive coverage from risk** — connect threats, assets, trust boundaries, abuse
  cases, architecture, change impact, and versioned verification requirements
- **Combine techniques** — select manual review, code and configuration analysis,
  dependency and secret checks, control tests, dynamic testing, fuzzing, and
  penetration testing according to the failure sought
- **Control execution** — minimize data and service impact, validate tool findings,
  preserve exact versions and inputs, and stop when authorization or safety is unclear
- **Report reproducibly** — describe affected subject, preconditions, steps, observed
  result, impact, likelihood or exposure, evidence, confidence, and safe remediation
- **Verify remediation** — retest the exact fix, check bypass and regression paths,
  update residual risk, and keep accountable acceptance external

## Gotchas

- Never test an external or production target without explicit scope and authorization.
- A clean automated scan has unknown coverage and false-negative risk.
- A proof of concept should demonstrate the issue with the least harmful action and
  data necessary.
- Severity must reflect the actual asset, exposure, controls, and business impact.
- Threat modeling predicts what to verify; security testing supplies evidence.

## Example

```text
Authorization: Checkout staging build rc-18, 09:00-12:00 UTC, no destructive payloads.
Basis: Threat TM-14 and ASVS v5.0.0 access-control requirements.
Techniques: Manual authorization matrix review, API negative tests, and code-path review.
Finding: A support role can read another tenant's draft order by identifier.
Evidence: Reproduced on rc-18 with synthetic tenants; no production data accessed.
Retest: Verify fixed object-level authorization plus adjacent list and mutation paths.
```

## Progressive Disclosure

- Read [references/plan-and-authorize.md](references/plan-and-authorize.md) - Load when defining scope, authorization, rules of engagement, verification requirements, safety, data handling, or stop conditions
- Read [references/techniques-and-execution.md](references/techniques-and-execution.md) - Load when selecting complementary security techniques, validating tools, executing tests, or controlling assessment impact
- Read [references/findings-and-retest.md](references/findings-and-retest.md) - Load when triaging results, writing reproducible findings, recommending remediation, retesting, or reporting residual risk
