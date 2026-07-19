#!/usr/bin/env node
import {spawn, spawnSync} from "node:child_process";
import {readFileSync} from "node:fs";
import {join, resolve} from "node:path";
import {createInterface} from "node:readline";
import {isFile, resolveGuideDirectory} from "@xonovex/script-moon-common/fs";
import {parseQueries, parseTriggerOptions} from "./validation.js";

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
    --disallowed-tools L / DISALLOWED_TOOLS=L tools blocked during the eval
    --max-budget-usd N   / MAX_BUDGET_USD=N   hard per-run spend cap (default: 0.10; 0 disables)
    -h, --help                                show this help and exit`;

const DEFAULT_DISALLOWED =
  "Bash,Edit,Write,NotebookEdit,WebFetch,WebSearch,Read,Glob,Grep,Task,TodoWrite";

interface ResultRecord {
  readonly query: string;
  readonly should_trigger: boolean;
  readonly triggers: number;
  readonly runs: number;
  readonly trigger_rate: number;
  readonly pass: boolean;
  readonly rationale: string;
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

/**
 * Resolve true if a matching Skill call appears in the claude stream.
 * Terminates the claude process on first match — no further tools fire.
 */
const checkTriggered = (
  query: string,
  claudeArgs: readonly string[],
  target: string,
  short: string,
): Promise<boolean> =>
  new Promise((resolvePromise) => {
    // eslint-disable-next-line sonarjs/no-os-command-from-path
    const proc = spawn("claude", [...claudeArgs, query], {
      stdio: ["ignore", "pipe", "ignore"],
    });

    let matched = false;
    let settled = false;
    let killTimer: NodeJS.Timeout | undefined;

    const rl = createInterface({input: proc.stdout, crlfDelay: Infinity});

    const finish = (): void => {
      if (settled) {
        return;
      }
      settled = true;
      if (killTimer) {
        clearTimeout(killTimer);
      }
      resolvePromise(matched);
    };

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
      if (checkLine(line, target, short)) {
        matched = true;
        rl.close();
        killProc();
      }
    });

    rl.on("close", () => {
      // If we matched, the process is being killed; wait for its exit below.
      if (!matched) {
        finish();
      }
    });

    proc.on("error", () => {
      finish();
    });

    proc.on("close", () => {
      finish();
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
  readonly disallowed?: string;
  readonly maxBudget?: string;
}

const OPTION_FLAGS = new Set([
  "--runs",
  "--threshold",
  "--model",
  "--disallowed-tools",
  "--max-budget-usd",
]);

const parseCli = (argv: readonly string[]): ParsedCli => {
  const positionals: string[] = [];
  let runs: string | undefined;
  let threshold: string | undefined;
  let model: string | undefined;
  let disallowed: string | undefined;
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
        case "--disallowed-tools": {
          disallowed = value;
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

  return {positionals, runs, threshold, model, disallowed, maxBudget};
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
  const split = cli.positionals[2] ?? "all";
  if (split !== "train" && split !== "validation" && split !== "all") {
    usageError(
      `argument split: invalid choice: '${split}' (choose from 'train', 'validation', 'all')`,
    );
  }

  if (!isFile(queriesFile)) {
    process.stderr.write(`Error: queries file not found: ${queriesFile}\n`);
    return 2;
  }

  // shutil.which("claude") equivalent — verify claude is available on PATH.
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const claudeProbe = spawnSync("claude", ["--version"], {stdio: "ignore"});
  if (claudeProbe.error) {
    process.stderr.write("Error: 'claude' CLI not found in PATH\n");
    return 2;
  }

  const runsRaw = cli.runs ?? process.env.RUNS ?? "3";
  const thresholdRaw = cli.threshold ?? process.env.THRESHOLD ?? "0.5";
  const claudeModel = cli.model ?? process.env.CLAUDE_MODEL ?? "haiku";
  const disallowed =
    cli.disallowed ?? process.env.DISALLOWED_TOOLS ?? DEFAULT_DISALLOWED;
  const budgetRaw = cli.maxBudget ?? process.env.MAX_BUDGET_USD ?? "0.10";
  const optionsResult = parseTriggerOptions({
    runs: runsRaw,
    threshold: thresholdRaw,
    budget: budgetRaw,
  });
  if (!optionsResult.success) {
    return usageError(`invalid evaluator options: ${optionsResult.error}`);
  }
  const {runs, threshold, budget} = optionsResult.data;

  const short = skillName.split(":").pop() ?? skillName;

  const claudeArgs: string[] = [
    "-p",
    "--output-format",
    "stream-json",
    "--verbose",
  ];
  if (claudeModel) {
    claudeArgs.push("--model", claudeModel);
  }
  if (disallowed) {
    // Use --opt=val to avoid the variadic parser swallowing the prompt.
    claudeArgs.push(`--disallowedTools=${disallowed}`);
  }
  if (budget && budget > 0) {
    // Hard per-run ceiling so a non-triggering run can't execute the whole task.
    claudeArgs.push("--max-budget-usd", String(budget));
  }

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

  let passed = 0;
  let failed = 0;
  let total = 0;

  for (const entry of queries) {
    const {query, rationale} = entry;
    const shouldTrigger = entry.should_trigger;

    let triggers = 0;
    for (let i = 0; i < runs; i++) {
      if (await checkTriggered(query, claudeArgs, skillName, short)) {
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
    process.stdout.write(`${JSON.stringify(result)}\n`);
  }

  process.stderr.write("---\n");
  const modelLabel = claudeModel || "<default>";
  const budgetLabel = budget && budget > 0 ? `$${String(budget)}` : "none";
  process.stderr.write(
    `skill: ${skillName}  split: ${split}  runs: ${String(runs)}  ` +
      `threshold: ${String(threshold)}  model: ${modelLabel}  ` +
      `budget/run: ${budgetLabel}  disallowed: ${disallowed}\n`,
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
