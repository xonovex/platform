# moon-skill-validate

Validates a `SKILL.md` against the Agent Skills spec and authoring best practices (frontmatter, body limits, reference links, progressive-disclosure triggers, harness-neutrality). Read-only.

## Usage

```bash
npx moon-skill-validate [skill-dir | path/to/SKILL.md]   # defaults to the current directory
```

Besides the Agent Skills frontmatter/body checks, validation requires the package evidence that keeps a catalog maintainable:

- structurally valid `evals.json` with at least three output probes;
- `eval-queries.json` with at least eight positive and eight negative routes, including train/validation coverage for each polarity;
- reviewed `SOURCES.md` URL or explicit repository-original provenance;
- `compatibility` and `allowed-tools` frontmatter for any skill that bundles scripts.

The workspace-level composition check verifies boundary-crossing Markdown links and named skill handoffs. It also requires matching Claude and Codex manifests whose names match their packages, requires every hard dependency to appear as a named handoff in the depending skill, and rejects duplicate plugin names, dangling dependencies, and dependency cycles:

```bash
npx moon run script-moon-skill-validate:composition-check
```
