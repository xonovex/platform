# moon-skill-validate

Use this read-only validator to check a skill package against the Agent Skills specification and Xonovex authoring rules. It checks frontmatter, body limits, reference links, progressive-disclosure triggers, harness neutrality, and supporting evidence.

The validator also rejects an em dash, an ellipsis character, or a typographic quote in every prose file in the package, in literal or Unicode-escaped form.

## Usage

Run strict mode in continuous integration so authoring warnings fail the task.

```bash
npx moon-skill-validate [--strict] [skill-dir | path/to/SKILL.md]   # defaults to the current directory
```

Each skill package must include this evidence:

- one loader-safe, double-quoted `description` scalar with single quotes for inner phrases;
- structurally valid `evals.json` with at least three output probes;
- `eval-queries.json` with at least eight positive and eight negative routes, including train/validation coverage for each polarity;
- reviewed `SOURCES.md` URL or explicit repository-original provenance;
- `compatibility` and `allowed-tools` frontmatter for any skill that bundles scripts.

The workspace-level composition check verifies cross-package Markdown links and named advisory handoffs. It also requires matching Claude and Codex manifests whose names match their packages. It rejects duplicate plugin names, dangling hard dependencies, and dependency cycles. Soft selection uses installed skill descriptions. Exact hard dependencies remain machine-readable manifest edges and load before their dependents.

```bash
npx moon run script-moon-skill-validate:composition-check
```
