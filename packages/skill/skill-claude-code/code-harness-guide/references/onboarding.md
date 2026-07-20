# Claude Code Transactional Onboarding

## Discover and diagnose

- Run `claude --version`; record absence as `not-installed`.
- Inspect `/status`, `/hooks`, `/permissions`, managed settings origin, `~/.claude/settings.json`, project settings, local settings, enabled plugins, and component-scoped hooks.
- Resolve each requested behavior and guarantee against the capability matrix before recommending a handler.
- Review every command, URL, MCP tool, prompt, agent definition, environment variable, transcript/path input, and executable source.

## Preview

Show the exact merge into the selected settings `hooks` object or plugin `hooks/hooks.json`. Include source scope, handler type, event/matcher, command arguments, URL/env allowlists, model/provider use, permissions, files, network, secrets, content exposure, timeout, parallel behavior, expected evidence, verification, disable, and rollback.

Never overwrite the whole settings file. Preserve unrelated hook groups and array entries. A plugin preview includes the pinned plugin source/version and the executable digest separately from settings.

## Apply and verify

Apply only after explicit authorization for the exact diff and authority request. Use the native settings or plugin mechanism, restart or reload where required, then verify source attribution in `/hooks` and active managed origin in `/status`. Run deterministic probes for allow, deny, context, non-blocking error, parallel siblings, and timeout.

## Disable, rollback, update, and drift

- Disable all hooks only when the scope-wide effect is intended; there is no native per-handler disable switch in a settings block.
- Remove only the owned handler or disable/uninstall the owned plugin, then verify it disappeared from `/hooks`.
- Roll back to the recorded settings/plugin digest and rerun probes.
- Treat handler type, managed allowlists, source precedence, command digest, permissions, URL/data flow, and observed product version as drift dimensions.

Project and plugin executables require trust and review. Organization-managed hook configuration does not prove its referenced script was securely distributed.
