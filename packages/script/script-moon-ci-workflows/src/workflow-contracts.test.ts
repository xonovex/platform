import {describe, expect, it} from "vitest";
import {
  workflowContractFailures,
  type WorkflowContractSources,
} from "./workflow-contracts.js";

const validSources = (): WorkflowContractSources => ({
  ci: `
name: CI
on: { pull_request: null }
jobs:
  ci:
    steps:
      - run: zizmor --offline .github/
      - run: npx moon ci :ci-check
`,
  release: `
name: Release
on: { workflow_dispatch: null, pull_request: null }
jobs:
  release:
    steps:
      - run: npx moon ci :ci-publish
      - run: npx moon ci :ci-publish-dry-run
`,
  e2e: `
name: E2E
on: { workflow_dispatch: null, schedule: [{ cron: "0 0 * * *" }] }
jobs:
  matrix:
    steps:
      - run: npx moon run script-moon-ci-workflows:build
      - run: moon-ci-workflows e2e-matrix
  test:
    needs: matrix
    steps:
      - run: moon-ci-workflows e2e-run "$SUITE"
      - run: moon-ci-workflows e2e-summary "$SUITE" suite.log
`,
  skillEvals: `
name: Skill evals
on: { pull_request: null, workflow_dispatch: null, schedule: [{ cron: "0 0 1 * *" }] }
jobs:
  matrix:
    steps:
      - run: npx moon run script-moon-ci-workflows:build
      - run: moon-ci-workflows skill-eval-matrix
  eval:
    steps:
      - run: moon-ci-workflows skill-eval-run "$MODE" "$PACKAGE" "$GUIDE"
        env: { ANTHROPIC_API_KEY: secret }
      - uses: actions/upload-artifact@sha
`,
  operatorProject: `
tasks:
  ci-check: { deps: [go-test-integration] }
  go-test-e2e: { command: go }
  go-test-e2e-gvisor: { command: go }
  go-test-e2e-kata: { command: go }
  go-test-e2e-coco: { command: go }
`,
  skillTasks: `
tasks:
  skill-eval-triggers: { options: { runInCI: false } }
  skill-eval-outputs: { options: { runInCI: false } }
`,
  workflowProject: `
tasks:
  workflow-validate:
    inputs:
      - /.github/workflows/ci.yml
      - /.github/workflows/release.yml
      - /.github/workflows/e2e.yml
      - /.github/workflows/skill-evals.yml
  ci-check: { deps: [workflow-validate] }
`,
  releaseValidatorProject: `
tasks:
  release-validate:
    inputs: [/.github/workflows/release.yml]
`,
});

describe("workflowContractFailures", () => {
  it("accepts thin specialized runners backed by ordinary static CI", () => {
    expect(workflowContractFailures(validSources())).toEqual([]);
  });

  it("rejects integration coverage duplicated in the E2E workflow", () => {
    const sources = validSources();
    expect(
      workflowContractFailures({
        ...sources,
        e2e: `${sources.e2e}\n# agent-operator-go:go-test-integration`,
      }),
    ).toContain(
      "E2E workflow must not repeat the integration suite from ordinary CI",
    );
  });

  it("rejects shell batching and direct evaluator calls", () => {
    const sources = validSources();
    const failures = workflowContractFailures({
      ...sources,
      skillEvals: `${sources.skillEvals}\n# while (( offset < count )); npx moon-skill-eval-outputs`,
    });

    expect(failures).toContain(
      "skill-eval workflow must not implement orchestration with while ((",
    );
    expect(failures).toContain(
      "skill-eval workflow must not implement orchestration with npx moon-skill-eval-outputs",
    );
  });

  it("requires workflow files to invalidate the Moon task cache", () => {
    const sources = validSources();
    expect(
      workflowContractFailures({
        ...sources,
        workflowProject: sources.workflowProject.replace(
          "      - /.github/workflows/e2e.yml\n",
          "",
        ),
      }),
    ).toContain(
      "workflow-validate inputs must include /.github/workflows/e2e.yml",
    );
  });
});
