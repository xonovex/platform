# moon-skill-validate

Validates a `SKILL.md` against the Agent Skills spec and authoring best practices (frontmatter, body limits, reference links, progressive-disclosure triggers, harness-neutrality). Every prose file in the package (guide, references, `SOURCES.md`, eval fixtures, bundled scripts, asset templates, manifests) is also checked for an em dash, ellipsis character, or typographic quote, in both the literal and unicode-escaped spelling. Read-only.

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

The workspace-level composition check verifies boundary-crossing Markdown links and
named advisory handoffs. It also requires matching Claude and Codex manifests whose
names match their packages and rejects duplicate plugin names, dangling hard
dependencies, and dependency cycles. Soft selection uses installed skill descriptions;
exact hard dependencies remain machine-readable manifest edges and load
dependency-first.

```bash
npx moon run script-moon-skill-validate:composition-check
```
