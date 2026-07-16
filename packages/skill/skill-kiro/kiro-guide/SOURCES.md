# Sources

Product-specific facts are pinned to the documentation snapshot below. The local runtime probe for this implementation returned `not-installed`; documentation conformance is not runtime conformance.

## Kiro hooks

- **URLs:**
  - https://kiro.dev/docs/hooks/
  - https://kiro.dev/docs/hooks/types/
  - https://kiro.dev/docs/hooks/actions/
  - https://kiro.dev/docs/hooks/management/
  - https://kiro.dev/docs/hooks/best-practices/
  - https://kiro.dev/docs/hooks/troubleshooting/
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **Aspects extracted:** v1 JSON format, triggers, command and agent actions, blocking exit codes, workspace scope, generated-hook review, testing, diagnostics, enable/disable, and deletion.

## Kiro CLI v3 and workspaces

- **URLs:**
  - https://kiro.dev/docs/cli/v3/
  - https://kiro.dev/docs/editor/multi-root-workspaces/
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`
- **Aspects extracted:** CLI v3 standalone hook alignment, diagnostics, early-access gaps, and multi-root ownership and file-event scope.

## Refresh workflow

1. Re-fetch the official pages above and compare event, handler, configuration, trust, and lifecycle behavior.
2. Run the documented version probe and replace `not-installed` only with the observed version.
3. Execute every named deterministic probe on the observed surface; keep unsupported and experimental results explicit.
4. Update the matrix version and review date when behavior changes.
