import {rmSync} from "node:fs";
import {join} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {
  checkCodexTriggered,
  checkTriggered,
} from "@xonovex/script-moon-skill-eval-common/trigger-process";
import {TRIGGER_OUTPUT_LIMIT} from "@xonovex/script-moon-skill-eval-common/trigger-scan";
import {
  MAX_TRIGGER_MODEL_RUNS,
  triggerModelRunCount,
} from "@xonovex/script-moon-skill-eval-common/validation";
import {usageError} from "./cli.js";
import {
  resolveTriggerConfig,
  type TriggerConfigOptions,
} from "./trigger-config.js";
import {runTriggerEvaluation} from "./trigger-evaluation.js";

/**
 * The effects the trigger evaluator reaches a harness through. A run supplies the
 * spawning implementations; a test supplies decisions directly, so a whole sweep is
 * scored without a process.
 */
export interface TriggerDependencies {
  readonly checkTriggered: typeof checkTriggered;
  readonly checkCodexTriggered: typeof checkCodexTriggered;
  // How the configuration step probes that the harness binary exists and runs.
  readonly configOptions: TriggerConfigOptions;
  readonly fs: FileSystem;
  // Removes a previous run's evidence so a rerun cannot be read as this run's.
  readonly discard: (path: string) => void;
}

export const defaultDependencies: TriggerDependencies = {
  checkTriggered,
  checkCodexTriggered,
  configOptions: {},
  fs: nodeFileSystem,
  discard: (path) => {
    rmSync(path, {force: true});
  },
};

export const main = async (
  argv: readonly string[],
  dependencies: TriggerDependencies = defaultDependencies,
): Promise<number> => {
  const configResult = resolveTriggerConfig(argv, {
    ...dependencies.configOptions,
    fs: dependencies.configOptions.fs ?? dependencies.fs,
  });
  if (!configResult.success) {
    if (configResult.kind === "usage") {
      return usageError(configResult.error);
    }
    process.stderr.write(`Error: ${configResult.error}\n`);
    return 2;
  }
  const config = configResult.data;

  if (config.workspace !== undefined) {
    dependencies.fs.makeDirectory(config.workspace);
    dependencies.discard(join(config.workspace, "results.jsonl"));
    dependencies.discard(join(config.workspace, "summary.json"));
    dependencies.discard(join(config.workspace, "invalid-run.json"));
  }

  const evaluation = await runTriggerEvaluation({
    queryBatches: config.queryBatches,
    runs: config.runs,
    threshold: config.threshold,
    skillName: config.skillName,
    workspace: config.workspace,
    fs: dependencies.fs,
    check:
      config.harness === "claude"
        ? (query) =>
            dependencies.checkTriggered(
              query,
              config.harnessArgs,
              config.skillName,
              config.shortName,
              config.harnessExecutable,
            )
        : (query) =>
            dependencies.checkCodexTriggered({
              args: config.harnessArgs,
              executable: config.harnessExecutable,
              guideDirectory: config.guideDirectory,
              query,
              shortName: config.shortName,
            }),
  });
  if (!evaluation.success) {
    return 2;
  }
  const {results, passed, failed, total} = evaluation;

  if (config.workspace !== undefined) {
    dependencies.fs.writeFile(
      join(config.workspace, "results.jsonl"),
      results.map((result) => JSON.stringify(result)).join("\n") + "\n",
    );
    dependencies.fs.writeFile(
      join(config.workspace, "summary.json"),
      `${JSON.stringify(
        {
          skill: config.skillName,
          harness: config.harness,
          model: config.model,
          split: config.split,
          batches: config.queryBatches.length,
          queries: total,
          runs_per_query: config.runs,
          passed,
          failed,
        },
        null,
        2,
      )}\n`,
    );
  }

  process.stderr.write("---\n");
  const budgetSummary =
    config.harness === "claude"
      ? `budget/run: $${String(config.budget)}  `
      : "";
  process.stderr.write(
    `skill: ${config.skillName}  harness: ${config.harness}  split: ${config.split}  ` +
      `runs: ${String(config.runs)}  threshold: ${String(config.threshold)}  ` +
      `model: ${config.model}  ` +
      budgetSummary +
      `timeout: 60s  ` +
      `output-limit: ${String(TRIGGER_OUTPUT_LIMIT)} chars  ` +
      `batches: ${String(config.queryBatches.length)}  ` +
      `model-runs: ${String(triggerModelRunCount(config.queryCount, config.runs))}  ` +
      `max-batch-runs: ${String(config.maxBatchModelRuns)}/${String(MAX_TRIGGER_MODEL_RUNS)}  ` +
      `routing-deferred: ${String(config.deferredToRouting)}\n`,
  );
  process.stderr.write(
    `passed: ${String(passed)} / ${String(total)}   failed: ${String(failed)}\n`,
  );
  return failed === 0 ? 0 : 1;
};
