# Sources

Product-specific facts are pinned to the documentation snapshot below. The local runtime probe for this implementation returned `not-installed`; documentation conformance is not runtime conformance.

## OpenCode plugins

- **URL:** https://opencode.ai/docs/plugins/
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **Aspects extracted:** Global/project plugin locations, npm installation, load order, JavaScript/TypeScript API, events, tool interception, shell environment, custom tools, logging, and experimental compaction hooks.

## Refresh workflow

1. Re-fetch the official pages above and compare event, handler, configuration, trust, and lifecycle behavior.
2. Run the documented version probe and replace `not-installed` only with the observed version.
3. Execute every named deterministic probe on the observed surface; keep unsupported and experimental results explicit.
4. Update the matrix version and review date when behavior changes.
