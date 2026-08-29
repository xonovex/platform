# moon-skill-audit-sources

Use this command to find stale, missing, or incomplete source evidence in a skill's `SOURCES.md`. It reports staleness, dangling provenance, and reference files without an upstream source. It does not rewrite distilled prose.

Use `--fetch` to check each URL over HTTP. Use `--mark-reviewed` only after a human verifies the sources again. For a skill distilled from a local, version-tagged source repository, a source block can opt in to upstream-drift detection. The report then compares the pinned version with the latest release and maps watched-path commits to the reference files they affect.

## Usage

Run the audit for one skill, the complete catalog, or a human-reviewed source update.

```bash
npx moon-skill-audit-sources [skill-dir | SOURCES.md]   # defaults to cwd
npx moon-skill-audit-sources --all packages/skill --max-age 180 --version-max-age 90
npx moon-skill-audit-sources <skill-dir> --mark-reviewed
npx moon-skill-audit-sources <skill-dir> --pull          # fetch tags before the drift check
```

Each source block must identify its source, the reference files it supports, and its last review date. Use `**URL:**`, an inline or bulleted `**URLs:**` list, or `**Provenance:**` for repository-original material without a truthful upstream URL. Declare coverage with `**References:** all`, a comma-separated list such as `**References:** references/auth.md, references/onboarding.md`, or an existing machine-readable `**Used for:**` or `**Aspects extracted:**` mapping. Prefer `**References:**` for new and edited blocks. Every reference file needs coverage from at least one source block. Every real source block needs `**Last reviewed:** YYYY-MM-DD`. A guide without a recognized source block fails. The `--all` option audits only `SOURCES.md` files next to a real `SKILL.md`, so templates and assets do not increase catalog counts.

JSON reports retain each source's declared version, commit, and watch count even when no local checkout is available. The catalog validator requires a `**Version:**` baseline whenever a skill description pins a `N+`, `N.N+`, or `N.N.N+` API generation; do not invent a checkout or commit when the source was reviewed from published documentation.

## Upstream-drift fields (optional, per source block)

Add these fields only when the audit can inspect a local source repository.

```
**Checkout:** ../some-source-repo        # local source repo, relative to workspace root
**Version:** 3.12.1                       # version the skill is pinned to
**Commit:** 76fa2c95...                     # commit the skill was distilled from
**Watch:** src/components -> components.md, navigation.md  # source subpath -> references it feeds
```

When these fields are present, the audit reports `pinned X, behind: latest released Y`. It lists commits on the watched paths since the pinned commit and the `references/*` files to review. Source blocks without `**Checkout:**` use the standard audit.
