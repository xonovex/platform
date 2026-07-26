import {spawnSync} from "node:child_process";
import {readFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {skillEvalModelDefaults} from "@xonovex/script-moon-common/skill-eval-models";
import {parseCli, parseFrontmatterName} from "./cli.js";
import {catalogQueryOwners} from "./routing-catalog.js";
import {
  buildTriggerClaudeArgs,
  buildTriggerCodexArgs,
  MAX_TRIGGER_MODEL_RUNS,
  parseQueries,
  parseQuerySplit,
  parseTriggerOptions,
  selectQueries,
  triggerModelRunCount,
  type Query,
} from "./validation.js";

type TriggerConfigResult =
  | {readonly success: true; readonly data: TriggerConfig}
  | {
      readonly success: false;
      readonly error: string;
      readonly kind: "runtime" | "usage";
    };

interface TriggerConfigOptions {
  readonly environment?: Readonly<Record<string, string | undefined>>;
  readonly executableRuns?: (executable: string) => boolean;
  readonly resolveExecutablePath?: (command: string) => string | undefined;
  readonly workingDirectory?: string;
}

export interface TriggerConfig {
  readonly budget: number;
  readonly guideDirectory: string;
  readonly harness: "claude" | "codex";
  readonly harnessArgs: readonly string[];
  readonly harnessExecutable: string;
  readonly maxBatchModelRuns: number;
  readonly model: string;
  readonly queryBatches: readonly (readonly Query[])[];
  readonly deferredToRouting: number;
  readonly queryCount: number;
  readonly runs: number;
  readonly shortName: string;
  readonly skillName: string;
  readonly split: "train" | "validation" | "all";
  readonly threshold: number;
  readonly workspace: string | undefined;
}

const defaultResolveExecutablePath = (command: string): string | undefined => {
  try {
    return resolveExecutable(command);
  } catch {
    return undefined;
  }
};

const defaultExecutableRuns = (executable: string): boolean => {
  const probe = spawnSync(executable, ["--version"], {stdio: "ignore"});
  return probe.error === undefined;
};

const failure = (
  kind: "runtime" | "usage",
  error: string,
): TriggerConfigResult => ({success: false, error, kind});

type PartitionResult =
  | {
      readonly success: true;
      readonly kept: readonly Query[];
      readonly deferred: readonly Query[];
    }
  | {readonly success: false; readonly error: string};

// This evaluator loads one skill, so it can only answer whether that skill claims a
// query when it is the sole candidate. A near miss the catalog assigns to another
// skill asks a different question — which of the two wins with both loaded — and the
// routing evaluator answers it. Scoring such a query here reports a failure for
// correct behaviour, because the rightful owner was never on offer. Without a catalog
// root every query is kept, so a standalone run still evaluates the whole file.
const partitionOwnedQueries = (
  queries: readonly Query[],
  shortName: string,
  catalogRootArgument: string | undefined,
  workingDirectory: string,
): PartitionResult => {
  if (catalogRootArgument === undefined) {
    return {success: true, kept: queries, deferred: []};
  }
  let owners: ReadonlyMap<string, string>;
  try {
    owners = catalogQueryOwners(resolve(workingDirectory, catalogRootArgument));
  } catch (error: unknown) {
    const detail = error instanceof Error ? error.message : String(error);
    return {success: false, error: `catalog root unreadable: ${detail}`};
  }
  const ownedElsewhere = (query: Query): boolean => {
    const owner = owners.get(query.query);
    return owner !== undefined && owner !== shortName;
  };
  return {
    success: true,
    kept: queries.filter((query) => !ownedElsewhere(query)),
    deferred: queries.filter(ownedElsewhere),
  };
};

export const resolveTriggerConfig = (
  argv: readonly string[],
  options: TriggerConfigOptions = {},
): TriggerConfigResult => {
  const cli = parseCli(argv);
  if (cli.positionals.length > 3) {
    return failure(
      "usage",
      `unrecognized arguments: ${cli.positionals.slice(3).join(" ")}`,
    );
  }

  const environment = options.environment ?? process.env;
  const workingDirectory = options.workingDirectory ?? resolve(".");
  const guideDirectory = resolveGuideDirectory(workingDirectory);
  const queriesArgument =
    cli.positionals[0] ?? join(guideDirectory, "eval-queries.json");
  const queriesFile = resolve(workingDirectory, queriesArgument);
  const skillMd = join(guideDirectory, "SKILL.md");
  const skillName =
    cli.positionals[1] ??
    (isFile(skillMd) ? parseFrontmatterName(skillMd) : undefined);
  if (skillName === undefined) {
    return failure(
      "usage",
      "the following arguments are required: skill_name " +
        "(no SKILL.md with a name frontmatter found)",
    );
  }

  const positionalSplit = cli.positionals[2];
  if (
    positionalSplit !== undefined &&
    cli.split !== undefined &&
    cli.split !== positionalSplit
  ) {
    return failure(
      "usage",
      "split must not be provided twice with different values",
    );
  }
  const splitInput = cli.split ?? positionalSplit ?? "all";
  const splitResult = parseQuerySplit(splitInput);
  if (!splitResult.success) {
    return failure(
      "usage",
      `argument split: invalid choice: '${splitInput}' ` +
        "(choose from 'train', 'validation', 'all')",
    );
  }

  if (!isFile(queriesFile)) {
    return failure("runtime", `queries file not found: ${queriesFile}`);
  }
  const resolveExecutablePath =
    options.resolveExecutablePath ?? defaultResolveExecutablePath;
  const harnessInput =
    cli.harness ?? environment.SKILL_EVAL_HARNESS ?? "claude";
  if (harnessInput !== "claude" && harnessInput !== "codex") {
    return failure(
      "usage",
      `argument harness: invalid choice: '${harnessInput}' (choose from 'claude', 'codex')`,
    );
  }
  const harness = harnessInput;
  const harnessExecutable = resolveExecutablePath(harness);
  if (harnessExecutable === undefined) {
    return failure("runtime", `'${harness}' CLI not found in PATH`);
  }
  const executableRuns = options.executableRuns ?? defaultExecutableRuns;
  if (!executableRuns(harnessExecutable)) {
    return failure("runtime", `'${harness}' CLI not found in PATH`);
  }

  const optionsResult = parseTriggerOptions({
    runs: cli.runs ?? environment.RUNS ?? "3",
    threshold: cli.threshold ?? environment.THRESHOLD ?? "0.5",
    budget: cli.maxBudget ?? environment.MAX_BUDGET_USD ?? "0.05",
    ...(cli.batchSize === undefined ? {} : {batchSize: cli.batchSize}),
  });
  if (!optionsResult.success) {
    return failure(
      "usage",
      `invalid evaluator options: ${optionsResult.error}`,
    );
  }
  const {runs, threshold, budget, batchSize} = optionsResult.data;

  const pluginDirectoryArgument =
    cli.pluginDir ?? environment.PLUGIN_DIR ?? dirname(guideDirectory);
  const pluginDirectory = resolve(workingDirectory, pluginDirectoryArgument);
  if (
    harness === "claude" &&
    (!isDirectory(pluginDirectory) ||
      !isFile(join(pluginDirectory, ".claude-plugin", "plugin.json")))
  ) {
    return failure(
      "runtime",
      `target plugin directory is invalid: ${pluginDirectory}`,
    );
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(readFileSync(queriesFile, "utf8"));
  } catch (error: unknown) {
    const detail = error instanceof Error ? error.message : String(error);
    return failure("runtime", `invalid JSON in ${queriesFile}: ${detail}`);
  }
  const queryResult = parseQueries(parsed);
  if (!queryResult.success) {
    return failure(
      "runtime",
      `invalid queries in ${queriesFile}: ${queryResult.error}`,
    );
  }
  const selectionResult = selectQueries(queryResult.data, splitResult.data);
  if (!selectionResult.success) {
    return failure("runtime", selectionResult.error);
  }

  const shortNameForCatalog = skillName.split(":").pop() ?? skillName;
  const deferredResult = partitionOwnedQueries(
    selectionResult.data,
    shortNameForCatalog,
    cli.catalogRoot ?? environment.CATALOG_ROOT,
    workingDirectory,
  );
  if (!deferredResult.success) {
    return failure("runtime", deferredResult.error);
  }
  const {kept, deferred} = deferredResult;

  const queryBatches = boundedBatches(
    kept,
    batchSize ?? Math.max(kept.length, 1),
  );
  const maxBatchModelRuns = Math.max(
    0,
    ...queryBatches.map((batch) => triggerModelRunCount(batch.length, runs)),
  );
  if (maxBatchModelRuns > MAX_TRIGGER_MODEL_RUNS) {
    return failure(
      "runtime",
      `trigger eval batch would launch ${String(maxBatchModelRuns)} model runs; ` +
        `maximum is ${String(MAX_TRIGGER_MODEL_RUNS)}`,
    );
  }

  const defaultModel = skillEvalModelDefaults(harness).generation;
  const environmentModel =
    harness === "claude" ? environment.CLAUDE_MODEL : environment.CODEX_MODEL;
  const modelInput = cli.model ?? environmentModel ?? defaultModel;
  const model = modelInput.trim().length > 0 ? modelInput : defaultModel;
  const shortName = skillName.split(":").pop() ?? skillName;

  return {
    success: true,
    data: {
      budget,
      guideDirectory,
      harness,
      harnessArgs:
        harness === "claude"
          ? buildTriggerClaudeArgs({
              model,
              budget,
              pluginDirectories:
                resolveClaudePluginDirectories(pluginDirectory),
            })
          : buildTriggerCodexArgs({model}),
      harnessExecutable,
      maxBatchModelRuns,
      model,
      queryBatches,
      deferredToRouting: deferred.length,
      queryCount: kept.length,
      runs,
      shortName,
      skillName,
      split: splitResult.data,
      threshold,
      workspace:
        cli.workspace === undefined
          ? undefined
          : resolve(workingDirectory, cli.workspace),
    },
  };
};
