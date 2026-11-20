# Claude Code Transactional Onboarding

Follows the lifecycle owned by **configuring-harness-integrations**; this file holds the Claude Code specifics.

## Discover

- Run `claude --version`.
- Inspect `/status`, `/hooks`, `/permissions`, managed settings origin, `~/.claude/settings.json`, project settings, local settings, enabled plugins, and component-scoped hooks.

## Preview

Show the exact merge into the selected settings `hooks` object or plugin `hooks/hooks.json`, including URL and environment allowlists and model/provider use. Preserve unrelated hook groups and array entries. A plugin preview includes the executable digest separately from settings.

## Apply and verify

Restart or reload where required, then verify source attribution in `/hooks` and the active managed origin in `/status`. Probe a non-blocking error beside the common probes.

## Lifecycle

- Disable all hooks only when the scope-wide effect is intended; there is no native per-handler disable switch in a settings block.
- Remove only the owned handler or disable/uninstall the owned plugin, then verify it disappeared from `/hooks`.
- Roll back to the recorded settings/plugin digest.
- Add handler type, managed allowlists, and URL flow to the drift dimensions.

Project and plugin executables require trust and review. Organization-managed hook configuration does not prove its referenced script was securely distributed.
