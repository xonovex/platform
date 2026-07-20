# Sources

## Repository runtime contract

- **Source:** Executable schemas and tests in `scripts/`
- **Provenance:** Repository-local executable contract maintained with this skill
- **Last reviewed:** 2026-07-20
- **Used for:** `SKILL.md`, `references/architecture.md`, `references/composition.md`, `references/controls.md`, `references/adapters.md`, `references/maturity.md`
- **Aspects extracted:** Open trigger vocabulary, plugin ports, explicit observe/enforce modes, explicit evidence failure behavior, capability requirements, composition explanation, command adapters, and caller-defined maturity assessment

## Refresh workflow

1. Run the focused runtime tests.
2. Compare every documented enforcement claim with `workflow-runtime.ts`.
3. Compare adapter examples with `workflow-command-runtime.ts` and `workflow-trigger-adapters.ts`.
4. Remove claims that are not executable or explicitly identified as deployment-owned.
