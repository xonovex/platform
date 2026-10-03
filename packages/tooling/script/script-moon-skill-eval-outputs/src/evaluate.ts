import {rmSync} from "node:fs";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {
  buildEvaluationPrompt,
  resolveEvaluationConfig,
} from "./evaluation-config.js";
import {spawnHarness, type HarnessRunner} from "./output-harness.js";
import {GENERATION_OUTPUT_LIMIT} from "./output-parse.js";
import {runJob, type RunContext} from "./output-process.js";
import {aggregateArm, fmean, round, type JobRecord} from "./output-results.js";
import {
  evaluateOutputGate,
  findEvaluationInfrastructureFailures,
  MAX_OUTPUT_MODEL_CALLS,
  OUTPUT_RUN_ATTEMPTS,
  outputModelCallCount,
  runFailFastPool,
  runWithTransientRetry,
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

export const writeRetryNotice = (attempt: number, error: string): void => {
  process.stderr.write(
    `retrying after transient failure (${String(attempt)}/${String(OUTPUT_RUN_ATTEMPTS)}): ${error}\n`,
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
    process.stderr.write(
      `batch ${String(batchIndex + 1)}/${String(batches.length)}: ` +
        `${String(batch.length)} evals\n`,
    );
    const batchRecords = await runFailFastPool(
      evaluationJobs(batch, context.runs),
      concurrency,
      (job) =>
        runWithTransientRetry(
          () => runJob(job.evaluation, job.arm, job.runIndex, context),
          OUTPUT_RUN_ATTEMPTS,
          writeRetryNotice,
        ),
      (record) => findEvaluationInfrastructureFailures([record]).length > 0,
    );
    records.push(...batchRecords);

    const failures = findEvaluationInfrastructureFailures(batchRecords);
    if (failures.length > 0) {
      context.fs.writeFile(
        invalidRunPath,
        JSON.stringify(
          {
            skill: skillName,
            iteration,
            status: "invalid",
            batch: batchIndex + 1,
            failures,
          },
          null,
          2,
        ),
      );
      process.stderr.write(
        `invalid benchmark evidence: ${String(failures.length)} infrastructure failure(s)\n` +
          `diagnostic: ${invalidRunPath}\n`,
      );
      return undefined;
    }
  }
  return records;
};

const averageRecordValue = (
  records: readonly JobRecord[],
  select: (record: JobRecord) => number,
): number => (records.length > 0 ? fmean(records.map(select)) : 0);

const writeEvaluationLines = (
  evaluations: readonly NormalizedEval[],
  records: readonly JobRecord[],
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
    process.stdout.write(
      `${JSON.stringify({
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
      })}\n`,
    );
  }
};

/**
 * The effects the evaluator reaches the outside world through. A run supplies the
 * real ones; a test supplies recorded harness output and a fixed clock, so the
 * whole sweep can be scored without a process.
 */
export interface EvaluatorDependencies {
  readonly runHarness: HarnessRunner;
  readonly now: () => number;
  readonly fs: FileSystem;
  // Removes a previous run's verdicts so a rerun cannot be read as this run's.
  readonly discard: (path: string) => void;
  // Where relative paths and the guide are resolved from. A run uses the process
  // directory; a test names a directory in its own tree.
  readonly workingDirectory?: string;
  // Whether the harness CLI can be run. A run probes PATH; a test answers from
  // its own tree, so the verdict does not depend on what the host has installed.
  readonly executableAvailable?: (command: string) => boolean;
  // The variables the harness and budget defaults are read from. A run uses the
  // process environment; a test names its own.
  readonly environment?: Readonly<Record<string, string | undefined>>;
}

export const defaultDependencies: EvaluatorDependencies = {
  runHarness: spawnHarness,
  now: () => Date.now(),
  fs: nodeFileSystem,
  discard: (path) => {
    rmSync(path, {force: true});
  },
};

export const main = async (
  argv: readonly string[],
  dependencies: EvaluatorDependencies = defaultDependencies,
): Promise<number> => {
  const configResult = resolveEvaluationConfig(argv, {
    fs: dependencies.fs,
    workingDirectory: dependencies.workingDirectory,
    executableAvailable: dependencies.executableAvailable,
    environment: dependencies.environment,
  });
  for (const warning of configResult.warnings) {
    process.stderr.write(`${warning}\n`);
  }
  if (!configResult.success) {
    process.stderr.write(`Error: ${configResult.error}\n`);
    return 2;
  }
  const config = configResult.data;

  dependencies.fs.makeDirectory(config.iterationDirectory);
  dependencies.discard(config.benchmarkPath);
  dependencies.discard(config.invalidRunPath);

  const ctx: RunContext = {
    withArgs: config.withArgs,
    withoutArgs: config.withoutArgs,
    cwd: config.cwd,
    guideDirectory: config.guideDirectory,
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
    runHarness: dependencies.runHarness,
    now: dependencies.now,
    fs: dependencies.fs,
  };

  const harnessCaps =
    config.harness === "claude"
      ? `generation=$${String(config.budget)}/${String(config.maxTurns)} turns  ` +
        `judge=$${String(config.judgeBudget)}/1 turn  `
      : `generation-timeout=${String(config.timeout)}s  ` +
        `output-limit=${String(GENERATION_OUTPUT_LIMIT)} chars  `;
  process.stderr.write(
    `skill: ${config.skillName}  evals: ${String(config.evaluations.length)}  ` +
      `runs/arm: ${String(config.runs)}  concurrency: ${String(config.concurrency)}  ` +
      `workspace: ${config.iterationDirectory}\n` +
      `harness: ${config.harness}  gen model: ${config.model}  ` +
      `judge model: ${config.judgeModel}\n` +
      `caps: ${harnessCaps}` +
      `model_calls=${String(outputModelCallCount(config.evaluations.length, config.runs))}  ` +
      `batches=${String(config.evaluationBatches.length)}  ` +
      `max_batch_calls=${String(config.maxBatchModelCalls)}/${String(MAX_OUTPUT_MODEL_CALLS)}  ` +
      `retries=${String(OUTPUT_RUN_ATTEMPTS - 1)}\n---\n`,
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
    return 2;
  }
  writeEvaluationLines(config.evaluations, records);

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
  dependencies.fs.writeFile(
    config.benchmarkPath,
    JSON.stringify(benchmark, null, 2),
  );

  const delta = benchmark.run_summary.delta;
  process.stderr.write(
    `---\nwith_skill pass_rate: ${String(withBlock.pass_rate.mean)}  ` +
      `(skill fired: ${String(withBlock.skill_trigger_rate?.mean)})  ` +
      `tokens: ${String(withBlock.tokens.mean)}\n` +
      `without_skill pass_rate: ${String(withoutBlock.pass_rate.mean)}  ` +
      `tokens: ${String(withoutBlock.tokens.mean)}\n` +
      `delta pass_rate: ${String(delta.pass_rate)}  tokens: ${String(delta.tokens)}\n` +
      `quality gate (${config.tier}): ${benchmark.quality_gate.passed ? "PASS" : "FAIL"}\n` +
      `benchmark: ${config.benchmarkPath}\n`,
  );
  return benchmark.quality_gate.passed ? 0 : 1;
};
