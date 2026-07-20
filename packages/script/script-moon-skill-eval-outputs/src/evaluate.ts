import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {runJob, type RunContext} from "./output-process.js";
import {aggregateArm, fmean, round, type JobRecord} from "./output-results.js";
import {
  buildGenerationClaudeArgs,
  evalEntries,
  findEvaluationInfrastructureFailures,
  MAX_OUTPUT_MODEL_CALLS,
  normalizeEval,
  outputModelCallCount,
  parseOutputOptions,
  runFailFastPool,
  type EvaluationArm,
  type NormalizedEval,
} from "./validation.js";

const which = (command: string): boolean => {
  try {
    resolveExecutable(command);
    return true;
  } catch {
    return false;
  }
};

// skillNameFromSkillMd resolves the declared skill name used by evaluator defaults.
const skillNameFromSkillMd = (path: string): string | undefined => {
  if (!isFile(path)) return undefined;
  const content = readFileSync(path, "utf8");
  const fmMatch = /^---\s*\n([\s\S]*?)\n---/.exec(content);
  const block = fmMatch?.[1] ?? content;
  const nameMatch = /^name:\s*(.+?)\s*$/m.exec(block);
  const name = nameMatch?.[1];
  if (!name) return undefined;
  return name.replaceAll(/^["']|["']$/g, "").trim() || undefined;
};

const loadEvaluations = (
  evaluationsFile: string,
  evaluationsArgument: string,
): readonly NormalizedEval[] | undefined => {
  let data: unknown;
  try {
    data = JSON.parse(readFileSync(evaluationsFile, "utf8"));
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(
      `Error: invalid JSON in ${evaluationsArgument}: ${message}\n`,
    );
    return undefined;
  }

  const entriesResult = evalEntries(data);
  if (!entriesResult.success) {
    process.stderr.write(
      `Error: invalid eval structure in ${evaluationsArgument}: ${entriesResult.error}\n`,
    );
    return undefined;
  }
  if (entriesResult.data.length === 0) {
    process.stderr.write(`Error: ${evaluationsArgument} has no evals\n`);
    return undefined;
  }

  const evaluations: NormalizedEval[] = [];
  for (const [index, raw] of entriesResult.data.entries()) {
    const result = normalizeEval(raw, index + 1);
    if (result.success) {
      evaluations.push(result.data);
    } else {
      process.stderr.write(
        `Skipping eval #${String(index + 1)}: ${result.error}\n`,
      );
    }
  }
  if (evaluations.length === 0) {
    process.stderr.write("Error: no gradable evals\n");
    return undefined;
  }
  return evaluations;
};

const nextIteration = (base: string): string => {
  if (!existsSync(base)) return "iteration-1";
  const existing = readdirSync(base).flatMap((entry) => {
    const match = /^iteration-(\d+)$/.exec(entry);
    return match !== null && isDirectory(join(base, entry))
      ? [Number(match[1])]
      : [];
  });
  return `iteration-${String(Math.max(0, ...existing) + 1)}`;
};

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
    process.stderr.write(
      `batch ${String(batchIndex + 1)}/${String(batches.length)}: ` +
        `${String(batch.length)} evals\n`,
    );
    const batchRecords = await runFailFastPool(
      evaluationJobs(batch, context.runs),
      concurrency,
      (job) => runJob(job.evaluation, job.arm, job.runIndex, context),
      (record) => findEvaluationInfrastructureFailures([record]).length > 0,
    );
    records.push(...batchRecords);

    const failures = findEvaluationInfrastructureFailures(batchRecords);
    if (failures.length > 0) {
      writeFileSync(
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
        {encoding: "utf8"},
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

export const main = async (argv: readonly string[]): Promise<number> => {
  const {values, positionals} = parseCliArgs(
    {
      name: "moon-skill-eval-outputs",
      description:
        "Run output-quality evals against a skill: with-skill vs without-skill.",
      options: {
        runs: {
          type: "string",
          description: "runs per arm per eval (env RUNS, default 1, maximum 3)",
        },
        concurrency: {
          type: "string",
          description:
            "parallel claude invocations (env CONCURRENCY, default/maximum 2)",
        },
        "batch-size": {
          type: "string",
          description: "evals per sequential batch",
        },
        model: {
          type: "string",
          description:
            "model for the generation runs (env CLAUDE_MODEL, default haiku)",
        },
        "judge-model": {
          type: "string",
          description: "model for grading (env JUDGE_MODEL)",
        },
        "disallowed-tools": {
          type: "string",
          description:
            "tools blocked in both arms (env DISALLOWED_TOOLS); without-skill also blocks Skill",
        },
        "gen-timeout": {
          type: "string",
          description:
            "per-generation timeout in seconds (env GEN_TIMEOUT, default 600)",
        },
        workspace: {
          type: "string",
          description:
            "workspace base dir (env WORKSPACE, default '<skill>-workspace')",
        },
        "eval-cwd": {
          type: "string",
          description:
            "working dir for generation runs (env EVAL_CWD, default current dir)",
        },
        "plugin-dir": {
          type: "string",
          description:
            "load the with-skill arm from a local plugin directory (env PLUGIN_DIR)",
        },
        "max-budget-usd": {
          type: "string",
          description:
            "hard per-generation spend cap up to 0.10 (env MAX_BUDGET_USD, default 0.10)",
        },
        "judge-max-budget-usd": {
          type: "string",
          description:
            "hard per-judge spend cap up to 0.10 (env JUDGE_MAX_BUDGET_USD, default 0.10)",
        },
      },
    },
    argv,
  );

  const guideDir = resolveGuideDirectory(resolve("."));
  const evalsArg = positionals[0] ?? join(guideDir, "evals.json");
  const evalsFile = resolve(evalsArg);

  if (!isFile(evalsFile)) {
    process.stderr.write(`Error: evals file not found: ${evalsArg}\n`);
    return 2;
  }
  if (!which("claude")) {
    process.stderr.write("Error: 'claude' CLI not found in PATH\n");
    return 2;
  }

  const skillArg = positionals[1];
  const skillMd = join(guideDir, "SKILL.md");
  const skillName =
    skillArg && skillArg.length > 0 ? skillArg : skillNameFromSkillMd(skillMd);
  if (!skillName) {
    process.stderr.write(
      `Error: no skill_name given and no 'name' frontmatter in ${skillMd}\n`,
    );
    return 2;
  }

  let iteration = positionals[2] ?? "";

  const norm = loadEvaluations(evalsFile, evalsArg);
  if (norm === undefined) {
    return 2;
  }

  const runsRaw =
    (values.runs as string | undefined) ?? process.env.RUNS ?? "1";
  const concurrencyRaw =
    (values.concurrency as string | undefined) ??
    process.env.CONCURRENCY ??
    "2";
  const claudeModel =
    (values.model as string | undefined) ?? process.env.CLAUDE_MODEL ?? "haiku";
  const judgeModel =
    (values["judge-model"] as string | undefined) ??
    process.env.JUDGE_MODEL ??
    "";
  const disallowed =
    (values["disallowed-tools"] as string | undefined) ??
    process.env.DISALLOWED_TOOLS ??
    "Bash,Edit,Write,NotebookEdit,WebFetch";
  const timeoutRaw =
    (values["gen-timeout"] as string | undefined) ??
    process.env.GEN_TIMEOUT ??
    "600";
  const workspaceArg =
    (values.workspace as string | undefined) ?? process.env.WORKSPACE ?? "";
  const cwdArg =
    (values["eval-cwd"] as string | undefined) ?? process.env.EVAL_CWD ?? "";
  const cwd = cwdArg || undefined;
  const pluginDirRaw =
    (values["plugin-dir"] as string | undefined) ??
    process.env.PLUGIN_DIR ??
    "";
  const pluginDir = pluginDirRaw ? resolve(pluginDirRaw) : undefined;
  if (pluginDir !== undefined && !isDirectory(pluginDir)) {
    process.stderr.write(
      `Error: local plugin directory not found: ${pluginDirRaw}\n`,
    );
    return 2;
  }
  const budgetRaw =
    (values["max-budget-usd"] as string | undefined) ??
    process.env.MAX_BUDGET_USD ??
    "0.10";
  const judgeBudgetRaw =
    (values["judge-max-budget-usd"] as string | undefined) ??
    process.env.JUDGE_MAX_BUDGET_USD ??
    "0.10";
  const batchSizeRaw = values["batch-size"] as string | undefined;
  const optionsResult = parseOutputOptions({
    runs: runsRaw,
    concurrency: concurrencyRaw,
    timeout: timeoutRaw,
    budget: budgetRaw,
    judgeBudget: judgeBudgetRaw,
    ...(batchSizeRaw === undefined ? {} : {batchSize: batchSizeRaw}),
    ...(iteration ? {iteration} : {}),
  });
  if (!optionsResult.success) {
    process.stderr.write(
      `Error: invalid evaluator options: ${optionsResult.error}\n`,
    );
    return 2;
  }
  const {runs, concurrency, timeout, batchSize} = optionsResult.data;
  const budget = optionsResult.data.budget;
  const judgeBudget = optionsResult.data.judgeBudget;
  iteration = optionsResult.data.iteration ?? "";

  const evalBatches = boundedBatches(
    norm,
    batchSize ?? Math.max(norm.length, 1),
  );
  const maxBatchModelCalls = Math.max(
    0,
    ...evalBatches.map((batch) => outputModelCallCount(batch.length, runs)),
  );
  if (maxBatchModelCalls > MAX_OUTPUT_MODEL_CALLS) {
    process.stderr.write(
      `Error: output eval batch would launch ${String(maxBatchModelCalls)} model calls; ` +
        `maximum is ${String(MAX_OUTPUT_MODEL_CALLS)}. Split the eval set into bounded batches.\n`,
    );
    return 2;
  }

  const short = skillName.split(":").pop() ?? skillName;

  const base = workspaceArg
    ? resolve(workspaceArg)
    : resolve(`${short}-workspace`);
  iteration ||= nextIteration(base);
  const iterDir = join(base, iteration);
  mkdirSync(iterDir, {recursive: true});
  const benchmarkPath = join(iterDir, "benchmark.json");
  const invalidRunPath = join(iterDir, "invalid-run.json");
  rmSync(benchmarkPath, {force: true});
  rmSync(invalidRunPath, {force: true});

  const withArgs = buildGenerationClaudeArgs({
    arm: "with_skill",
    model: claudeModel,
    budget,
    disallowedTools: disallowed,
    ...(pluginDir
      ? {pluginDirectories: resolveClaudePluginDirectories(pluginDir)}
      : {}),
  });
  const withoutArgs = buildGenerationClaudeArgs({
    arm: "without_skill",
    model: claudeModel,
    budget,
    disallowedTools: disallowed,
  });

  const evalsDir = dirname(evalsFile);

  const buildPrompt = (e: NormalizedEval): string => {
    let prompt = e.prompt;
    const files = e.files;
    if (files.length > 0) {
      const paths = files.map((f) => resolve(evalsDir, f));
      prompt +=
        "\n\nRelevant input files (read them as needed):\n" +
        paths.map((p) => `- ${p}`).join("\n");
    }
    return prompt;
  };

  const ctx: RunContext = {
    withArgs,
    withoutArgs,
    cwd,
    timeout,
    target: skillName,
    shortName: short,
    judgeModel,
    judgeBudget,
    iterationDirectory: iterDir,
    runs,
    buildPrompt,
  };

  process.stderr.write(
    `skill: ${skillName}  evals: ${String(norm.length)}  runs/arm: ${String(runs)}  ` +
      `concurrency: ${String(concurrency)}  workspace: ${iterDir}\n` +
      `gen model: ${claudeModel || "<default>"}  judge model: ${judgeModel || "<default>"}\n` +
      `caps: generation=$${String(budget)}/6 turns  judge=$${String(judgeBudget)}/1 turn  ` +
      `model_calls=${String(outputModelCallCount(norm.length, runs))}  ` +
      `batches=${String(evalBatches.length)}  ` +
      `max_batch_calls=${String(maxBatchModelCalls)}/${String(MAX_OUTPUT_MODEL_CALLS)}  retries=0\n---\n`,
  );

  const records = await runEvaluationBatches(
    evalBatches,
    concurrency,
    ctx,
    invalidRunPath,
    skillName,
    iteration,
  );
  if (records === undefined) {
    return 2;
  }
  writeEvaluationLines(norm, records);

  const withBlock = aggregateArm(records, "with_skill", runs);
  const withoutBlock = aggregateArm(records, "without_skill", runs);
  const benchmark = {
    skill: skillName,
    iteration,
    runs_per_arm: runs,
    eval_count: norm.length,
    batch_count: evalBatches.length,
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
  };
  writeFileSync(benchmarkPath, JSON.stringify(benchmark, null, 2), {
    encoding: "utf8",
  });

  const delta = benchmark.run_summary.delta;
  process.stderr.write(
    `---\nwith_skill pass_rate: ${String(withBlock.pass_rate.mean)}  ` +
      `(skill fired: ${String(withBlock.skill_trigger_rate?.mean)})  ` +
      `tokens: ${String(withBlock.tokens.mean)}\n` +
      `without_skill pass_rate: ${String(withoutBlock.pass_rate.mean)}  ` +
      `tokens: ${String(withoutBlock.tokens.mean)}\n` +
      `delta pass_rate: ${String(delta.pass_rate)}  tokens: ${String(delta.tokens)}\n` +
      `benchmark: ${benchmarkPath}\n`,
  );
  return delta.pass_rate > 0 ? 0 : 1;
};
