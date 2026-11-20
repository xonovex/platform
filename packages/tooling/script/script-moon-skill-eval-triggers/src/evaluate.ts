import {join} from "node:path";
import {
  clearEvidence,
  writeJsonEvidence,
  writeJsonLinesEvidence,
} from "@xonovex/script-moon-skill-eval-common/evidence";
import type {EvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import {
  EVAL_EXIT,
  reporter,
  verdictExit,
} from "@xonovex/script-moon-skill-eval-common/report";
import {TRIGGER_OUTPUT_LIMIT} from "@xonovex/script-moon-skill-eval-common/trigger";
import {checkTrigger} from "@xonovex/script-moon-skill-eval-common/trigger-check";
import {
  MAX_TRIGGER_MODEL_RUNS,
  triggerModelRunCount,
} from "@xonovex/script-moon-skill-eval-common/validation";
import {parseCli, usageErrorText} from "./cli.js";
import {harnessStarts} from "./harness-probe.js";
import {resolveTriggerConfig} from "./trigger-config.js";
import {runTriggerEvaluation} from "./trigger-evaluation.js";

export const main = async (
  argv: readonly string[],
  workingDirectory: string,
  ports: EvaluatorPorts,
): Promise<number> => {
  const report = reporter(ports.output);
  const parsed = parseCli(argv);
  if (parsed.kind === "help") {
    ports.output.stdout(parsed.text);
    return EVAL_EXIT.passed;
  }
  if (parsed.kind === "usage") {
    ports.output.stderr(usageErrorText(parsed.error));
    return EVAL_EXIT.invalid;
  }
  const configResult = resolveTriggerConfig(parsed.cli, {
    environment: ports.environment,
    workingDirectory,
    platform: ports.platform,
    fs: ports.fs,
  });
  if (!configResult.success) {
    if (configResult.kind === "usage") {
      ports.output.stderr(usageErrorText(configResult.error));
    } else {
      report.error(configResult.error);
    }
    return EVAL_EXIT.invalid;
  }
  const config = configResult.data;
  if (!(await harnessStarts(config.harnessExecutable, ports))) {
    report.error(`'${config.harness}' CLI not found in PATH`);
    return EVAL_EXIT.invalid;
  }

  const {workspace} = config;
  if (workspace !== undefined) {
    ports.fs.makeDirectory(workspace);
    clearEvidence(ports.scratch, [
      join(workspace, "results.jsonl"),
      join(workspace, "summary.json"),
      join(workspace, "invalid-run.json"),
    ]);
  }

  const evaluation = await runTriggerEvaluation({
    queryBatches: config.queryBatches,
    runs: config.runs,
    threshold: config.threshold,
    skillName: config.skillName,
    workspace,
    fs: ports.fs,
    report,
    check: (query) =>
      checkTrigger(
        config.harness,
        {
          executable: config.harnessExecutable,
          args: config.harnessArgs,
          query,
          target: config.skillName,
          shortName: config.shortName,
          candidates: [
            {
              skillDirectory: config.skillDirectory,
              shortName: config.shortName,
            },
          ],
        },
        ports,
      ),
  });
  if (!evaluation.success) {
    return EVAL_EXIT.invalid;
  }
  const {results, passed, failed, total} = evaluation;

  if (workspace !== undefined) {
    writeJsonLinesEvidence(ports.fs, join(workspace, "results.jsonl"), results);
    writeJsonEvidence(ports.fs, join(workspace, "summary.json"), {
      skill: config.skillName,
      harness: config.harness,
      model: config.model,
      split: config.split,
      batches: config.queryBatches.length,
      queries: total,
      runs_per_query: config.runs,
      passed,
      failed,
    });
  }

  const budgetSummary =
    config.harness === "claude"
      ? `budget/run: $${String(config.budget)}  `
      : "";
  report.info("---");
  report.info(
    `skill: ${config.skillName}  harness: ${config.harness}  split: ${config.split}  ` +
      `runs: ${String(config.runs)}  threshold: ${String(config.threshold)}  ` +
      `model: ${config.model}  ` +
      budgetSummary +
      `timeout: 60s  ` +
      `output-limit: ${String(TRIGGER_OUTPUT_LIMIT)} chars  ` +
      `batches: ${String(config.queryBatches.length)}  ` +
      `model-runs: ${String(triggerModelRunCount(config.queryCount, config.runs))}  ` +
      `max-batch-runs: ${String(config.maxBatchModelRuns)}/${String(MAX_TRIGGER_MODEL_RUNS)}  ` +
      `routing-deferred: ${String(config.deferredToRouting)}`,
  );
  report.info(
    `passed: ${String(passed)} / ${String(total)}   failed: ${String(failed)}`,
  );
  return verdictExit(failed === 0);
};
