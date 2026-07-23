# moon-skill-validate

Validates a `SKILL.md` against the Agent Skills spec and authoring best practices (frontmatter, body limits, reference links, progressive-disclosure triggers, harness-neutrality). Read-only.

## Usage

```bash
npx moon-skill-validate [--strict] [skill-dir | path/to/SKILL.md]   # defaults to the current directory
```

Besides the Agent Skills frontmatter/body checks, validation requires the package evidence that keeps a catalog maintainable:

- one loader-safe, double-quoted `description` scalar with single quotes for inner phrases;
- structurally valid `evals.json` with at least three output probes;
- `eval-queries.json` with at least eight positive and eight negative routes, including train/validation coverage for each polarity;
- reviewed `SOURCES.md` URL or explicit repository-original provenance;
- `compatibility` and `allowed-tools` frontmatter for any skill that bundles scripts.

Use `--strict` in CI so authoring warnings fail validation instead of returning a successful exit code.

The workspace-level composition check verifies boundary-crossing Markdown links and named advisory handoffs. It also requires matching Claude and Codex manifests whose names match their packages and rejects duplicate plugin names, dangling dependencies, and dependency cycles. Exact hard dependencies are machine-readable manifest edges; the runtime resolver traverses them instead of inferring loading behavior from prose.

The same check validates `packages/skill/composition-catalog.json`: every installed guide has one lifecycle and primary functional role; provision versions and requirement ranges use SemVer; required semantic requirements resolve to exactly one compatible installed provider; and the selected semantic graph is acyclic. Semantic selection records the catalog digest, exact installed implementation, reason, and content-provenance path without changing manifest dependency behavior. It also requires the generated catalog snapshot packaged under workflow-guide assets to be byte-identical, so installed workflow composition never depends on access to the monorepo root.

```bash
npx moon run script-moon-skill-validate:composition-sync
npx moon run script-moon-skill-validate:composition-check
```

Run `composition-sync` after editing the canonical catalog. It copies the exact
canonical bytes into the workflow skill's packaged snapshot; `composition-check`
then verifies identity and all semantic contracts without writing files.
