# Claude Code Sources

Product-specific facts are pinned to the documentation snapshot below. The local runtime probe for this implementation returned `not-installed`; documentation conformance is not runtime conformance.

## Hooks and hook workflows

- **URLs:**
  - https://code.claude.com/docs/en/hooks
  - https://code.claude.com/docs/en/hooks-guide
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **Aspects extracted:** Events, matchers, command/prompt/agent/HTTP/MCP-tool handlers, output and exit behavior, context injection, parallel matching, diagnostics, and experimental agent hooks.

## Settings and managed configuration

- **URLs:**
  - https://code.claude.com/docs/en/configuration
  - https://code.claude.com/docs/en/server-managed-settings
  - https://code.claude.com/docs/en/debug-your-config
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`
- **Aspects extracted:** Settings scopes and precedence, managed hook restrictions, source diagnostics, security approval, fail-closed startup option, and configuration troubleshooting.

## Plugin packaging

- **URLs:**
  - https://code.claude.com/docs/en/plugins
  - https://code.claude.com/docs/en/plugins-reference
- **Last reviewed:** 2026-07-16
- **Used for:** `onboarding.md`, `patterns.md`
- **Aspects extracted:** Plugin hook layout, root/data placeholders, copied installation cache, executable packaging, and testing.

## Refresh workflow

1. Re-fetch the official pages above and compare event, handler, configuration, trust, and lifecycle behavior.
2. Run the documented version probe and replace `not-installed` only with the observed version.
3. Execute every named deterministic probe on the observed surface; keep unsupported and experimental results explicit.
4. Update the matrix version and review date when behavior changes.
