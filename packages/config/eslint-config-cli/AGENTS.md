# ESLint Config CLI

- Same as `eslint-config-base` — `"import"` before `"node"` in exports
- Especially matters for `packages/script/` (`typescript-script` tag removes `^:build` deps to break circular cycles, so config may not be built at lint time)
