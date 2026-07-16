# publish: npm Publish

Publish a package with correct access, provenance, dist-tag, and platform handling.

## Core Workflow

1. **(Optional) Inject platform fields** from a sibling `platform.json` (`os` / `cpu` / `libc`) into `package.json`, wrapped in try/finally so the original is **restored even on failure** — npm packs from the on-disk `package.json`, so injected fields must be present at pack time and gone afterward. Don't let sibling-platform publishes race on the same file.
2. **Publish:** `npm publish [--provenance] --access public [--tag <tag>]`, or `npm publish --dry-run --access public` to preview.

## Access / Provenance / Dist-tag

- **Access** — unscoped packages are always public (`--access public` is a harmless no-op); scoped packages default to restricted, so `--access public` is required on the first public publish. Passing it unconditionally is safe for both.
- **Provenance** — `--provenance` only works from a supported CI runner (GitHub Actions / GitLab CI) with OIDC (`id-token: write`), npm ≥ 9.5.0, the public npm registry, and a `repository` field that matches (case-sensitively) the publish source. Outside that it **fails** — gate it on CI; it does not degrade to a no-op. With OIDC trusted publishing (npm ≥ 11.5.1) provenance is automatic and the flag is redundant.
- **Dist-tag** — with no `--tag`, npm publishes to `latest`, which `npm install <name>` resolves. A pre-release version (its `version` carries a pre-release identifier) must use a tag like `--tag next`, or it ships as `latest` and installs by default.

## Platform-Specific Packages

`os` / `cpu` / `libc` gate **installation** to matching hosts (the native-binary optional-dependency pattern). `os` takes `process.platform` values, `cpu` takes `process.arch` values, both support `!` blocklisting; `libc` applies **only when `os` includes `linux`** (otherwise a silent no-op).

## Error Handling

- **Error** — `npm publish` fails: provenance prerequisites unmet, missing auth / `--otp` for a 2FA account, or a 403 "cannot publish over previously published version" (the registry's own overwrite guard).
- A probe E404 from a mis-pointed registry / `.npmrc` reads as "not published" and triggers a publish that then fails safely — verify the registry when an unexpected publish is attempted.
