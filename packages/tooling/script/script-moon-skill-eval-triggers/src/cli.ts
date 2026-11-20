const PROGRAM = "moon-skill-eval-triggers";

const USAGE = `Usage: ${PROGRAM} [queries.json] [skill_name] [split] [options]
    queries.json = path to eval-queries.json (default: ./eval-queries.json)
    skill_name   = bare ("using-git") or plugin-namespaced ("plugin:using-git")
                   (default: the "name" frontmatter from ./SKILL.md)
    split        = train | validation | all   (default: all)

Options (flag overrides env):
    --runs N             / RUNS=N             runs per query (default: 3)
    --threshold F        / THRESHOLD=F        trigger-rate cutoff for a pass (default: 0.5)
    --harness H          / SKILL_EVAL_HARNESS harness: claude | codex (default: claude)
    --model M                                  harness model (defaults: Claude claude-haiku-5-5; Codex gpt-5.3-codex)
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

// Help and usage errors are results rather than side effects, so the entry point
// owns the output stream and the exit code.
export type CliParse =
  | {readonly kind: "parsed"; readonly cli: ParsedCli}
  | {readonly kind: "help"; readonly text: string}
  | {readonly kind: "usage"; readonly error: string};

/** The usage text and the error line a usage error prints to stderr. */
export const usageErrorText = (message: string): string =>
  `${USAGE}\n${PROGRAM}: error: ${message}`;

const usage = (error: string): CliParse => ({kind: "usage", error});

export const parseCli = (argv: readonly string[]): CliParse => {
  const positionals: string[] = [];
  const options: Record<string, string> = {};

  let index = 0;
  while (index < argv.length) {
    const arg = argv[index] ?? "";
    if (arg === "--help" || arg === "-h") {
      return {kind: "help", text: USAGE};
    }

    const equals = arg.indexOf("=");
    const flag =
      equals !== -1 && arg.startsWith("--") ? arg.slice(0, equals) : arg;
    const inline =
      equals !== -1 && arg.startsWith("--") ? arg.slice(equals + 1) : undefined;

    if (OPTION_FLAGS.has(flag)) {
      const value = inline ?? argv[index + 1];
      if (value === undefined) {
        return usage(`argument ${flag}: expected one argument`);
      }
      options[flag] = value;
      index += inline === undefined ? 2 : 1;
      continue;
    }
    if (arg !== "-" && arg.startsWith("-")) {
      return usage(`unrecognized arguments: ${arg}`);
    }
    positionals.push(arg);
    index += 1;
  }

  return {
    kind: "parsed",
    cli: {
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
    },
  };
};
