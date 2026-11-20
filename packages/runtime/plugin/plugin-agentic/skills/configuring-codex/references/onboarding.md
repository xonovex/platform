# Codex Onboarding

Follows the lifecycle owned by **configuring-harness-integrations**; this file holds the Codex specifics.

## Discover

- Run `codex --version`.
- Inspect active config layers, `~/.codex/hooks.json`, `~/.codex/config.toml`, repository `.codex` files, plugins, managed requirements, project trust, and `/hooks` review state.
- Inventory hook hashes, enabled/disabled state, covered tool names, command paths, timeouts, and managed script distribution.

## Preview

Show the exact JSON or TOML merge, source layer, command and Windows override, working-directory resolution, trust hash, and managed source.

Prefer one representation per layer. If a layer already contains both hooks.json and inline hooks, report the native warning and recommend consolidation without changing state automatically.

## Apply and verify

Apply through the selected native layer or pinned plugin. For non-managed commands, complete native hash review before claiming the hook can run. Use `/hooks` and startup diagnostics to verify source, trust, and enabled state. Probe covered allow and denial, an unsupported equivalent path, and stop continuation.

## Lifecycle

- Disable an individual non-managed hook through the native hook browser when available, or remove only its owned definition.
- Managed hooks change through requirements or the controlling managed layer; user configuration cannot disable them.
- Roll back both the config digest and separately delivered managed script digest.
- Add schema behavior, trust hash, coverage, feature flags, and the managed-only restriction to the drift dimensions.

Never use `--dangerously-bypass-hook-trust` as persistent onboarding. It is an explicit one-run bypass for already-vetted automation and remains outside the default recipe.
