import {mkdirSync, rmSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {
  checkCodexTriggered,
  checkTriggered,
  TRIGGER_OUTPUT_LIMIT,
} from "@xonovex/script-moon-skill-eval-common/trigger-process";
import {
  MAX_TRIGGER_MODEL_RUNS,
  triggerModelRunCount,
} from "@xonovex/script-moon-skill-eval-common/validation";
import {usageError} from "./cli.js";
import {resolveTriggerConfig} from "./trigger-config.js";
import {runTriggerEvaluation} from "./trigger-evaluation.js";

export const main = async (argv: readonly string[]): Promise<number> => {
  const configResult = resolveTriggerConfig(argv);
  if (!configResult.success) {
    if (configResult.kind === "usage") {
      return usageError(configResult.error);
    }
    process.stderr.write(`Error: ${configResult.error}\n`);
    return 2;
  }
  const config = configResult.data;

  if (config.workspace !== undefined) {
    mkdirSync(config.workspace, {recursive: true});
    rmSync(join(config.workspace, "results.jsonl"), {force: true});
    rmSync(join(config.workspace, "summary.json"), {force: true});
    rmSync(join(config.workspace, "invalid-run.json"), {force: true});
  }

  const evaluation = await runTriggerEvaluation({
    queryBatches: config.queryBatches,
    runs: config.runs,
    threshold: config.threshold,
    skillName: config.skillName,
    workspace: config.workspace,
    check:
      config.harness === "claude"
        ? (query) =>
            checkTriggered(
              query,
              config.harnessArgs,
              config.skillName,
              config.shortName,
              config.harnessExecutable,
            )
        : (query) =>
            checkCodexTriggered({
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
    writeFileSync(
      join(config.workspace, "results.jsonl"),
      results.map((result) => JSON.stringify(result)).join("\n") + "\n",
      "utf8",
    );
    writeFileSync(
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
      "utf8",
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
