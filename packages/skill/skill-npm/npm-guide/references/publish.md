# publish: Idempotent npm Publish

Publish a package only if its exact version isn't already on the registry, with correct access, provenance, dist-tag, and platform handling.

## Contents

- [Goal](#goal)
- [Core Workflow](#core-workflow)
- [Access / Provenance / Dist-tag](#access--provenance--dist-tag)
- [Platform-Specific Packages](#platform-specific-packages)
- [Output](#output)
- [Error Handling](#error-handling)
- [Gotchas](#gotchas)

## Goal

- Make publishing safe to re-run: never attempt to overwrite an existing `name@version`.
- Apply the right access, provenance, and dist-tag for the version and environment.

## Core Workflow

1. **Read** `name` and `version`.
2. **Probe** the registry: `npm view <name>@<version> version`. Exit 0 with non-empty output → already published → skip. E404 / non-zero → proceed.
3. **(Optional) Inject platform fields** from a sibling `platform.json` (`os` / `cpu` / `libc`) into `package.json` — see [Platform-Specific Packages](#platform-specific-packages). Wrap in try/finally so the original is **restored even on failure**.
4. **Publish:** `npm publish [--provenance] --access public [--tag <tag>]`, or `npm publish --dry-run --access public` to preview.
5. **Restore** `package.json` if it was modified.

## Access / Provenance / Dist-tag

- **Access** — unscoped packages are always public (`--access public` is a harmless no-op); scoped packages default to restricted, so `--access public` is required on the first public publish. Passing `--access public` unconditionally is safe for both.
- **Provenance** — `--provenance` only works from a supported CI runner (GitHub Actions / GitLab CI) with OIDC (`id-token: write`), npm ≥ 9.5.0, the public npm registry, and a `repository` field that matches (case-sensitively) the publish source. Outside that it **fails** — gate it on CI; don't run it locally expecting a no-op. With OIDC trusted publishing (npm ≥ 11.5.1) provenance is automatic and the flag is redundant.
- **Dist-tag** — with no `--tag`, npm publishes to `latest`, which `npm install <name>` resolves. A pre-release version (its `version` carries a pre-release identifier) must use a tag like `--tag next`, or it ships as `latest` and installs by default.

## Platform-Specific Packages

- `os` / `cpu` / `libc` gate **installation** to matching hosts — the native-binary optional-dependency pattern.
- `os` takes `process.platform` values, `cpu` takes `process.arch` values, both support `!` blocklisting; `libc` applies **only when `os` includes `linux`** (otherwise a silent no-op).
- Inject from `platform.json`, publish, then restore: npm packs from the on-disk `package.json`, so the injected fields must be present at pack time and gone afterward. Don't let sibling-platform publishes race on the same file.

## Output

```
Skipping @scope/pkg@1.5.0 — already published
```

or the streamed `npm publish` output, followed by a restore note when platform fields were injected.

## Error Handling

- **Skip (not error)** — the probe finds the version already published.
- **Error** — `npm publish` fails: provenance prerequisites unmet, missing auth / `--otp` for a 2FA account, or a 403 "cannot publish over previously published version" (the registry's own overwrite guard — the probe is a fail-safe, not a guarantee).
- A probe E404 from a mis-pointed registry / `.npmrc` reads as "not published" and triggers a publish that then fails safely — verify the registry when an unexpected publish is attempted.

## Gotchas

- `--provenance` is not a no-op off-CI — it errors; never assume it degrades gracefully.
- Restore `package.json` in a `finally` — a crash mid-publish otherwise leaves injected `os` / `cpu` / `libc` baked into the file.
- Probing a bare name or `@latest` (no explicit version) can exit 0 with empty output on some registries — always probe the explicit `name@version`.
- A pre-release published to `latest` becomes the default install — set a dist-tag for non-stable versions.
