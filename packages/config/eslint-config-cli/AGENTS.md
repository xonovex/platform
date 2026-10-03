# ESLint Config CLI

Resolve ESLint configuration from source before built output, especially when script tasks run before configuration builds.

- Same as `eslint-config-base` — `"import"` before `"node"` in exports
- Especially matters for `packages/tooling/script/` (`typescript-script` tag removes `^:build` deps to break circular cycles, so config may not be built at lint time)
