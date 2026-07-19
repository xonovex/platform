import {spawnSync} from "node:child_process";
import {mkdirSync, mkdtempSync, readFileSync, writeFileSync} from "node:fs";
import {join, resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  assertSkillSegment,
  outputEvalBatches,
  triggerEvalBatches,
} from "./skill-evals.js";

export type SkillEvalMode = "trigger" | "output";

export interface SkillEvalRun {
  readonly mode: SkillEvalMode;
  readonly packageName: string;
  readonly guide: string;
  readonly repositoryRoot: string;
  readonly temporaryRoot: string;
}

interface EvaluatorResult {
  readonly output: string;
  readonly status: number | null;
  readonly signal: NodeJS.Signals | null;
}

const writeJson = (path: string, value: unknown): void => {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`, "utf8");
};

const runEvaluator = (
  executable: string,
  args: readonly string[],
  cwd: string,
  captureOutput: boolean,
): EvaluatorResult => {
  const result = spawnSync(resolveExecutable(executable), [...args], {
    cwd,
    encoding: "utf8",
    maxBuffer: 50 * 1024 * 1024,
    stdio: captureOutput ? ["ignore", "pipe", "inherit"] : "inherit",
  });
  if (result.error !== undefined) throw result.error;
  return {
    output: captureOutput ? result.stdout : "",
    status: result.status,
    signal: result.signal,
  };
};

const assertEvaluatorSucceeded = (
  executable: string,
  result: EvaluatorResult,
): void => {
  if (result.status !== 0) {
    throw new Error(
      `${executable} exited with status ${String(result.status ?? result.signal)}`,
    );
  }
};

const runTriggerEvals = (run: SkillEvalRun): void => {
  const pluginDirectory = join(
    run.repositoryRoot,
    "packages",
    "skill",
    `skill-${run.packageName}`,
  );
  const source = join(pluginDirectory, run.guide, "eval-queries.json");
  const batches = triggerEvalBatches(readFileSync(source, "utf8"));
  const batchDirectory = mkdtempSync(
    join(run.temporaryRoot, `${run.packageName}-trigger-`),
  );
  const resultDirectory = join(
    run.repositoryRoot,
    ".skill-eval-results",
    run.packageName,
    "trigger",
  );
  mkdirSync(resultDirectory, {recursive: true});

  for (const [index, batch] of batches.entries()) {
    const batchName = `batch-${String(index + 1)}`;
    const batchFile = join(batchDirectory, `${batchName}.json`);
    writeJson(batchFile, batch);
    const result = runEvaluator(
      "moon-skill-eval-triggers",
      [
        batchFile,
        `xonovex-skill-${run.packageName}:${run.guide}`,
        "validation",
        "--runs",
        "3",
        "--model",
        "haiku",
        "--max-budget-usd",
        "0.05",
        "--plugin-dir",
        pluginDirectory,
      ],
      run.repositoryRoot,
      true,
    );
    writeFileSync(
      join(resultDirectory, `${batchName}.txt`),
      result.output,
      "utf8",
    );
    process.stdout.write(result.output);
    assertEvaluatorSucceeded("moon-skill-eval-triggers", result);
  }
};

const runOutputEvals = (run: SkillEvalRun): void => {
  const pluginDirectory = join(
    run.repositoryRoot,
    "packages",
    "skill",
    `skill-${run.packageName}`,
  );
  const source = join(pluginDirectory, run.guide, "evals.json");
  const batches = outputEvalBatches(readFileSync(source, "utf8"));
  const batchDirectory = mkdtempSync(
    join(run.temporaryRoot, `${run.packageName}-output-`),
  );
  const workspace = join(
    run.repositoryRoot,
    ".skill-eval-results",
    run.packageName,
    "output",
  );

  for (const [index, batch] of batches.entries()) {
    const batchName = `batch-${String(index + 1)}`;
    const batchFile = join(batchDirectory, `${batchName}.json`);
    writeJson(batchFile, batch);
    assertEvaluatorSucceeded(
      "moon-skill-eval-outputs",
      runEvaluator(
        "moon-skill-eval-outputs",
        [
          batchFile,
          `xonovex-skill-${run.packageName}:${run.guide}`,
          batchName,
          "--runs",
          "1",
          "--concurrency",
          "2",
          "--model",
          "haiku",
          "--judge-model",
          "sonnet",
          "--max-budget-usd",
          "0.10",
          "--judge-max-budget-usd",
          "0.10",
          "--plugin-dir",
          pluginDirectory,
          "--eval-cwd",
          run.repositoryRoot,
          "--workspace",
          workspace,
        ],
        run.repositoryRoot,
        false,
      ),
    );
  }
};

export const runSkillEvals = (input: SkillEvalRun): void => {
  assertSkillSegment(input.packageName, "skill package");
  assertSkillSegment(input.guide, "skill guide");
  const run = {
    ...input,
    repositoryRoot: resolve(input.repositoryRoot),
    temporaryRoot: resolve(input.temporaryRoot),
  };
  mkdirSync(run.temporaryRoot, {recursive: true});
  if (run.mode === "trigger") {
    runTriggerEvals(run);
    return;
  }
  runOutputEvals(run);
};
