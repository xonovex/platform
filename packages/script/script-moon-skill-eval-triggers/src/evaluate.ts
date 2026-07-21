import {mkdirSync, rmSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {usageError} from "./cli.js";
import {resolveTriggerConfig} from "./trigger-config.js";
import {runTriggerEvaluation} from "./trigger-evaluation.js";
import {TRIGGER_OUTPUT_LIMIT} from "./trigger-process.js";
import {MAX_TRIGGER_MODEL_RUNS, triggerModelRunCount} from "./validation.js";

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
    claudeArgs: config.claudeArgs,
    skillName: config.skillName,
    shortName: config.shortName,
    claudeExecutable: config.claudeExecutable,
    workspace: config.workspace,
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
  process.stderr.write(
    `skill: ${config.skillName}  split: ${config.split}  runs: ${String(config.runs)}  ` +
      `threshold: ${String(config.threshold)}  model: ${config.claudeModel}  ` +
      `budget/run: $${String(config.budget)}  tools: Skill  timeout: 60s  ` +
      `output-limit: ${String(TRIGGER_OUTPUT_LIMIT)} chars  ` +
      `batches: ${String(config.queryBatches.length)}  ` +
      `model-runs: ${String(triggerModelRunCount(config.queryCount, config.runs))}  ` +
      `max-batch-runs: ${String(config.maxBatchModelRuns)}/${String(MAX_TRIGGER_MODEL_RUNS)}\n`,
  );
  process.stderr.write(
    `passed: ${String(passed)} / ${String(total)}   failed: ${String(failed)}\n`,
  );
  return failed === 0 ? 0 : 1;
};
