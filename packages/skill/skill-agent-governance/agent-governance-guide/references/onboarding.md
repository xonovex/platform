# Advisory and Transactional Onboarding

## Lifecycle

1. **Discover** — identify actual products, versions, editions, scopes, existing controls, trust state, and available native mechanisms.
2. **Assess** — compare requirements with tested capabilities, gaps, conflicts, risks, permissions, and data flows.
3. **Recommend** — propose a dependency-valid composition and distinguish advisory, evidence, enforcing, configuration-changing, and privileged modules.
4. **Preview** — show exact native changes, versions, permissions, secrets, network/data access, ordering, failure behavior, expected evidence, verification, and rollback.
5. **Approve** — obtain informed user consent or the required managed authorization.
6. **Apply** — use native adapters idempotently; do not hide mutations inside prose or an opaque agent task.
7. **Verify** — dry-run and execute conformance probes against expected native behavior.
8. **Record** — publish provider-native configuration and evidence references.
9. **Operate** — detect drift, upgrade, disable, roll back, remove, and retire safely.

## Result semantics

Each step has an independent result containing subject/scope, actor or executor, native references, versions, evidence, outcome, limitations, and follow-up capabilities. Preview and approval are distinct; apply never substitutes for verify; rollback evidence remains available after removal.

## Advisor rules

- Prefer deterministic discovery; use a bounded model or specialist agent only for genuine interpretation or adaptive investigation.
- Never mutate before authorization.
- Never claim installed means enforced.
- Fail clearly when a requested feature is unsupported, experimental, tier-restricted, stale, or conflicts with a mandatory control.
- Preserve product edition, account/tier, tested version, native identifiers, credentials/data flows, and source-of-truth boundaries.
