# semver: Computing, Validating, and Ordering Versions

How to increment, validate, and order versions under Semantic Versioning 2.0.0.

## Increment Rules

- Versions are `MAJOR.MINOR.PATCH`, each a non-negative integer with **no leading zeroes**.
- **PATCH** — backward-compatible bug fixes.
- **MINOR** — backward-compatible features (or marking the API deprecated); resets PATCH to 0.
- **MAJOR** — any backward-incompatible change; resets MINOR and PATCH to 0.
- `0.y.z` is for initial development: anything may change at any time and the public API is not considered stable.

## Pre-releases

- Append `-` then dot-separated identifiers: `1.2.3-alpha`, `1.2.3-alpha.1`, `1.2.3-0.3.7`, `1.2.3-x-y-z.--`.
- Each identifier is drawn from `[0-9A-Za-z-]`. Numeric identifiers carry **no leading zeroes**, but `0` alone is valid — so a `-beta.0` start is spec-legal.
- A pre-release has **lower** precedence than its release: `1.2.3-beta.2 < 1.2.3`. Finalizing `X.Y.Z-id.N` therefore yields `X.Y.Z` (the same core), not the next patch.
- A `--preid` flow: the first prerelease is `X.Y.Z-id.0`; bumping again with the same id increments the trailing numeric counter (`id.0` → `id.1`).

## Build Metadata

- Append `+` then identifiers (`1.0.0+20130313`, `1.0.0-beta+exp.sha.5114f85`). Leading zeroes are allowed here.
- Build metadata is **ignored** for precedence — never order two versions by it.

## Validation

Match the full grammar, not a shortcut. Common pitfalls:

- `\w` is wrong for identifiers — it adds `_` (illegal) and drops `-` (legal). Use `[0-9A-Za-z-]`.
- `\d+` over-accepts leading zeroes — reject `01.2.3` and `1.2.3-beta.01`.
- A pre-release is _N_ dot-separated identifiers, not a single `word.number` pair — `1.0.0-alpha`, `1.0.0-0.3.7`, and `1.0.0-x.7.z.92` are all valid and a `word.number`-only regex wrongly rejects them.
- When in doubt, defer to a semver library's parse/validate rather than a hand-rolled regex.

## Ordering

- Compare core fields numerically; a version carrying a pre-release ranks below the same core without one.
- Compare pre-release identifiers field by field: numeric identifiers compare numerically and rank below non-numeric ones, and a longer set of fields outranks a shorter prefix-equal set.
