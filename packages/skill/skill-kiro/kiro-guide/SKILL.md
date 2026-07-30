---
name: kiro-guide
description: "Use when configuring Kiro IDE or CLI v3 hooks, command actions, or agent actions. Triggers on .kiro/hooks JSON, SessionStart, UserPromptSubmit, PreToolUse, PreTaskExec, file events, exit-code blocking, workspace roots, generated-hook review, hook testing, enable/disable, or Kiro onboarding, even when the user doesn't say 'Kiro hook'."
---

# Kiro Hooks

Configure Kiro's native hook and action surface without inventing unsupported guarantees or a universal hook file.

## Essentials

- **Pin schema and surface** - Record v1 hook schema, IDE/CLI surface, documentation snapshot, and observed runtime.
- **Prefer command actions** - Use deterministic commands for policy, formatting, validation, and audit.
- **Bound agent actions** - Treat Ask Kiro actions as model execution with credits, authority, and output limits.
- **Use exit code 2 precisely** - Blocking applies only to documented pre-events.
- **Respect workspace roots** - File events and configuration are scoped to the defining workspace root.
- **Review generated hooks** - Natural-language generation is a draft requiring diff, security review, tests, and authorization.

## Gotchas

- Command exit code `2` blocks only `PreToolUse`, `UserPromptSubmit`, and `PreTaskExec` in the v1 reference.
- Other non-zero exits warn and normally continue.
- Agent actions start another agent loop and consume credits; they are not deterministic commands.
- Setting timeout `0` disables the command timeout and violates bounded-module defaults.
- Generated hook configurations must be reviewed before saving or sharing.

## Example

A pre-tool protected-path hook uses a v1 workspace JSON file and a bounded command action; its preview includes the matcher, exit-2 denial, command timeout, workspace trust, stdout/stderr exposure, enable flag, test cases, and removal rollback.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
