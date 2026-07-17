---
name: security-assurance-guide
description: "Use when planning, reviewing, or operating security assurance across risk, controls, secure development, verification, software supply chain, privileged changes, exceptions, incidents, and evidence. Triggers on threat models, NIST CSF or SP 800-53 profiles, SSDF, control selection and coverage, SLSA or artifact provenance, signatures and attestations, security release gates, scanner outages, compensating controls, emergency exceptions, or security claims, even when the user doesn't say 'security assurance'."
---

# Security Assurance

Build justified confidence in security outcomes without equating a checklist, scanner, signature, framework mapping, or passing control with compliance or absence of risk.

## Essentials

- **Scope risk before controls** — define exact subjects, assets, actors, trust boundaries, threats, impacts, assumptions, applicability, and risk owners, see [references/risk-and-controls.md](references/risk-and-controls.md)
- **Select outcome-based controls** — bind preventive, detective, responsive, and recovery controls to owners, enforcement, failure behavior, evidence, gaps, and residual risk, see [references/risk-and-controls.md](references/risk-and-controls.md)
- **Integrate secure development** — govern requirements, design, implementation, review, testing, vulnerabilities, changes, release, and retirement, see [references/secure-development.md](references/secure-development.md)
- **Verify the supply chain** — inventory and verify source, dependencies, builders, artifacts, modules, provenance, signatures, permissions, updates, and rollback, see [references/supply-chain.md](references/supply-chain.md)
- **Assure exact revisions** — layer deterministic evidence, tests, scanners, independent review, intended-denial probes, and freshness checks, see [references/verification-and-operations.md](references/verification-and-operations.md)
- **Operate the control system** — handle outages, exceptions, incidents, drift, canaries, emergency disable, corrective action, and retirement without silent weakening, see [references/verification-and-operations.md](references/verification-and-operations.md)

## Gotchas

- A control catalog is a source of selectable controls, not a universal baseline or automatic applicability decision.
- A signature proves only what the configured verifier establishes about subject, digest, identity, and transparency/provenance claims. It does not prove the artifact is secure.
- A scanner's absence of findings is bounded by its rules, configuration, reachability, exact subject, version, and time.
- An installed hook or policy is not enforcement evidence. Mandatory controls need an adequate protected enforcement point and explicit failure behavior.
- Crosswalks are contextual and often many-to-many; never represent them as equivalence, certification, or legal compliance.

## Example

```text
Subject: release-artifact@sha256:abc · threat model v7 · security profile v4
Controls: least privilege + protected build + provenance verification + vulnerability gate
Evidence: source/build/dependency identities + signature policy result + denied-action probes
Gaps: scanner outage; release fails closed unless independent fresh evidence or scoped exception
Operations: canary policy update · verified rollback · incident owner · retained audit references
```

## Progressive Disclosure

- Read [references/risk-and-controls.md](references/risk-and-controls.md) - Load when threat modelling, selecting or tailoring controls, assigning enforcement and evidence, mapping frameworks, or accepting residual risk
- Read [references/secure-development.md](references/secure-development.md) - Load when integrating security requirements, design review, implementation, verification, vulnerability handling, release, or retirement into development
- Read [references/supply-chain.md](references/supply-chain.md) - Load when inventorying or verifying dependencies, artifacts, builders, signatures, attestations, executable modules, updates, provenance, or trust
- Read [references/verification-and-operations.md](references/verification-and-operations.md) - Load when designing assurance evidence, release gates, negative tests, scanner outages, exceptions, emergency exceptions, incidents, drift, rollback, or corrective action
