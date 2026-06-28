# publish-check: Validate Publish-Readiness

Check a package's `package.json` before publishing, separating npm's hard rules from house policy.

## Goal

- Fail fast on anything that would make `npm publish` reject the package or ship a broken artifact.
- Be explicit about which checks are npm requirements versus project policy.

## Core Workflow

1. **Read** `package.json` (abort on a parse error).
2. **Skip** when `private` is `true` — npm refuses to publish it; this is not an error.
3. **Hard rules (npm):** require `name` and `version`; `version` must be valid semver; `name` ≤ 214 chars, lowercase, URL/CLI/folder-safe.
4. **Policy checks (label them):** presence of `license`, `repository`, `files`. npm only _warns_ on a missing `license` and does not require the others — enforce them as a quality gate, not as "npm requires".
5. **`repository`** — accept the object `{type, url[, directory]}` **or** a shorthand string (`github:org/repo`, `org/repo`); don't force `type` when the url is a host shorthand.
6. **`access`** — only meaningful for `@scoped` names; unscoped packages are always public and cannot be restricted. Require an explicit access (via `publishConfig.access` or `--access`) only for scoped packages.
7. **`files` existence** — warn on a literal path that doesn't exist, but **skip glob entries** (`*`, `**/*`), build outputs created later (e.g. `dist/`), and the always-included `package.json` / `README` / `LICENSE` / `main`.
8. **Report** errors (block) and warnings (continue); exit non-zero on any error.

## Output

Ready:

```
@scope/pkg@1.5.0 is ready for publishing
```

Not ready:

```
@scope/pkg is not ready for publishing:
  - Missing required field: license
  Warning: file "dist/index.js" does not exist yet (built later)
```

## Error Handling

- **Error** — unparsable `package.json` → abort.
- **Error** — missing `name` / `version`, or an invalid `version` → not publishable.
- **Warning** — a missing policy field, or a literal `files` path absent → surface and continue.

## Gotchas

- Don't report "files path missing" for globs or `dist/` — they resolve at pack time, after the check, and `package.json` / `README` / `LICENSE` / `main` ship regardless.
- `publishConfig.access: "restricted"` (or its alias `"private"`) is **not** the same as top-level `private: true`; only the latter blocks publishing.
- Requiring `license` / `repository` / `files` is stricter than npm — keep it if you like, but call it policy so it isn't mistaken for an npm rule.
