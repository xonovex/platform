---
name: codex-guide
description: "Use when configuring Codex hooks, plugins, skills, config layers, or managed requirements. Triggers on hooks.json, config.toml, requirements.toml, SessionStart, PreToolUse, PermissionRequest, SubagentStart, Stop, plugin-bundled hooks, hook trust hashes, concurrent commands, skipped prompt/agent handlers, or Codex guardrail coverage — even when the user doesn't say 'Codex hook'."
---

# Codex Hooks and Guardrails

Configure Codex's native hooks and guardrails without inventing unsupported guarantees or a universal hook file.

## Essentials

- **Probe the release** - Record `codex --version` or an explicit not-installed result with the matrix snapshot.
- **Use command handlers** - Only documented executing handler types can satisfy a capability.
- **Reject parsed-only support** - Prompt, agent, and asynchronous command handlers are not executing release behavior in this snapshot.
- **Scope the guardrail** - `PreToolUse` covers documented shell, patch, and MCP paths, not every equivalent operation.
- **Preserve trust layers** - Non-managed definitions are hash-reviewed; managed hooks require separately distributed scripts.
- **Transact native config** - Preview and verify hooks.json, config.toml, plugin, or requirements changes without replacing unrelated state.

## Gotchas

- Only `type: command` executes in the documented snapshot; prompt and agent handlers are parsed but skipped.
- The async option is parsed but asynchronous command hooks are skipped.
- `PreToolUse` is a partial guardrail because not every shell or non-shell path is intercepted.
- Matching command hooks start concurrently, so a denial cannot prevent another matching hook from starting.
- `SubagentStart` can add context, but `continue: false` does not stop the subagent from starting.

## Example

Codex `PreToolUse` alone cannot block every privileged operation because its documented coverage is partial. Use it only for covered Bash, patch, and MCP calls, and state the uncovered paths explicitly.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
