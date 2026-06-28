# Sources

## npm publish (CLI v11)

- **URL:** https://docs.npmjs.com/cli/v11/commands/npm-publish
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/publish.md`
  - `SKILL.md` → Core Principles
- **Aspects extracted:**
  - Pack/upload behavior, `--access` / `--provenance` / `--dry-run` / `--tag` flags, default `latest` dist-tag

## npm view (CLI v11)

- **URL:** https://docs.npmjs.com/cli/v11/commands/npm-view
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/publish.md` → Core Workflow (registry probe)
  - `SKILL.md` → Core Principles (publish idempotently)
- **Aspects extracted:**
  - Reading a field for an explicit `name@version`; a missing package/version yields E404 (non-zero exit) while a present version exits 0 with output; the bare-name / `@latest` empty-output edge case that motivates probing the explicit version

## npm package.json reference (CLI v11)

- **URL:** https://docs.npmjs.com/cli/v11/configuring-npm/package-json/
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/publish-check.md`
  - `references/publish.md` → Platform-Specific Packages
- **Aspects extracted:**
  - Required (`name`/`version`) vs recommended fields, `repository` object/shorthand forms, `files` allowlist and always-included files, `publishConfig`, `private`, `os`/`cpu`/`libc`

## Generating provenance statements

- **URL:** https://docs.npmjs.com/generating-provenance-statements
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/publish.md` → Access / Provenance / Dist-tag
- **Aspects extracted:**
  - Provenance prerequisites: supported CI with OIDC, `id-token: write`, public registry, matching `repository`, npm ≥ 9.5.0

## npm scope and access (CLI v11)

- **URL:** https://docs.npmjs.com/cli/v11/using-npm/scope/
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/publish.md` → Access / Provenance / Dist-tag
  - `references/publish-check.md` → access check
- **Aspects extracted:**
  - Scoped packages default to restricted; unscoped packages are always public and cannot be restricted

## Trusted publishers (OIDC)

- **URL:** https://docs.npmjs.com/trusted-publishers
- **Last reviewed:** 2026-06-28
- **Used for:**
  - `references/publish.md` → Provenance
- **Aspects extracted:**
  - Automatic provenance via OIDC trusted publishing and its version prerequisites (npm ≥ 11.5.1, Node ≥ 22.14.0)

## Refresh Workflow

1. Re-fetch the sources above
2. Diff against the prior pull (or scan for newly added sections)
3. Update the corresponding `references/*.md`
4. Bump the **Last reviewed** dates above
