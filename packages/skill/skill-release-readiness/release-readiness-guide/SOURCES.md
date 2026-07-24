# Sources

## Release engineering and canarying

- **URL:** https://sre.google/workbook/canarying-releases/
- **Last reviewed:** 2026-07-24
- **References:** `SKILL.md`, `references/rollout-and-recovery.md`,
  `references/decision-and-verification.md`
- **Aspects extracted:** Treat releases as controlled production changes, use partial
  time-bounded deployments, compare canary and control, evaluate explicit metrics, and
  proceed or stop stages based on evidence.

## Build provenance

- **URL:** https://slsa.dev/spec/v1.2/provenance
- **Last reviewed:** 2026-07-24
- **References:** `SKILL.md`, `references/candidate-and-evidence.md`
- **Aspects extracted:** Use verifiable information about where, when, and how an
  artifact was produced to bind release evidence and verification to artifact identity.

## Secure release practices

- **URL:** https://csrc.nist.gov/pubs/sp/800/218/final
- **Last reviewed:** 2026-07-24
- **References:** `references/candidate-and-evidence.md`,
  `references/decision-and-verification.md`
- **Aspects extracted:** Protect software and release integrity, verify release
  artifacts, preserve provenance, and identify and address residual vulnerabilities.

## Operational change readiness

- **URL:** https://docs.aws.amazon.com/wellarchitected/latest/operational-excellence-pillar/operational-readiness.html
- **Last reviewed:** 2026-07-24
- **References:** `references/rollout-and-recovery.md`
- **Aspects extracted:** Test and validate changes, make them small and reversible,
  plan for unsuccessful changes, use safe deployment strategies, and make informed
  deployment decisions with runbooks and support plans.
