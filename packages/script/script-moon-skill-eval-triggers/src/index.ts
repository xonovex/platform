#!/usr/bin/env node
import {spawn, spawnSync} from "node:child_process";
import {mkdirSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {createInterface} from "node:readline";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {
  buildTriggerClaudeArgs,
  MAX_TRIGGER_MODEL_RUNS,
  parseQueries,
  parseTriggerOptions,
  streamTextDeltaLength,
  triggerModelRunCount,
} from "./validation.js";

const PROG = "moon-skill-eval-triggers";

const USAGE = `Usage: ${PROG} [queries.json] [skill_name] [split] [options]
    queries.json = path to eval-queries.json (default: ./eval-queries.json)
    skill_name   = bare ("git-commit") or plugin-namespaced ("plugin:git-commit")
                   (default: the "name" frontmatter from ./SKILL.md)
    split        = train | validation | all   (default: all)

Options (flag overrides env):
    --runs N             / RUNS=N             runs per query (default: 3)
    --threshold F        / THRESHOLD=F        trigger-rate cutoff for a pass (default: 0.5)
    --model M            / CLAUDE_MODEL=M     model for \`claude --model\` (default: haiku)
    --split S                                  train | validation | all (default: all)
    --batch-size N                             queries per sequential batch
    --workspace PATH                           directory for JSONL and summary evidence
    --plugin-dir PATH    / PLUGIN_DIR=PATH     target-only local plugin directory
    --max-budget-usd N   / MAX_BUDGET_USD=N   hard per-run spend cap (default/max: 0.05)
    -h, --help                                show this help and exit`;

const TRIGGER_TIMEOUT_MS = 60_000;
const TRIGGER_OUTPUT_LIMIT = 2000;

interface ResultRecord {
  readonly query: string;
  readonly should_trigger: boolean;
  readonly triggers: number;
  readonly runs: number;
  readonly trigger_rate: number;
  readonly pass: boolean;
  readonly rationale: string;
}

interface TriggerOutcome {
  readonly triggered: boolean;
  readonly error: string | null;
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const usageError = (message: string): never => {
  process.stderr.write(`${USAGE}\n${PROG}: error: ${message}\n`);
  process.exit(2);
};

const matchSkill = (
  skillField: unknown,
  target: string,
  short: string,
): boolean => {
  if (typeof skillField !== "string") {
    return false;
  }
  return (
    skillField === target ||
    skillField === short ||
    skillField.endsWith(`:${short}`)
  );
};

const checkLine = (line: string, target: string, short: string): boolean => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return false;
  }
  if (!isRecord(obj)) {
    return false;
  }

  // Check message.content[].type === "tool_use" and .name === "Skill"
  const message = obj.message;
  if (isRecord(message)) {
    const content = Array.isArray(message.content) ? message.content : [];
    for (const item of content) {
      if (isRecord(item) && item.type === "tool_use" && item.name === "Skill") {
        const inputField = item.input;
        if (
          isRecord(inputField) &&
          matchSkill(inputField.skill, target, short)
        ) {
          return true;
        }
      }
    }
  }

  // Check permission_denials[].tool_name === "Skill"
  const denials = Array.isArray(obj.permission_denials)
    ? obj.permission_denials
    : [];
  for (const denial of denials) {
    if (isRecord(denial) && denial.tool_name === "Skill") {
      const toolInput = denial.tool_input;
      if (isRecord(toolInput) && matchSkill(toolInput.skill, target, short)) {
        return true;
      }
    }
  }

  return false;
};

const skillAvailableLine = (
  line: string,
  target: string,
  short: string,
): boolean => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return false;
  }
  if (!isRecord(obj) || obj.type !== "system" || obj.subtype !== "init") {
    return false;
  }
  const skills = Array.isArray(obj.skills) ? obj.skills : [];
  return skills.some((skill) => matchSkill(skill, target, short));
};

const textDeltaLength = (line: string): number => {
  try {
    return streamTextDeltaLength(JSON.parse(line));
  } catch {
    return 0;
  }
};

/**
 * Resolve the target trigger result or an infrastructure error.
 * Terminates the claude process on first match — no further tools fire.
 */
const checkTriggered = (
  query: string,
  claudeArgs: readonly string[],
  target: string,
  short: string,
  claudeExecutable: string,
): Promise<TriggerOutcome> =>
  new Promise((resolvePromise) => {
    const proc = spawn(claudeExecutable, [...claudeArgs, query], {
      stdio: ["ignore", "pipe", "pipe"],
    });

    let matched = false;
    let targetAvailable = false;
    let timedOut = false;
    let outputLimitExceeded = false;
    let outputCharacters = 0;
    let stderr = "";
    let spawnError: string | null = null;
    let settled = false;
    let killTimer: NodeJS.Timeout | undefined;

    const rl = createInterface({input: proc.stdout, crlfDelay: Infinity});

    const finish = (outcome: TriggerOutcome): void => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timeoutTimer);
      if (killTimer) {
        clearTimeout(killTimer);
      }
      resolvePromise(outcome);
    };

    const timeoutTimer = setTimeout(() => {
      timedOut = true;
      killProc();
    }, TRIGGER_TIMEOUT_MS);

    const killProc = (): void => {
      if (proc.exitCode === null && proc.signalCode === null) {
        // SIGKILL on POSIX; on Windows Node maps this to TerminateProcess.
        proc.kill("SIGKILL");
        // Safety net if the process refuses to die.
        killTimer = setTimeout(() => proc.kill("SIGKILL"), 5000);
      }
    };

    rl.on("line", (raw) => {
      if (matched) {
        return;
      }
      const line = raw.trim();
      if (!line) {
        return;
      }
      targetAvailable ||= skillAvailableLine(line, target, short);
      if (checkLine(line, target, short)) {
        matched = true;
        rl.close();
        killProc();
        return;
      }
      outputCharacters += textDeltaLength(line);
      if (outputCharacters > TRIGGER_OUTPUT_LIMIT) {
        outputLimitExceeded = true;
        rl.close();
        killProc();
      }
    });

    proc.stderr.setEncoding("utf8");
    proc.stderr.on("data", (chunk: string) => {
      stderr += chunk;
    });

    proc.on("error", (error) => {
      spawnError = error.message;
    });

    proc.on("close", (code) => {
      if (matched) {
        finish({triggered: true, error: null});
        return;
      }
      if (timedOut) {
        finish({triggered: false, error: "timeout"});
        return;
      }
      if (outputLimitExceeded) {
        finish({triggered: false, error: "output-limit"});
        return;
      }
      if (spawnError !== null) {
        finish({triggered: false, error: spawnError});
        return;
      }
      if (code !== 0) {
        const detail = stderr.trim();
        const detailSuffix = detail.length > 0 ? `: ${detail}` : "";
        finish({
          triggered: false,
          error: `claude exited ${String(code)}${detailSuffix}`,
        });
        return;
      }
      if (!targetAvailable) {
        finish({triggered: false, error: "target skill unavailable"});
        return;
      }
      finish({triggered: false, error: null});
    });
  });

const parseFrontmatterName = (skillMd: string): string | undefined => {
  const text = readFileSync(skillMd, "utf8");
  const match = /^---\r?\n([\s\S]*?)\r?\n---/.exec(text);
  const frontmatter = match?.[1];
  if (frontmatter === undefined) {
    return undefined;
  }
  for (const fmLine of frontmatter.split(/\r?\n/)) {
    const nameMatch = /^name:\s*(.+?)\s*$/.exec(fmLine);
    const name = nameMatch?.[1];
    if (name !== undefined) {
      return name.replaceAll(/^["']|["']$/g, "");
    }
  }
  return undefined;
};

interface ParsedCli {
  readonly positionals: readonly string[];
  readonly runs?: string;
  readonly threshold?: string;
  readonly model?: string;
  readonly split?: string;
  readonly batchSize?: string;
  readonly workspace?: string;
  readonly pluginDir?: string;
  readonly maxBudget?: string;
}

const OPTION_FLAGS = new Set([
  "--runs",
  "--threshold",
  "--model",
  "--split",
  "--batch-size",
  "--workspace",
  "--plugin-dir",
  "--max-budget-usd",
]);

const parseCli = (argv: readonly string[]): ParsedCli => {
  const positionals: string[] = [];
  let runs: string | undefined;
  let threshold: string | undefined;
  let model: string | undefined;
  let split: string | undefined;
  let batchSize: string | undefined;
  let workspace: string | undefined;
  let pluginDir: string | undefined;
  let maxBudget: string | undefined;

  const takeValue = (
    flag: string,
    inline: string | undefined,
    i: number,
  ): {value: string; next: number} => {
    if (inline !== undefined) {
      return {value: inline, next: i};
    }
    const value = argv[i + 1];
    if (value === undefined) {
      return usageError(`argument ${flag}: expected one argument`);
    }
    return {value, next: i + 1};
  };

  let i = 0;
  while (i < argv.length) {
    const arg = argv[i];
    if (arg === undefined) {
      i += 1;
      continue;
    }
    if (arg === "--help" || arg === "-h") {
      process.stdout.write(`${USAGE}\n`);
      process.exit(0);
    }

    const eq = arg.indexOf("=");
    const flag = arg.startsWith("--") && eq !== -1 ? arg.slice(0, eq) : arg;
    const inline =
      arg.startsWith("--") && eq !== -1 ? arg.slice(eq + 1) : undefined;

    if (OPTION_FLAGS.has(flag)) {
      const {value, next} = takeValue(flag, inline, i);
      i = next + 1;
      switch (flag) {
        case "--runs": {
          runs = value;
          break;
        }
        case "--threshold": {
          threshold = value;
          break;
        }
        case "--model": {
          model = value;
          break;
        }
        case "--split": {
          split = value;
          break;
        }
        case "--batch-size": {
          batchSize = value;
          break;
        }
        case "--workspace": {
          workspace = value;
          break;
        }
        case "--plugin-dir": {
          pluginDir = value;
          break;
        }
        case "--max-budget-usd": {
          maxBudget = value;
          break;
        }
      }
      continue;
    }

    if (arg.startsWith("-") && arg !== "-") {
      usageError(`unrecognized arguments: ${arg}`);
    }
    positionals.push(arg);
    i += 1;
  }

  return {
    positionals,
    runs,
    threshold,
    model,
    split,
    batchSize,
    workspace,
    pluginDir,
    maxBudget,
  };
};

const main = async (argv: readonly string[]): Promise<number> => {
  const cli = parseCli(argv);

  if (cli.positionals.length > 3) {
    usageError(`unrecognized arguments: ${cli.positionals.slice(3).join(" ")}`);
  }

  // ERGONOMIC ADDITIONS (defaults only; explicit positionals are untouched).
  // The simple moon task runs this bin from a skill package root whose SKILL.md
  // and eval-queries.json live one level down in a single guide subdir, so the
  // cwd-based defaults must descend into that guide dir first.
  const guideDir = resolveGuideDirectory(resolve("."));

  // queries defaults to <guideDir>/eval-queries.json
  const queriesArg = cli.positionals[0] ?? join(guideDir, "eval-queries.json");
  const queriesFile = resolve(queriesArg);

  // skill_name defaults to the "name" frontmatter parsed from <guideDir>/SKILL.md
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
    cli.split !== undefined &&
    positionalSplit !== undefined &&
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

  let queries = queryResult.data;
  if (split !== "all") {
    queries = queries.filter((query) => query.split === split);
  }
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

  let passed = 0;
  let failed = 0;
  let total = 0;
  const results: ResultRecord[] = [];

  for (const [batchIndex, batch] of queryBatches.entries()) {
    process.stderr.write(
      `batch ${String(batchIndex + 1)}/${String(queryBatches.length)}: ` +
        `${String(batch.length)} queries\n`,
    );
    for (const entry of batch) {
      const {query, rationale} = entry;
      const shouldTrigger = entry.should_trigger;

      let triggers = 0;
      for (let i = 0; i < runs; i += 1) {
        const outcome = await checkTriggered(
          query,
          claudeArgs,
          skillName,
          short,
          claudeExecutable,
        );
        if (outcome.error !== null) {
          if (workspace !== undefined) {
            writeFileSync(
              join(workspace, "invalid-run.json"),
              `${JSON.stringify({query, error: outcome.error}, null, 2)}\n`,
              "utf8",
            );
          }
          process.stderr.write(
            `Error: trigger infrastructure failure for query ${JSON.stringify(query)}: ${outcome.error}\n`,
          );
          return 2;
        }
        if (outcome.triggered) {
          triggers += 1;
        }
      }

      const rate = runs ? triggers / runs : 0;
      const triggeredMajority = rate >= threshold;
      const passes = triggeredMajority === shouldTrigger;

      total += 1;
      if (passes) {
        passed += 1;
      } else {
        failed += 1;
      }

      const result: ResultRecord = {
        query,
        should_trigger: shouldTrigger,
        triggers,
        runs,
        trigger_rate: Math.round(rate * 1000) / 1000,
        pass: passes,
        rationale,
      };
      results.push(result);
      process.stdout.write(`${JSON.stringify(result)}\n`);
    }
  }

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

try {
  process.exitCode = await main(process.argv.slice(2));
} catch (error: unknown) {
  const detail = error instanceof Error ? error.message : String(error);
  process.stderr.write(`Error: ${detail}\n`);
  process.exitCode = 2;
}
