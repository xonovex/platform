# Codex Onboarding

## Discover

- Run `codex --version`; preserve `not-installed` when absent.
- Inspect active config layers, `~/.codex/hooks.json`, `~/.codex/config.toml`, repository `.codex` files, plugins, managed requirements, project trust, and `/hooks` review state.
- Inventory hook hashes, enabled/disabled state, covered tool names, command paths, timeouts, and managed script distribution.
- Resolve each requested behavior and guarantee against coverage and executing handler types.

## Preview

Show the exact JSON or TOML merge, source layer, event/matcher, command and Windows override, working-directory resolution, timeout, permissions, files, network, secrets, transcript/data access, concurrency, evidence, trust hash, managed source, verification, disable, and rollback.

Prefer one representation per layer. If a layer already contains both hooks.json and inline hooks, report the native warning and recommend consolidation without changing state automatically.

## Apply and verify

After explicit authorization, apply idempotently through the selected native layer or pinned plugin. For non-managed commands, complete native hash review before claiming the hook can run. Use `/hooks` and startup diagnostics to verify source, trust, and enabled state. Probe a covered allow, covered denial, unsupported equivalent path, context output, concurrent siblings, timeout, and stop continuation.

## Lifecycle

- Disable an individual non-managed hook through the native hook browser when available, or remove only its owned definition.
- Managed hooks change through requirements or the controlling managed layer; user configuration cannot disable them.
- Roll back both the config digest and separately delivered managed script digest.
- Drift includes release version, schema behavior, trust hash, command digest, source layer, coverage, feature flags, managed-only restriction, and permissions/data flow.

Never use `--dangerously-bypass-hook-trust` as persistent onboarding. It is an explicit one-run bypass for already-vetted automation and remains outside the default recipe.
