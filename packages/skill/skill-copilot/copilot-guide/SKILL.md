---
name: copilot-guide
description: "Use when configuring GitHub Copilot CLI or cloud-agent hooks, policy hooks, plugins, or skills. Triggers on .github/hooks, ~/.copilot/hooks, policy.d, preToolUse, permissionRequest, agentStop, command/HTTP/prompt handlers, cloud sandbox differences, plugin.json, repository trust, disableAllHooks, or Copilot onboarding — even when the user doesn't say 'Copilot hook'."
---

# GitHub Copilot Hooks and Policy

Configure GitHub Copilot's native hook and extension surface without inventing unsupported guarantees or a universal hook file.

## Essentials

- **Separate CLI and cloud** - Resolve the product surface before selecting an event or guarantee.
- **Use version-1 config** - Pin configuration schema, docs snapshot, and observed `copilot --version`.
- **Preserve handler support** - Commands cover hook types; HTTP and prompt handlers have narrower rules and surface limits.
- **Respect policy precedence** - CLI policy hooks load first and cannot be disabled by lower scopes.
- **Review repository execution** - Hooks, plugin extensions, and skill scripts require trust and security review.
- **Onboard reversibly** - Preview scope, handler, cloud data flow, permissions, firewall, verification, disable, and rollback.

## Gotchas

- Current documentation is not command-only: HTTP handlers are documented, and prompt handlers are limited to `sessionStart`.
- Cloud agent supports a subset of events, runs non-interactively, and does not use `permissionRequest` as a decision gate.
- Cloud hook files are repository-scoped and execute in an ephemeral Linux sandbox with restricted network.
- CLI policy hooks cannot be disabled with lower-scope `disableAllHooks`.
- Pre-approving shell in a skill expands risk and requires review of the skill and its scripts.

## Example

A policy requiring cloud-agent command blocking selects repository `preToolUse`, not `permissionRequest`; the preview includes version-1 JSON, bash-only cloud behavior, firewall/data flow, repository trust, output decision schema, job-timeout impact, and file-level disable rollback.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
