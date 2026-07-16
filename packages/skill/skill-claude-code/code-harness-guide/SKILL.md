---
name: code-harness-guide
description: "Use when mapping agent-governance intents to Claude Code hooks, plugins, skills, settings, or managed configuration. Triggers on SessionStart, UserPromptSubmit, PreToolUse, PermissionRequest, PostToolUse, compaction, command/prompt/agent/HTTP/MCP-tool handlers, parallel hook behavior, allowManagedHooksOnly, or onboarding and testing Claude Code automation — even when the user doesn't say 'harness adapter'."
---

# Claude Code Harness Governance

Map the semantic contracts owned by **agent-governance-guide** to Claude Code's native configuration and extension surface without redefining policy or inventing a universal hook file.

## Essentials

- **Pin the surface** - Record the documentation snapshot and observed `claude --version` result before selecting a guarantee.
- **Map event semantics** - Translate governance intents through the versioned capability matrix, never by event-name similarity.
- **Preserve handler differences** - Command, HTTP, MCP-tool, prompt, and agent handlers have event-specific support and output behavior.
- **Assume parallel matching** - Make handlers reentrant and do not expect a denial to cancel sibling side effects.
- **Respect settings authority** - Managed, user, project, local, plugin, and component hooks have native precedence and trust rules.
- **Onboard transactionally** - Preview exact settings/plugin changes, permissions, data flow, verification, disable, and rollback before apply.

## Gotchas

- Agent handlers are experimental and cannot guarantee a mandatory control.
- Exit code `2` is the blocking signal for most blocking events; exit code `1` is normally a non-blocking error.
- There is no standalone `.claude/hooks.json`; standalone hooks live under the `hooks` key in a settings file.
- All matching handlers run in parallel, so a blocking result does not undo a sibling handler that already produced a side effect.
- Managed settings may restrict hook sources, but executable distribution and provenance remain separate.

## Example

A mandatory protected-path control selects `PreToolUse` with a deterministic command handler only after the observed version, covered tools, exit-2 denial, managed precedence, sibling concurrency, script provenance, and rollback probe all pass.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
