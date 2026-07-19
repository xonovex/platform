# moon-ci-workflows

Owns the deterministic parts of the repository's specialized GitHub Actions workflows. GitHub Actions retains scheduling, credentials, runner setup, artifact upload, and matrix fan-out; this package owns E2E suite selection and execution, skill-eval selection and bounded batching, E2E summaries, and static workflow contracts.

## Usage

```bash
npx moon-ci-workflows validate
npx moon-ci-workflows e2e-matrix
npx moon-ci-workflows e2e-run e2e-kata
npx moon-ci-workflows e2e-summary e2e-kata suite.log
npx moon-ci-workflows skill-eval-matrix packages/skill
npx moon-ci-workflows skill-eval-run trigger testing testing-guide
```

The model-eval command intentionally uses the repository's fixed budget, batch-size, model, and concurrency policy. It requires the evaluator CLI and model credential at runtime; ordinary `ci-check` performs only static contract validation and unit tests.
