---
name: pi-guide
description: "Use when configuring Pi extensions, packages, skills, settings, context injection, tool interception, permissions, or subagent patterns. Triggers on .pi/extensions, pi packages, project_trust, tool_call, tool_result, context, compaction, registerTool, full-system permissions, no built-in sandbox, package pinning, or Pi onboarding — even when the user doesn't say 'Pi extension'."
---

# Pi Extensions and Tool Interception

Configure Pi's native extension and interception surface without inventing unsupported guarantees or a universal hook file.

## Essentials

- **Treat extensions as full code** - They execute with the Pi process and user account permissions.
- **Resolve project trust first** - Project settings, packages, extensions, and skills wait for the native trust decision.
- **Map extension events** - Preserve sequential preflight, concurrent tool execution, middleware result order, and mutable inputs.
- **Separate skills from enforcement** - Skills add instructions and scripts but do not prove an extension blocked.
- **Own the sandbox boundary** - Pi has no built-in sandbox; isolation is external or extension/package responsibility.
- **Bound child processes** - Subagent patterns explicitly enforce depth, permissions, cwd trust, budgets, cancellation, and results.

## Gotchas

- Pi has no built-in sandbox; extensions and tools run with the user's process permissions.
- Project trust controls resource loading but does not constrain what trusted tools can do.
- `tool_call` preflight is sequential while sibling tools may execute concurrently.
- Mutated tool input is not revalidated automatically.
- Versioned package specifications are pinned and skipped by ordinary package updates; moving sources require a fresh trust decision.

## Example

A project extension that blocks destructive shell calls remains disabled until project trust and source review pass; its `tool_call` handler validates mutable input, returns a documented block, declares full process authority, and is tested with concurrent sibling tools and a rollback to the prior settings digest.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
