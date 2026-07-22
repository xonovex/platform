# moon-skill-audit-sources

Audits a skill's `SOURCES.md` for drift: staleness (days since **Last reviewed** vs `--max-age`, or the shorter `--version-max-age` for versioned sources), dangling provenance (referenced files that no longer exist), and reference files with no upstream source. `--fetch` HTTP-checks each URL; `--mark-reviewed` stamps the date after a human re-verifies. Does **not** rewrite distilled prose.

For skills distilled from a locally-checked-out, version-tagged source repo, a source block can also opt into **upstream-drift detection** — it then reports the latest released tag vs the pinned version and the commits since the pinned commit on watched paths, mapped to the reference files they feed.

## Usage

```bash
npx moon-skill-audit-sources [skill-dir | SOURCES.md]   # defaults to cwd
npx moon-skill-audit-sources --all packages/skill --max-age 180 --version-max-age 90
npx moon-skill-audit-sources <skill-dir> --mark-reviewed
npx moon-skill-audit-sources <skill-dir> --pull          # fetch tags before the drift check
```

Source blocks may use `**URL:**`, an inline or bulleted `**URLs:**` list, or `**Provenance:**` for repository-original material without a truthful upstream URL. Every source block must declare what it feeds with `**References:** all`, a comma-separated list such as `**References:** references/auth.md, references/onboarding.md`, or an existing machine-readable `**Used for:**` / `**Aspects extracted:**` mapping that names or points to the reference files. Prefer `**References:**` for new and edited blocks. Every reference file must be covered by at least one source block. Every real source block needs `**Last reviewed:** YYYY-MM-DD`; a guide with no recognized source blocks fails. `--all` audits only `SOURCES.md` files adjacent to a real `SKILL.md`, so templates and assets do not inflate catalog counts.

JSON reports retain each source's declared version, commit, and watch count even when no local checkout is available. The catalog validator requires a `**Version:**` baseline whenever a skill description pins a `N+`, `N.N+`, or `N.N.N+` API generation; do not invent a checkout or commit when the source was reviewed from published documentation.

## Upstream-drift fields (optional, per source block)

```
**Checkout:** ../some-source-repo        # local source repo, relative to workspace root
**Version:** 3.12.1                       # version the skill is pinned to
**Commit:** 76fa2c95…                      # commit the skill was distilled from
**Watch:** src/components → components.md, navigation.md   # source subpath → references it feeds
```

When present, the audit reports `pinned X, behind — latest released Y` and lists the commits on the watched paths since the pinned commit, plus the `references/*` to review. Source blocks without `**Checkout:**` are audited exactly as before.
