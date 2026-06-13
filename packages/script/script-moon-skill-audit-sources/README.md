# moon-skill-audit-sources

Audits a skill's `SOURCES.md` for drift: staleness (days since **Last reviewed** vs `--max-age`), dangling provenance (referenced files that no longer exist), and reference files with no upstream source. `--fetch` HTTP-checks each URL; `--mark-reviewed` stamps the date after a human re-verifies. Does **not** rewrite distilled prose.

## Usage

```bash
npx moon-skill-audit-sources [skill-dir | SOURCES.md]   # defaults to cwd
npx moon-skill-audit-sources --all packages/skill --max-age 180
npx moon-skill-audit-sources <skill-dir> --mark-reviewed
```
