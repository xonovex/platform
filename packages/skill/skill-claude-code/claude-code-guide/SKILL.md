---
name: claude-code-guide
description: "Use when configuring Claude Code hooks, plugins, skills, settings, or managed configuration. Triggers on SessionStart, UserPromptSubmit, PreToolUse, PermissionRequest, PostToolUse, compaction, command/prompt/agent/HTTP/MCP-tool handlers, parallel hook behavior, allowManagedHooksOnly, or onboarding and testing Claude Code automation — even when the user doesn't say 'harness adapter'."
---

# Claude Code Hooks

Configure Claude Code's native hook and extension surface without inventing unsupported guarantees or a universal hook format.

## Essentials

- **Pin the surface** - Record the documentation snapshot and observed `claude --version` result before selecting a guarantee.
- **Map event semantics** - Translate requested behavior through the versioned capability matrix, never by event-name similarity.
- **Preserve handler differences** - Command, HTTP, MCP-tool, prompt, and agent handlers have event-specific support and output behavior.
- **Assume parallel matching** - Make handlers reentrant and do not expect a denial to cancel sibling side effects.
- **Respect settings authority** - Managed, user, project, local, plugin, and component hooks have native precedence and trust rules.
- **Onboard transactionally** - Preview exact settings/plugin changes, permissions, data flow, verification, disable, and rollback before apply.
- **Keep hooks focused** - Give each handler one native event contract and one explicit responsibility.

## Gotchas

- Agent handlers are experimental; do not select them when the requested control requires a verified blocking guarantee.
- Exit code `2` is the blocking signal for most blocking events; exit code `1` is normally a non-blocking error.
- There is no standalone `.claude/hooks.json`; standalone hooks live under the `hooks` key in a settings file.
- All matching handlers run in parallel, so a blocking result does not undo a sibling handler that already produced a side effect.
- Managed settings may restrict hook sources, but executable distribution and provenance remain separate.

## Example

A project can use a deterministic `PreToolUse` command to inspect covered tool calls and return Claude Code's documented blocking response when its explicit rule rejects one.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
