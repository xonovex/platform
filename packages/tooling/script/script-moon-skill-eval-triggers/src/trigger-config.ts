import {join, resolve} from "node:path";
import {resolveClaudePluginDirectories} from "@xonovex/script-moon-common/claude-plugin";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  findPluginDirectory,
  resolveSkillDirectory,
} from "@xonovex/script-moon-common/skill-directory";
import {planBatches} from "@xonovex/script-moon-skill-eval-common/batch-plan";
import {buildTriggerClaudeArgs} from "@xonovex/script-moon-skill-eval-common/claude/args";
import {buildIsolatedCodexArgs} from "@xonovex/script-moon-skill-eval-common/codex/args";
import type {Environment} from "@xonovex/script-moon-skill-eval-common/environment";
import {
  declaredSkillName,
  harnessInput,
  locateHarness,
  parseHarness,
  resolveHarnessModel,
  shortSkillName,
} from "@xonovex/script-moon-skill-eval-common/harness-config";
import {catalogQueryOwners} from "@xonovex/script-moon-skill-eval-common/routing-catalog";
import {
  MAX_TRIGGER_MODEL_RUNS,
  parseQueries,
  parseQuerySplit,
  parseTriggerOptions,
  selectQueries,
  triggerModelRunCount,
  type Query,
} from "@xonovex/script-moon-skill-eval-common/validation";
import type {ParsedCli} from "./cli.js";

type TriggerConfigResult =
  | {readonly success: true; readonly data: TriggerConfig}
  | {
      readonly success: false;
      readonly error: string;
      readonly kind: "runtime" | "usage";
    };

/** Where the configuration reads its defaults, files and harness from. */
export interface TriggerConfigContext {
  readonly environment: Environment;
  readonly workingDirectory: string;
  readonly platform: NodeJS.Platform;
  readonly fs: FileSystem;
}

export interface TriggerConfig {
  readonly budget: number;
  readonly skillDirectory: string;
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

// This evaluator loads the target plugin and whatever plugins it declares as
// dependencies, so an overlay is measured alongside the base skill it extends but
// never alongside the rest of the catalog. It can therefore answer whether the target
// claims a query against that narrow field, and a dependency can legitimately win one
// of the queries kept here. A near miss the catalog assigns to a skill outside that
// field asks a different question — which of the two wins with both loaded — and the
// routing evaluator answers it. Scoring such a query here reports a failure for
// correct behaviour, because the rightful owner was never on offer. Without a catalog
// root every query is kept, so a standalone run still evaluates the whole file.
const partitionOwnedQueries = (
  queries: readonly Query[],
  shortName: string,
  catalogRootArgument: string | undefined,
  workingDirectory: string,
  fs: FileSystem,
): PartitionResult => {
  if (catalogRootArgument === undefined) {
    return {success: true, kept: queries, deferred: []};
  }
  let owners: ReadonlyMap<string, string>;
  try {
    owners = catalogQueryOwners(
      resolve(workingDirectory, catalogRootArgument),
      fs,
    );
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
  cli: ParsedCli,
  context: TriggerConfigContext,
): TriggerConfigResult => {
  if (cli.positionals.length > 3) {
    return failure(
      "usage",
      `unrecognized arguments: ${cli.positionals.slice(3).join(" ")}`,
    );
  }

  const {environment, workingDirectory, fs} = context;
  const skillDirectory = resolveSkillDirectory(workingDirectory, fs);
  const queriesArgument =
    cli.positionals[0] ?? join(skillDirectory, "eval-queries.json");
  const queriesFile = resolve(workingDirectory, queriesArgument);
  const skillName = cli.positionals[1] ?? declaredSkillName(skillDirectory, fs);
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

  if (!fs.isFile(queriesFile)) {
    return failure("runtime", `queries file not found: ${queriesFile}`);
  }
  const requestedHarness = harnessInput(cli.harness, environment);
  const harness = parseHarness(requestedHarness);
  if (harness === undefined) {
    return failure(
      "usage",
      `argument harness: invalid choice: '${requestedHarness}' (choose from 'claude', 'codex')`,
    );
  }
  const harnessExecutable = locateHarness(
    harness,
    environment,
    context.platform,
    fs,
  );
  if (harnessExecutable === undefined) {
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
    cli.pluginDir ??
    environment.PLUGIN_DIR ??
    findPluginDirectory(skillDirectory, fs);
  const pluginDirectory = resolve(workingDirectory, pluginDirectoryArgument);
  if (
    harness === "claude" &&
    (!fs.isDirectory(pluginDirectory) ||
      !fs.isFile(join(pluginDirectory, ".claude-plugin", "plugin.json")))
  ) {
    return failure(
      "runtime",
      `target plugin directory is invalid: ${pluginDirectory}`,
    );
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(fs.readText(queriesFile));
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

  const shortName = shortSkillName(skillName);
  const deferredResult = partitionOwnedQueries(
    selectionResult.data,
    shortName,
    cli.catalogRoot ?? environment.CATALOG_ROOT,
    workingDirectory,
    fs,
  );
  if (!deferredResult.success) {
    return failure("runtime", deferredResult.error);
  }
  const {kept, deferred} = deferredResult;

  const {batches: queryBatches, maxBatchRuns: maxBatchModelRuns} = planBatches(
    kept,
    batchSize,
    (count) => triggerModelRunCount(count, runs),
  );
  if (maxBatchModelRuns > MAX_TRIGGER_MODEL_RUNS) {
    return failure(
      "runtime",
      `trigger eval batch would launch ${String(maxBatchModelRuns)} model runs; ` +
        `maximum is ${String(MAX_TRIGGER_MODEL_RUNS)}`,
    );
  }

  const model = resolveHarnessModel(
    harness,
    "generation",
    cli.model,
    environment,
  );

  return {
    success: true,
    data: {
      budget,
      skillDirectory,
      harness,
      harnessArgs:
        harness === "claude"
          ? buildTriggerClaudeArgs({
              model,
              budget,
              pluginDirectories: resolveClaudePluginDirectories(
                pluginDirectory,
                fs,
              ),
            })
          : buildIsolatedCodexArgs({model}),
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
