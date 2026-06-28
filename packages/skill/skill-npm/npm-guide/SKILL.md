---
name: npm-guide
description: "Use when publishing an npm package or checking that one is ready to publish. Triggers on prompts about npm publish, publish-readiness or package.json validation (name/version/license/repository/files/publishConfig/private), idempotent publishing (skip if the version already exists), provenance and access flags, dist-tags, or platform-specific (os/cpu/libc) native-binary packages — even when the user doesn't say 'npm'."
---

# npm Publishing Guidelines

Validate publish-readiness and publish packages idempotently, with correct access, provenance, dist-tags, and platform fields.

## Core Principles

- **npm requires only name + version** — `license` / `repository` / `files` are house policy, not npm rules, and `private: true` blocks publishing outright, see [references/publish-check.md](references/publish-check.md)
- **Publish idempotently** — probe `npm view <name>@<version>` and skip if present; the registry also refuses to overwrite an existing version, see [references/publish.md](references/publish.md)
- **Provenance needs CI + OIDC** — `--provenance` only works from a supported runner with `id-token` permission, the public registry, and a matching `repository`; it errors elsewhere, see [references/publish.md](references/publish.md)
- **Access is derived** — unscoped packages are always public; scoped packages default to restricted and need `--access public` on first publish, see [references/publish.md](references/publish.md)
- **Platform packages gate install** — `os` / `cpu` / `libc` restrict where a native-binary package installs; inject and restore those fields safely, see [references/publish.md](references/publish.md)

## Operations

- **publish-check** — validate a package.json is publish-ready, separating npm rules from policy — see [references/publish-check.md](references/publish-check.md)
- **publish** — publish a package idempotently with provenance / access / dist-tag / platform handling — see [references/publish.md](references/publish.md)

## Gotchas

- Hard-requiring `license` / `repository` / `files` exceeds npm (it warns, it doesn't block), and rejecting a `repository` shorthand string like `github:org/repo` is wrong — treat the three as policy
- `publishConfig.access` is only meaningful for `@scoped` names; unscoped packages cannot be restricted — don't hard-require it universally
- `--provenance` fails outside a supported CI/OIDC environment — gate it on CI, never expect a local no-op
- `files` supports globs and `dist/` is built later — don't existence-check glob entries or build outputs; `package.json` / `README` / `LICENSE` / `main` are always included
- Inject-then-restore of `os` / `cpu` / `libc` must restore on failure (try/finally), and `libc` is a silent no-op unless `os` includes `linux`
- A pre-release published without `--tag` lands on `latest` and installs by default — derive a dist-tag from the version

## Progressive Disclosure

- Read [references/publish-check.md](references/publish-check.md) - Load when validating that a package is ready to publish
- Read [references/publish.md](references/publish.md) - Load when publishing (idempotency, provenance, access, dist-tag, platform fields)
