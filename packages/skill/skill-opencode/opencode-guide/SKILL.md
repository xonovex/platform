---
name: opencode-guide
description: "Use when configuring OpenCode JavaScript or TypeScript plugins, events, custom tools, or settings. Triggers on .opencode/plugins, opencode.json plugin packages, tool.execute.before/after, permission.asked, session events, file events, shell.env, custom tools, sequential plugin load order, Bun dependency install, plugin trust, or OpenCode onboarding, even when the user doesn't say 'OpenCode plugin'."
---

# OpenCode Plugins and Events

Configure OpenCode's native plugin and event surface without inventing unsupported guarantees or a universal hook file.

## Essentials

- **Treat plugins as executables** - Local and npm plugins run with OpenCode process authority.
- **Preserve load order** - Global config, project config, global plugins, then project plugins run sequentially.
- **Use tool interception precisely** - `tool.execute.before` can inspect/mutate or fail a call; other events may only observe.
- **Separate permission events** - Event availability does not itself prove decision control.
- **Review automatic installs** - Config dependencies and npm plugins may trigger Bun installation at startup.
- **Onboard through native scopes** - Preview config/plugin source, permissions, data flow, ordering, verification, disable, and rollback.

## Gotchas

- Plugins from all sources load, and all hooks run in documented sequence; similar local and npm plugins can both load.
- An npm dependency may be installed automatically at startup.
- A custom tool with the same name as a built-in tool takes precedence.
- `shell.env` affects AI shell tools and user terminals; injecting secrets broadens exposure.
- Compaction customization uses an experimental event and cannot provide a requested stable guarantee.

## Example

A secret-file control uses a reviewed project plugin with `tool.execute.before`; onboarding verifies the project source and dependencies, tests the thrown denial, records sequential load order and conflicting plugins, and rolls back by removing only the owned plugin and config dependency.

## Progressive Disclosure

- Read [references/capabilities.md](references/capabilities.md) - Load when mapping semantic intents to native events, handlers, guarantees, versions, ordering, context, or limitations
- Read [references/onboarding.md](references/onboarding.md) - Load when discovering, previewing, applying, verifying, disabling, rolling back, updating, or detecting drift
- Read [references/patterns.md](references/patterns.md) - Load when translating deterministic hooks, model evaluators, specialist agents, or managed adoption recipes
