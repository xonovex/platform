import {spawnSync} from "node:child_process";
import {mkdirSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {parseCli, parseFrontmatterName, usageError} from "./cli.js";
import {runTriggerEvaluation} from "./trigger-evaluation.js";
import {TRIGGER_OUTPUT_LIMIT} from "./trigger-process.js";
import {
  buildTriggerClaudeArgs,
  MAX_TRIGGER_MODEL_RUNS,
  parseQueries,
  parseTriggerOptions,
  selectQueries,
  triggerModelRunCount,
} from "./validation.js";

export const main = async (argv: readonly string[]): Promise<number> => {
  const cli = parseCli(argv);

  if (cli.positionals.length > 3) {
    usageError(`unrecognized arguments: ${cli.positionals.slice(3).join(" ")}`);
  }

  const guideDir = resolveGuideDirectory(resolve("."));

  const queriesArg = cli.positionals[0] ?? join(guideDir, "eval-queries.json");
  const queriesFile = resolve(queriesArg);

  const resolveSkillName = (): string => {
    const fromArg = cli.positionals[1];
    if (fromArg !== undefined) {
      return fromArg;
    }
    const skillMd = join(guideDir, "SKILL.md");
    const fromFrontmatter = isFile(skillMd)
      ? parseFrontmatterName(skillMd)
      : undefined;
    if (fromFrontmatter === undefined) {
      return usageError(
        "the following arguments are required: skill_name (no SKILL.md with a name frontmatter found)",
      );
    }
    return fromFrontmatter;
  };
  const skillName = resolveSkillName();

  // split defaults to all; argparse validates against the choices.
  const positionalSplit = cli.positionals[2];
  if (
    positionalSplit !== undefined &&
    cli.split !== undefined &&
    cli.split !== positionalSplit
  ) {
    usageError("split must not be provided twice with different values");
  }
  const split = cli.split ?? positionalSplit ?? "all";
  if (split !== "train" && split !== "validation" && split !== "all") {
    usageError(
      `argument split: invalid choice: '${split}' (choose from 'train', 'validation', 'all')`,
    );
  }

  if (!isFile(queriesFile)) {
    process.stderr.write(`Error: queries file not found: ${queriesFile}\n`);
    return 2;
  }

  let claudeExecutable: string;
  try {
    claudeExecutable = resolveExecutable("claude");
  } catch {
    process.stderr.write("Error: 'claude' CLI not found in PATH\n");
    return 2;
  }
  const claudeProbe = spawnSync(claudeExecutable, ["--version"], {
    stdio: "ignore",
  });
  if (claudeProbe.error) {
    process.stderr.write("Error: 'claude' CLI not found in PATH\n");
    return 2;
  }

  const runsRaw = cli.runs ?? process.env.RUNS ?? "3";
  const thresholdRaw = cli.threshold ?? process.env.THRESHOLD ?? "0.5";
  const modelRaw = cli.model ?? process.env.CLAUDE_MODEL ?? "haiku";
  const claudeModel = modelRaw.trim().length > 0 ? modelRaw : "haiku";
  const budgetRaw = cli.maxBudget ?? process.env.MAX_BUDGET_USD ?? "0.05";
  const optionsResult = parseTriggerOptions({
    runs: runsRaw,
    threshold: thresholdRaw,
    budget: budgetRaw,
    ...(cli.batchSize === undefined ? {} : {batchSize: cli.batchSize}),
  });
  if (!optionsResult.success) {
    return usageError(`invalid evaluator options: ${optionsResult.error}`);
  }
  const {runs, threshold, budget, batchSize} = optionsResult.data;

  const short = skillName.split(":").pop() ?? skillName;
  const pluginDirRaw =
    cli.pluginDir ?? process.env.PLUGIN_DIR ?? dirname(guideDir);
  const pluginDirectory = resolve(pluginDirRaw);
  if (
    !isDirectory(pluginDirectory) ||
    !isFile(join(pluginDirectory, ".claude-plugin", "plugin.json"))
  ) {
    process.stderr.write(
      `Error: target plugin directory is invalid: ${pluginDirectory}\n`,
    );
    return 2;
  }
  const claudeArgs = buildTriggerClaudeArgs({
    model: claudeModel,
    budget,
    pluginDirectories: resolveClaudePluginDirectories(pluginDirectory),
  });

  let parsed: unknown;
  try {
    parsed = JSON.parse(readFileSync(queriesFile, "utf8"));
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    process.stderr.write(`Error: invalid JSON in ${queriesFile}: ${detail}\n`);
    return 2;
  }
  const queryResult = parseQueries(parsed);
  if (!queryResult.success) {
    process.stderr.write(
      `Error: invalid queries in ${queriesFile}: ${queryResult.error}\n`,
    );
    return 2;
  }

  const selectionResult = selectQueries(queryResult.data, split);
  if (!selectionResult.success) {
    process.stderr.write(`Error: ${selectionResult.error}\n`);
    return 2;
  }
  const queries = selectionResult.data;
  const queryBatches = boundedBatches(
    queries,
    batchSize ?? Math.max(queries.length, 1),
  );
  const maxBatchModelRuns = Math.max(
    0,
    ...queryBatches.map((batch) => triggerModelRunCount(batch.length, runs)),
  );
  if (maxBatchModelRuns > MAX_TRIGGER_MODEL_RUNS) {
    process.stderr.write(
      `Error: trigger eval batch would launch ${String(maxBatchModelRuns)} model runs; ` +
        `maximum is ${String(MAX_TRIGGER_MODEL_RUNS)}\n`,
    );
    return 2;
  }

  const workspace =
    cli.workspace === undefined ? undefined : resolve(cli.workspace);
  if (workspace !== undefined) {
    mkdirSync(workspace, {recursive: true});
    rmSync(join(workspace, "results.jsonl"), {force: true});
    rmSync(join(workspace, "summary.json"), {force: true});
    rmSync(join(workspace, "invalid-run.json"), {force: true});
  }

  const evaluation = await runTriggerEvaluation({
    queryBatches,
    runs,
    threshold,
    claudeArgs,
    skillName,
    shortName: short,
    claudeExecutable,
    workspace,
  });
  if (!evaluation.success) {
    return 2;
  }
  const {results, passed, failed, total} = evaluation;

  if (workspace !== undefined) {
    writeFileSync(
      join(workspace, "results.jsonl"),
      results.map((result) => JSON.stringify(result)).join("\n") + "\n",
      "utf8",
    );
    writeFileSync(
      join(workspace, "summary.json"),
      `${JSON.stringify(
        {
          skill: skillName,
          split,
          batches: queryBatches.length,
          queries: total,
          runs_per_query: runs,
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
  const modelLabel = claudeModel || "<default>";
  process.stderr.write(
    `skill: ${skillName}  split: ${split}  runs: ${String(runs)}  ` +
      `threshold: ${String(threshold)}  model: ${modelLabel}  ` +
      `budget/run: $${String(budget)}  tools: Skill  timeout: 60s  ` +
      `output-limit: ${String(TRIGGER_OUTPUT_LIMIT)} chars  ` +
      `batches: ${String(queryBatches.length)}  ` +
      `model-runs: ${String(triggerModelRunCount(queries.length, runs))}\n`,
  );
  process.stderr.write(
    `passed: ${String(passed)} / ${String(total)}   failed: ${String(failed)}\n`,
  );
  return failed === 0 ? 0 : 1;
};
