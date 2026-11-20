import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {
  clearEvidence,
  writeJsonEvidence,
} from "@xonovex/script-moon-skill-eval-common/evidence";
import type {EvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import {
  EVAL_EXIT,
  reporter,
  verdictExit,
  type Reporter,
} from "@xonovex/script-moon-skill-eval-common/report";
import {
  runWithTransientRetry,
  TRANSIENT_RUN_ATTEMPTS,
  transientRetryNotice,
} from "@xonovex/script-moon-skill-eval-common/retry";
import {
  buildEvaluationPrompt,
  cliDefinition,
  resolveEvaluationConfig,
} from "./evaluation-config.js";
import {GENERATION_OUTPUT_LIMIT} from "./output-parse.js";
import {runJob, type RunContext} from "./output-process.js";
import {aggregateArm, fmean, round, type JobRecord} from "./output-results.js";
import {
  evaluateOutputGate,
  findEvaluationInfrastructureFailures,
  MAX_OUTPUT_MODEL_CALLS,
  outputModelCallCount,
  runFailFastPool,
  type EvaluationArm,
  type NormalizedEval,
} from "./validation.js";

interface EvaluationJob {
  readonly evaluation: NormalizedEval;
  readonly arm: EvaluationArm;
  readonly runIndex: number;
}

const evaluationJobs = (
  evaluations: readonly NormalizedEval[],
  runs: number,
): readonly EvaluationJob[] => {
  const arms = ["with_skill", "without_skill"] as const;
  return evaluations.flatMap((evaluation) =>
    arms.flatMap((arm) =>
      Array.from({length: runs}, (_, runIndex) => ({
        evaluation,
        arm,
        runIndex,
      })),
    ),
  );
};

const runEvaluationBatches = async (
  batches: readonly (readonly NormalizedEval[])[],
  concurrency: number,
  context: RunContext,
  invalidRunPath: string,
  skillName: string,
  iteration: string,
): Promise<readonly JobRecord[] | undefined> => {
  const records: JobRecord[] = [];
  for (const [batchIndex, batch] of batches.entries()) {
    context.report.info(
      `batch ${String(batchIndex + 1)}/${String(batches.length)}: ` +
        `${String(batch.length)} evals`,
    );
    const batchRecords = await runFailFastPool(
      evaluationJobs(batch, context.runs),
      concurrency,
      // A with-skill run whose skill never activated carries no error, so it is
      // not retried: that is a finding, not a flake.
      (job) =>
        runWithTransientRetry(
          () => runJob(job.evaluation, job.arm, job.runIndex, context),
          (attempt, error) => {
            context.report.info(transientRetryNotice(attempt, error));
          },
        ),
      (record) => findEvaluationInfrastructureFailures([record]).length > 0,
    );
    records.push(...batchRecords);

    const failures = findEvaluationInfrastructureFailures(batchRecords);
    if (failures.length > 0) {
      writeJsonEvidence(context.fs, invalidRunPath, {
        skill: skillName,
        iteration,
        status: "invalid",
        batch: batchIndex + 1,
        failures,
      });
      context.report.info(
        `invalid benchmark evidence: ${String(failures.length)} infrastructure failure(s)`,
      );
      context.report.info(`diagnostic: ${invalidRunPath}`);
      return undefined;
    }
  }
  return records;
};

const averageRecordValue = (
  records: readonly JobRecord[],
  select: (record: JobRecord) => number,
): number => (records.length > 0 ? fmean(records.map(select)) : 0);

const reportEvaluationLines = (
  evaluations: readonly NormalizedEval[],
  records: readonly JobRecord[],
  report: Reporter,
): void => {
  for (const evaluation of evaluations) {
    const withSkill = records.filter(
      (record) => record.id === evaluation.id && record.arm === "with_skill",
    );
    const withoutSkill = records.filter(
      (record) => record.id === evaluation.id && record.arm === "without_skill",
    );
    const withPassRate = averageRecordValue(
      withSkill,
      (record) => record.pass_rate,
    );
    const withoutPassRate = averageRecordValue(
      withoutSkill,
      (record) => record.pass_rate,
    );
    const withTokens = averageRecordValue(withSkill, (record) => record.tokens);
    const withoutTokens = averageRecordValue(
      withoutSkill,
      (record) => record.tokens,
    );
    report.result({
      id: evaluation.id,
      prompt: evaluation.prompt,
      with_skill: {
        pass_rate: round(withPassRate, 3),
        tokens: round(withTokens),
        skill_triggered: withSkill.some((record) => record.skill_triggered),
      },
      without_skill: {
        pass_rate: round(withoutPassRate, 3),
        tokens: round(withoutTokens),
      },
      delta_pass_rate: round(withPassRate - withoutPassRate, 3),
      delta_tokens: round(withTokens - withoutTokens),
    });
  }
};

/** The ports of an output evaluation: an evaluator's ports and the clock. */
export interface OutputEvaluatorPorts extends EvaluatorPorts {
  // When the clock is read, so a recorded run reports a fixed duration.
  readonly now: () => number;
}

export const main = async (
  argv: readonly string[],
  workingDirectory: string,
  ports: OutputEvaluatorPorts,
): Promise<number> => {
  const report = reporter(ports.output);
  const args = parseCliArgs(cliDefinition, argv);
  if (args.kind === "help") {
    ports.output.stdout(args.text);
    return EVAL_EXIT.passed;
  }
  const configResult = resolveEvaluationConfig(args, {
    environment: ports.environment,
    workingDirectory,
    platform: ports.platform,
    fs: ports.fs,
  });
  for (const warning of configResult.warnings) {
    report.info(warning);
  }
  if (!configResult.success) {
    report.error(configResult.error);
    return EVAL_EXIT.invalid;
  }
  const config = configResult.data;

  ports.fs.makeDirectory(config.iterationDirectory);
  clearEvidence(ports.scratch, [config.benchmarkPath, config.invalidRunPath]);

  const ctx: RunContext = {
    ...ports,
    report,
    executable: config.harnessExecutable,
    withArgs: config.withArgs,
    withoutArgs: config.withoutArgs,
    cwd: config.cwd,
    skillDirectory: config.skillDirectory,
    harness: config.harness,
    timeout: config.timeout,
    target: config.skillName,
    shortName: config.shortName,
    judgeModel: config.judgeModel,
    judgeBudget: config.judgeBudget,
    iterationDirectory: config.iterationDirectory,
    runs: config.runs,
    buildPrompt: (evaluation) =>
      buildEvaluationPrompt(config.evaluationsDirectory, evaluation),
  };

  const harnessCaps =
    config.harness === "claude"
      ? `generation=$${String(config.budget)}/${String(config.maxTurns)} turns  ` +
        `judge=$${String(config.judgeBudget)}/1 turn  `
      : `generation-timeout=${String(config.timeout)}s  ` +
        `output-limit=${String(GENERATION_OUTPUT_LIMIT)} chars  `;
  report.info(
    `skill: ${config.skillName}  evals: ${String(config.evaluations.length)}  ` +
      `runs/arm: ${String(config.runs)}  concurrency: ${String(config.concurrency)}  ` +
      `workspace: ${config.iterationDirectory}\n` +
      `harness: ${config.harness}  gen model: ${config.model}  ` +
      `judge model: ${config.judgeModel}\n` +
      `caps: ${harnessCaps}` +
      `model_calls=${String(outputModelCallCount(config.evaluations.length, config.runs))}  ` +
      `batches=${String(config.evaluationBatches.length)}  ` +
      `max_batch_calls=${String(config.maxBatchModelCalls)}/${String(MAX_OUTPUT_MODEL_CALLS)}  ` +
      `retries=${String(TRANSIENT_RUN_ATTEMPTS - 1)}\n---`,
  );

  const records = await runEvaluationBatches(
    config.evaluationBatches,
    config.concurrency,
    ctx,
    config.invalidRunPath,
    config.skillName,
    config.iteration,
  );
  if (records === undefined) {
    return EVAL_EXIT.invalid;
  }
  reportEvaluationLines(config.evaluations, records, report);

  const withBlock = aggregateArm(records, "with_skill", config.runs);
  const withoutBlock = aggregateArm(records, "without_skill", config.runs);
  const benchmark = {
    skill: config.skillName,
    tier: config.tier,
    harness: config.harness,
    model: config.model,
    judge_model: config.judgeModel,
    iteration: config.iteration,
    runs_per_arm: config.runs,
    eval_count: config.evaluations.length,
    batch_count: config.evaluationBatches.length,
    run_summary: {
      with_skill: withBlock,
      without_skill: withoutBlock,
      delta: {
        pass_rate: round(
          withBlock.pass_rate.mean - withoutBlock.pass_rate.mean,
          3,
        ),
        tokens: round(withBlock.tokens.mean - withoutBlock.tokens.mean),
        duration_ms: round(
          withBlock.duration_ms.mean - withoutBlock.duration_ms.mean,
        ),
      },
    },
    quality_gate: evaluateOutputGate(
      config.tier,
      withBlock.pass_rate.mean,
      withoutBlock.pass_rate.mean,
      withBlock.skill_trigger_rate?.mean ?? 0,
    ),
  };
  writeJsonEvidence(ports.fs, config.benchmarkPath, benchmark);

  const delta = benchmark.run_summary.delta;
  report.info(
    `---\nwith_skill pass_rate: ${String(withBlock.pass_rate.mean)}  ` +
      `(skill fired: ${String(withBlock.skill_trigger_rate?.mean)})  ` +
      `tokens: ${String(withBlock.tokens.mean)}\n` +
      `without_skill pass_rate: ${String(withoutBlock.pass_rate.mean)}  ` +
      `tokens: ${String(withoutBlock.tokens.mean)}\n` +
      `delta pass_rate: ${String(delta.pass_rate)}  tokens: ${String(delta.tokens)}\n` +
      `quality gate (${config.tier}): ${benchmark.quality_gate.passed ? "PASS" : "FAIL"}\n` +
      `benchmark: ${config.benchmarkPath}`,
  );
  return verdictExit(benchmark.quality_gate.passed);
};
