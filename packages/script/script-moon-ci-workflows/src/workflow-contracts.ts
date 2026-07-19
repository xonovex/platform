import {parse as parseYaml} from "yaml";
import {E2E_SUITES} from "./e2e.js";

type YamlRecord = Readonly<Record<string, unknown>>;

export interface WorkflowContractSources {
  readonly ci: string;
  readonly release: string;
  readonly e2e: string;
  readonly skillEvals: string;
  readonly operatorProject: string;
  readonly skillTasks: string;
  readonly workflowProject: string;
  readonly releaseValidatorProject: string;
}

const asRecord = (value: unknown): YamlRecord | undefined =>
  typeof value === "object" && value !== null && !Array.isArray(value)
    ? (value as YamlRecord)
    : undefined;

const parseRecord = (
  content: string,
  label: string,
  failures: string[],
): YamlRecord | undefined => {
  try {
    const parsed = asRecord(parseYaml(content) as unknown);
    if (parsed === undefined) {
      failures.push(`${label} must contain a YAML mapping`);
    }
    return parsed;
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    failures.push(`${label} is invalid YAML: ${detail}`);
    return undefined;
  }
};

const triggerNames = (workflow: YamlRecord | undefined): string[] => {
  const triggers = workflow?.on;
  if (typeof triggers === "string") return [triggers];
  return Object.keys(asRecord(triggers) ?? {}).toSorted();
};

const task = (
  project: YamlRecord | undefined,
  name: string,
): YamlRecord | undefined => asRecord(asRecord(project?.tasks)?.[name]);

const stringArray = (value: unknown): string[] =>
  Array.isArray(value)
    ? value.flatMap((entry) => (typeof entry === "string" ? [entry] : []))
    : [];

const dependencyNames = (value: unknown): string[] =>
  Array.isArray(value)
    ? value.flatMap((entry) => {
        if (typeof entry === "string") return [entry];
        const target = asRecord(entry)?.target;
        return typeof target === "string" ? [target] : [];
      })
    : [];

const requireText = (
  content: string,
  needle: string,
  message: string,
  failures: string[],
): void => {
  if (!content.includes(needle)) failures.push(message);
};

const forbidText = (
  content: string,
  needle: string,
  message: string,
  failures: string[],
): void => {
  if (content.includes(needle)) failures.push(message);
};

const requireExactTriggers = (
  workflow: YamlRecord | undefined,
  expected: readonly string[],
  label: string,
  failures: string[],
): void => {
  const actual = triggerNames(workflow);
  if (actual.join("\0") !== [...expected].toSorted().join("\0")) {
    failures.push(
      `${label} triggers must be exactly ${expected.toSorted().join(", ")}`,
    );
  }
};

export const workflowContractFailures = (
  sources: WorkflowContractSources,
): readonly string[] => {
  const failures: string[] = [];
  const ci = parseRecord(sources.ci, "CI workflow", failures);
  const release = parseRecord(sources.release, "release workflow", failures);
  const e2e = parseRecord(sources.e2e, "E2E workflow", failures);
  const skillEvals = parseRecord(
    sources.skillEvals,
    "skill-eval workflow",
    failures,
  );
  const operatorProject = parseRecord(
    sources.operatorProject,
    "operator Moon project",
    failures,
  );
  const skillTasks = parseRecord(sources.skillTasks, "skill tasks", failures);
  const workflowProject = parseRecord(
    sources.workflowProject,
    "CI workflow Moon project",
    failures,
  );
  const releaseValidatorProject = parseRecord(
    sources.releaseValidatorProject,
    "release validator Moon project",
    failures,
  );

  if (ci !== undefined) {
    requireText(
      sources.ci,
      "zizmor --offline .github/",
      "CI workflow must statically lint GitHub Actions",
      failures,
    );
    requireText(
      sources.ci,
      ":ci-check",
      "CI workflow must run the ordinary Moon ci-check target",
      failures,
    );
  }

  if (release !== undefined) {
    requireText(
      sources.release,
      ":ci-publish",
      "release workflow must run ci-publish",
      failures,
    );
    requireText(
      sources.release,
      ":ci-publish-dry-run",
      "release workflow must support a ci-publish dry run",
      failures,
    );
    forbidText(
      sources.release,
      ":ci-check",
      "release workflow must not duplicate ordinary CI",
      failures,
    );
  }

  requireExactTriggers(
    e2e,
    ["schedule", "workflow_dispatch"],
    "E2E workflow",
    failures,
  );
  for (const needle of [
    "script-moon-ci-workflows:build",
    "moon-ci-workflows e2e-matrix",
    "moon-ci-workflows e2e-run",
    "moon-ci-workflows e2e-summary",
  ]) {
    requireText(
      sources.e2e,
      needle,
      `E2E workflow must delegate through ${needle}`,
      failures,
    );
  }
  forbidText(
    sources.e2e,
    "go-test-integration",
    "E2E workflow must not repeat the integration suite from ordinary CI",
    failures,
  );
  const e2eJobs = asRecord(e2e?.jobs);
  if (asRecord(e2eJobs?.matrix) === undefined) {
    failures.push("E2E workflow must select its matrix in a dedicated job");
  }
  if (asRecord(e2eJobs?.test)?.needs !== "matrix") {
    failures.push("E2E test jobs must depend on matrix selection");
  }

  requireExactTriggers(
    skillEvals,
    ["pull_request", "schedule", "workflow_dispatch"],
    "skill-eval workflow",
    failures,
  );
  for (const needle of [
    "script-moon-ci-workflows:build",
    "moon-ci-workflows skill-eval-matrix",
    "moon-ci-workflows skill-eval-run",
    "ANTHROPIC_API_KEY",
    "actions/upload-artifact",
  ]) {
    requireText(
      sources.skillEvals,
      needle,
      `skill-eval workflow must retain ${needle}`,
      failures,
    );
  }
  for (const forbidden of [
    "npx moon-skill-eval-triggers",
    "npx moon-skill-eval-outputs",
    "list-eval-matrix.mjs",
    "while ((",
  ]) {
    forbidText(
      sources.skillEvals,
      forbidden,
      `skill-eval workflow must not implement orchestration with ${forbidden}`,
      failures,
    );
  }

  const operatorCiDeps = dependencyNames(
    task(operatorProject, "ci-check")?.deps,
  );
  if (!operatorCiDeps.includes("go-test-integration")) {
    failures.push("operator ci-check must retain the integration suite");
  }
  for (const {task: taskName} of E2E_SUITES) {
    const localTask = taskName.slice("agent-operator-go:".length);
    if (task(operatorProject, localTask) === undefined) {
      failures.push(`operator Moon project must define ${localTask}`);
    }
  }

  for (const taskName of ["skill-eval-triggers", "skill-eval-outputs"]) {
    if (asRecord(task(skillTasks, taskName)?.options)?.runInCI !== false) {
      failures.push(`${taskName} must remain excluded from ordinary CI`);
    }
  }

  const workflowCiDeps = dependencyNames(
    task(workflowProject, "ci-check")?.deps,
  );
  if (!workflowCiDeps.includes("workflow-validate")) {
    failures.push("workflow package ci-check must depend on workflow-validate");
  }
  const workflowInputs = stringArray(
    task(workflowProject, "workflow-validate")?.inputs,
  );
  for (const input of [
    "/.github/workflows/ci.yml",
    "/.github/workflows/release.yml",
    "/.github/workflows/e2e.yml",
    "/.github/workflows/skill-evals.yml",
  ]) {
    if (!workflowInputs.includes(input)) {
      failures.push(`workflow-validate inputs must include ${input}`);
    }
  }

  const releaseInputs = stringArray(
    task(releaseValidatorProject, "release-validate")?.inputs,
  );
  if (!releaseInputs.includes("/.github/workflows/release.yml")) {
    failures.push("release-validate inputs must include the release workflow");
  }

  return failures;
};
