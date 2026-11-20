# OpenCode Onboarding

Applies the lifecycle of **configuring-harness-integrations** to OpenCode plugins.

## Discover

- Probe `opencode --version`, global/project configuration, global/project plugin directories, npm plugin list, config-directory package dependencies, Bun cache, custom tools, and effective load order.
- Inspect every plugin source, import, package version, install script risk, event handler, custom tool schema, command execution, filesystem/network/secret/model/data access, and name collision.
- Detect a local and npm plugin with similar ownership as a conflict even when names differ.

## Preview

Show the exact `opencode.json` merge, local plugin file, or pinned package change. Include source/version/digest, scope and load position, Bun install behavior, event keys, custom tool names, argument mutations, failure behavior, and timeout/bounds supplied by the module.

Do not inject secrets through `shell.env` by default. Prefer narrow runtime retrieval and declare whether AI tools, user terminals, logs, or provider requests can observe the value.

## Apply and verify

Review the source first, then verify load order and duplicates. Probe tool mutation, sequential handler order, post-tool observation, permission events, file events, shell environment minimization, custom tool schema/cancellation/output bounds, and experimental compaction behavior.

## Lifecycle

- Disable by removing only the owned config entry or moving/removing the owned local plugin through a recorded native change.
- Roll back config, plugin, dependency manifest, and cached/pinned package reference as applicable, then verify the handler and custom tools disappear.
- Drift includes the dependency graph, load order, duplicate modules, custom-tool collisions, and shell environment.
