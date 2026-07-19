#!/usr/bin/env node
import {spawnSync} from "node:child_process";
import {existsSync, readFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  e2eMatrix,
  e2eTask,
  renderE2eSummary,
  renderMissingE2eSummary,
} from "./e2e.js";
import {runSkillEvals, type SkillEvalMode} from "./eval-runner.js";
import {
  changedFiles,
  discoverSkills,
  selectChangedSkills,
} from "./skill-evals.js";
import {
  workflowContractFailures,
  type WorkflowContractSources,
} from "./workflow-contracts.js";

const PROG = "moon-ci-workflows";
const USAGE = `Usage:
  ${PROG} validate [repository-root]
  ${PROG} e2e-matrix
  ${PROG} e2e-run SUITE
  ${PROG} e2e-summary SUITE LOG
  ${PROG} skill-eval-matrix [catalog-root] [--changed BASE HEAD]
  ${PROG} skill-eval-run MODE PACKAGE GUIDE [--repository-root PATH] [--temporary-root PATH]`;

const usageError = (message: string): never => {
  throw new Error(`${message}\n${USAGE}`);
};

const requiredArgument = (value: string | undefined, message: string): string =>
  value ?? usageError(message);

const readSources = (repositoryRoot: string): WorkflowContractSources => {
  const read = (path: string): string =>
    readFileSync(resolve(repositoryRoot, path), "utf8");
  return {
    ci: read(".github/workflows/ci.yml"),
    release: read(".github/workflows/release.yml"),
    e2e: read(".github/workflows/e2e.yml"),
    skillEvals: read(".github/workflows/skill-evals.yml"),
    operatorProject: read("packages/agent/agent-operator-go/moon.yml"),
    skillTasks: read(".moon/tasks/tag-skill.yml"),
    workflowProject: read("packages/script/script-moon-ci-workflows/moon.yml"),
    releaseValidatorProject: read(
      "packages/script/script-moon-release-validate/moon.yml",
    ),
  };
};

const validate = (rootArgument: string | undefined): number => {
  const repositoryRoot = resolve(rootArgument ?? ".");
  const failures = workflowContractFailures(readSources(repositoryRoot));
  if (failures.length > 0) {
    for (const failure of failures) process.stderr.write(`FAIL: ${failure}\n`);
    return 1;
  }
  process.stdout.write("CI workflow contracts valid\n");
  return 0;
};

const runE2e = (suite: string | undefined): number => {
  const selectedSuite = requiredArgument(suite, "e2e-run requires SUITE");
  const result = spawnSync(
    resolveExecutable("npx"),
    ["moon", "run", e2eTask(selectedSuite)],
    {stdio: "inherit"},
  );
  if (result.error !== undefined) throw result.error;
  return result.status ?? 1;
};

const summarizeE2e = (
  suite: string | undefined,
  log: string | undefined,
): number => {
  const selectedSuite = requiredArgument(
    suite,
    "e2e-summary requires SUITE and LOG",
  );
  const logPath = requiredArgument(log, "e2e-summary requires SUITE and LOG");
  process.stdout.write(
    existsSync(logPath)
      ? renderE2eSummary(selectedSuite, readFileSync(logPath, "utf8"))
      : renderMissingE2eSummary(selectedSuite),
  );
  return 0;
};

interface MatrixArguments {
  readonly catalogRoot: string;
  readonly changed?: {readonly base: string; readonly head: string};
}

const parseMatrixArguments = (args: readonly string[]): MatrixArguments => {
  const changedIndex = args.indexOf("--changed");
  const excluded = new Set(
    changedIndex === -1
      ? []
      : [changedIndex, changedIndex + 1, changedIndex + 2],
  );
  const positionals = args.filter((_, index) => !excluded.has(index));
  if (positionals.length > 1) usageError("expected at most one catalog root");
  const catalogRoot = resolve(positionals[0] ?? "packages/skill");
  if (changedIndex === -1) return {catalogRoot};
  const base = requiredArgument(
    args[changedIndex + 1],
    "--changed requires BASE and HEAD revisions",
  );
  const head = requiredArgument(
    args[changedIndex + 2],
    "--changed requires BASE and HEAD revisions",
  );
  return {catalogRoot, changed: {base, head}};
};

const skillEvalMatrix = (args: readonly string[]): number => {
  const parsed = parseMatrixArguments(args);
  const skills = discoverSkills(parsed.catalogRoot);
  const selected =
    parsed.changed === undefined
      ? skills
      : selectChangedSkills(
          skills,
          changedFiles(parsed.changed.base, parsed.changed.head, resolve(".")),
        );
  process.stdout.write(`${JSON.stringify(selected)}\n`);
  return 0;
};

const option = (
  args: readonly string[],
  name: string,
  fallback: string,
): string => {
  const index = args.indexOf(name);
  if (index === -1) return fallback;
  return requiredArgument(args[index + 1], `${name} requires a value`);
};

const skillEvalRun = (args: readonly string[]): number => {
  const [mode, packageName, guide] = args;
  if (mode !== "trigger" && mode !== "output") {
    usageError("skill-eval-run MODE must be trigger or output");
  }
  const selectedPackage = requiredArgument(
    packageName,
    "skill-eval-run requires MODE, PACKAGE, and GUIDE",
  );
  const selectedGuide = requiredArgument(
    guide,
    "skill-eval-run requires MODE, PACKAGE, and GUIDE",
  );
  runSkillEvals({
    mode: mode as SkillEvalMode,
    packageName: selectedPackage,
    guide: selectedGuide,
    repositoryRoot: option(args, "--repository-root", resolve(".")),
    temporaryRoot: option(args, "--temporary-root", tmpdir()),
  });
  return 0;
};

const main = (argv: readonly string[]): number => {
  const [command, ...args] = argv;
  if (command === undefined || ["-h", "--help"].includes(command)) {
    process.stdout.write(`${USAGE}\n`);
    return command === undefined ? 2 : 0;
  }
  switch (command) {
    case "validate": {
      return validate(args[0]);
    }
    case "e2e-matrix": {
      process.stdout.write(`${JSON.stringify(e2eMatrix())}\n`);
      return 0;
    }
    case "e2e-run": {
      return runE2e(args[0]);
    }
    case "e2e-summary": {
      return summarizeE2e(args[0], args[1]);
    }
    case "skill-eval-matrix": {
      return skillEvalMatrix(args);
    }
    case "skill-eval-run": {
      return skillEvalRun(args);
    }
    default: {
      return usageError(`unknown command: ${command}`);
    }
  }
};

try {
  process.exitCode = main(process.argv.slice(2));
} catch (error) {
  const detail = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${PROG}: error: ${detail}\n`);
  process.exitCode = 2;
}
