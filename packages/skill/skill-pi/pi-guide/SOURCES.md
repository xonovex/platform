# Sources

Product-specific facts are pinned to the documentation snapshot below. The local runtime probe for this implementation returned `not-installed`; documentation conformance is not runtime conformance.

## Pi extensions

- **URL:** https://pi.dev/docs/latest/extensions
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **Aspects extracted:** Events, event ordering, tool blocking and mutation, result middleware, context injection, compaction, custom tools, output limits, session behavior, and full-process extension authority.

## Pi packages, settings, skills, and security

- **URLs:**
  - https://pi.dev/docs/latest/packages
  - https://pi.dev/docs/latest/settings
  - https://pi.dev/docs/latest/skills
  - https://pi.dev/docs/latest/security
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **Aspects extracted:** Package sources and pins, install/update/remove, user/project scopes, filters, project trust, skill discovery, full system access, and absence of a built-in sandbox.

## Refresh workflow

1. Re-fetch the official pages above and compare event, handler, configuration, trust, and lifecycle behavior.
2. Run the documented version probe and replace `not-installed` only with the observed version.
3. Execute every named deterministic probe on the observed surface; keep unsupported and experimental results explicit.
4. Update the matrix version and review date when behavior changes.
