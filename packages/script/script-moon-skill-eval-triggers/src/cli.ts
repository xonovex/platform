const PROGRAM = "moon-skill-eval-triggers";

const USAGE = `Usage: ${PROGRAM} [queries.json] [skill_name] [split] [options]
    queries.json = path to eval-queries.json (default: ./eval-queries.json)
    skill_name   = bare ("git-commit") or plugin-namespaced ("plugin:git-commit")
                   (default: the "name" frontmatter from ./SKILL.md)
    split        = train | validation | all   (default: all)

Options (flag overrides env):
    --runs N             / RUNS=N             runs per query (default: 3)
    --threshold F        / THRESHOLD=F        trigger-rate cutoff for a pass (default: 0.5)
    --harness H          / SKILL_EVAL_HARNESS harness: claude | codex (default: claude)
    --model M                                  harness model (defaults: Claude claude-haiku-4-5-20251001; Codex gpt-5.3-codex)
    --split S                                  train | validation | all (default: all)
    --batch-size N                             queries per sequential batch
    --workspace PATH                           directory for JSONL and summary evidence
    --plugin-dir PATH    / PLUGIN_DIR=PATH     target-only local plugin directory
    --catalog-root DIR   / CATALOG_ROOT=DIR   skill catalog root; queries another skill in it
                                              owns are left to the routing evaluator
    --max-budget-usd N   / MAX_BUDGET_USD=N   hard per-run spend cap (default/max: 0.05)
    -h, --help                                show this help and exit`;

export interface ParsedCli {
  readonly positionals: readonly string[];
  readonly runs?: string;
  readonly threshold?: string;
  readonly harness?: string;
  readonly model?: string;
  readonly split?: string;
  readonly batchSize?: string;
  readonly workspace?: string;
  readonly pluginDir?: string;
  readonly catalogRoot?: string;
  readonly maxBudget?: string;
}

const OPTION_FLAGS = new Set([
  "--runs",
  "--threshold",
  "--harness",
  "--model",
  "--split",
  "--batch-size",
  "--workspace",
  "--plugin-dir",
  "--catalog-root",
  "--max-budget-usd",
]);

export const usageError = (message: string): never => {
  process.stderr.write(`${USAGE}\n${PROGRAM}: error: ${message}\n`);
  process.exit(2);
};

export const parseCli = (argv: readonly string[]): ParsedCli => {
  const positionals: string[] = [];
  const options: Record<string, string> = {};

  const takeValue = (
    flag: string,
    inline: string | undefined,
    index: number,
  ): {value: string; next: number} => {
    if (inline !== undefined) {
      return {value: inline, next: index};
    }
    const value = argv[index + 1];
    if (value === undefined) {
      return usageError(`argument ${flag}: expected one argument`);
    }
    return {value, next: index + 1};
  };

  let index = 0;
  while (index < argv.length) {
    const arg = argv[index];
    if (arg === undefined) {
      index += 1;
      continue;
    }
    if (arg === "--help" || arg === "-h") {
      process.stdout.write(`${USAGE}\n`);
      process.exit(0);
    }

    const equals = arg.indexOf("=");
    const flag =
      equals !== -1 && arg.startsWith("--") ? arg.slice(0, equals) : arg;
    const inline =
      equals !== -1 && arg.startsWith("--") ? arg.slice(equals + 1) : undefined;

    if (OPTION_FLAGS.has(flag)) {
      const {value, next} = takeValue(flag, inline, index);
      options[flag] = value;
      index = next + 1;
      continue;
    }
    if (arg !== "-" && arg.startsWith("-")) {
      usageError(`unrecognized arguments: ${arg}`);
    }
    positionals.push(arg);
    index += 1;
  }

  return {
    positionals,
    runs: options["--runs"],
    threshold: options["--threshold"],
    harness: options["--harness"],
    model: options["--model"],
    split: options["--split"],
    batchSize: options["--batch-size"],
    workspace: options["--workspace"],
    pluginDir: options["--plugin-dir"],
    catalogRoot: options["--catalog-root"],
    maxBudget: options["--max-budget-usd"],
  };
};
