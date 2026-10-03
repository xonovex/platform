# publish-check: Validate Publish-Readiness

Check a package's `package.json` before publishing, separating npm's hard rules from house policy.

## Workflow

1. **Read** `package.json`: abort (error) on a parse error.
2. **Skip** when `private` is `true`: npm refuses to publish it; this is not an error.
3. **Hard rules (npm, block):** require `name` and `version`; `version` must be valid semver; `name` ≤ 214 chars, lowercase, URL/CLI/folder-safe.
4. **Policy checks (label them, don't call them npm rules):** presence of `license`, `repository`, `files`. npm only _warns_ on a missing `license` and does not require the others.
5. **`repository`**: accept the object `{type, url[, directory]}` **or** a shorthand string (`github:org/repo`, `org/repo`); don't force `type` when the url is a host shorthand.
6. **`access`**: only meaningful for `@scoped` names; unscoped packages are always public and cannot be restricted. Require explicit access (via `publishConfig.access` or `--access`) only for scoped packages.
7. **`files` existence**: warn on a literal path that doesn't exist, but **skip glob entries** (`*`, `**/*`), build outputs created later (e.g. `dist/`), and the always-included `package.json` / `README` / `LICENSE` / `main`.
8. **Report** errors (block) and warnings (continue); exit non-zero on any error.

Not-ready output:

```
@scope/pkg is not ready for publishing:
  - Missing required field: license
  Warning: file "dist/index.js" does not exist yet (built later)
```

## Gotchas

- `publishConfig.access: "restricted"` (or its alias `"private"`) is **not** the same as top-level `private: true`; only the latter blocks publishing.
- Requiring `license` / `repository` / `files` is stricter than npm: keep it if you like, but call it policy so it isn't mistaken for an npm rule.
