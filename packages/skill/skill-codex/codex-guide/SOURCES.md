# Sources

Product-specific facts are pinned to the documentation snapshot below. The local runtime probe for this implementation returned `not-installed`; documentation conformance is not runtime conformance.

## Codex hooks

- **URL:** https://learn.chatgpt.com/docs/hooks
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **Aspects extracted:** Hook locations, config shapes, trust review, managed requirements, plugin hooks, concurrency, executing command handlers, skipped prompt/agent/async handlers, events, output schemas, and guardrail coverage.

## Codex plugins and skills

- **URLs:**
  - https://learn.chatgpt.com/docs/plugins
  - https://learn.chatgpt.com/docs/skills
- **Last reviewed:** 2026-07-16
- **Used for:** `onboarding.md`, `patterns.md`
- **Aspects extracted:** Plugin installation, bundled skills and hooks, new-session activation, and the distinction between instructions and enforcement.

## Codex managed configuration

- **URL:** https://learn.chatgpt.com/docs/managed-configuration
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`
- **Aspects extracted:** Managed configuration and organizational control layers; native behavior remains subject to the hook release reference and runtime probes.

## Refresh workflow

1. Re-fetch the official pages above and compare event, handler, configuration, trust, and lifecycle behavior.
2. Run the documented version probe and replace `not-installed` only with the observed version.
3. Execute every named deterministic probe on the observed surface; keep unsupported and experimental results explicit.
4. Update the matrix version and review date when behavior changes.
