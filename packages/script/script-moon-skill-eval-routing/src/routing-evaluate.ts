import {rmSync} from "node:fs";
import {join, resolve} from "node:path";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {resolveClaudePluginDirectories} from "@xonovex/script-moon-common/fs";
import {skillEvalModelDefaults} from "@xonovex/script-moon-common/skill-eval-models";
import {
  buildRoutingScenarios,
  type RoutingScenario,
} from "@xonovex/script-moon-skill-eval-common/routing-catalog";
import {
  checkCodexTriggered,
  checkTriggered,
} from "@xonovex/script-moon-skill-eval-common/trigger-process";
import {type TriggerOutcome} from "@xonovex/script-moon-skill-eval-common/trigger-scan";
import {
  buildIsolatedCodexArgs,
  buildTriggerClaudeArgs,
  MAX_TRIGGER_MODEL_RUNS,
  parseQuerySplit,
  parseTriggerOptions,
  triggerModelRunCount,
} from "@xonovex/script-moon-skill-eval-common/validation";

const cliDefinition = {
  name: "moon-skill-eval-routing",
  description:
    "Evaluate shared catalog queries with the expected owner and its competing skills loaded together.",
  options: {
    runs: {type: "string", description: "runs per scenario (default 1)"},
    threshold: {
      type: "string",
      description: "owner-selection cutoff (default 0.5)",
    },
    harness: {type: "string", description: "claude or codex"},
    model: {type: "string", description: "harness model"},
    split: {type: "string", description: "train, validation, or all"},
    "batch-size": {type: "string", description: "scenarios per bounded batch"},
    workspace: {type: "string", description: "evidence output directory"},
    owners: {
      type: "string",
      description: "comma-separated expected skill owners; empty selects all",
    },
    limit: {type: "string", description: "maximum scenarios after offset"},
    offset: {type: "string", description: "first scenario index (default 0)"},
    "max-budget-usd": {type: "string", description: "Claude spend cap per run"},
  },
} as const;

const parseNonNegativeInteger = (name: string, raw: string): number => {
  if (!/^\d+$/.test(raw))
    throw new Error(`${name} must be a non-negative integer`);
  return Number(raw);
};

const unique = (values: readonly string[]): readonly string[] => [
  ...new Set(values),
];

interface RoutingRecord {
  readonly query: string;
  readonly split: "train" | "validation";
  readonly expected_skill: string;
  readonly candidate_skills: readonly string[];
  readonly owner_selections: number;
  readonly runs: number;
  readonly selection_rate: number;
  readonly pass: boolean;
}

/**
 * The effects the routing evaluator reaches a harness through. A run supplies the
 * spawning implementations; a test supplies decisions directly, so a whole catalog
 * sweep is scored without a process.
 */
export interface RoutingDependencies {
  readonly checkTriggered: typeof checkTriggered;
  readonly checkCodexTriggered: typeof checkCodexTriggered;
  readonly resolveExecutable: (command: string) => string;
  readonly fs: FileSystem;
  // Removes a previous run's evidence so a rerun cannot be read as this run's.
  readonly discard: (path: string) => void;
}

export const defaultDependencies: RoutingDependencies = {
  checkTriggered,
  checkCodexTriggered,
  resolveExecutable,
  fs: nodeFileSystem,
  discard: (path) => {
    rmSync(path, {force: true});
  },
};

interface RoutingContext {
  readonly budget: number;
  readonly executable: string;
  readonly harness: "claude" | "codex";
  readonly model: string;
  readonly runs: number;
  readonly threshold: number;
  readonly dependencies: RoutingDependencies;
}

interface RoutingFailure {
  readonly error: string;
}

const evaluateScenario = async (
  scenario: RoutingScenario,
  context: RoutingContext,
): Promise<RoutingRecord | RoutingFailure> => {
  const candidateNames = scenario.candidates.map(({shortName}) => shortName);
  const target = scenario.expectedSkill;
  const targetGuide = scenario.candidates.find(
    ({shortName}) => shortName === target,
  );
  if (targetGuide === undefined) throw new Error(`owner missing: ${target}`);

  const args =
    context.harness === "claude"
      ? buildTriggerClaudeArgs({
          model: context.model,
          budget: context.budget,
          pluginDirectories: unique(
            scenario.candidates.flatMap(({pluginDirectory}) =>
              resolveClaudePluginDirectories(
                pluginDirectory,
                context.dependencies.fs,
              ),
            ),
          ),
        })
      : buildIsolatedCodexArgs({model: context.model});
  let ownerSelections = 0;
  for (let runIndex = 0; runIndex < context.runs; runIndex += 1) {
    const outcome: TriggerOutcome =
      context.harness === "claude"
        ? await context.dependencies.checkTriggered(
            scenario.query,
            args,
            target,
            target,
            context.executable,
          )
        : await context.dependencies.checkCodexTriggered({
            args,
            executable: context.executable,
            guideDirectory: targetGuide.guideDirectory,
            query: scenario.query,
            shortName: target,
            candidateGuides: scenario.candidates,
          });
    if (outcome.error !== null) return {error: outcome.error};
    if (outcome.triggered) ownerSelections += 1;
  }

  const selectionRate = ownerSelections / context.runs;
  return {
    query: scenario.query,
    split: scenario.split,
    expected_skill: target,
    candidate_skills: candidateNames,
    owner_selections: ownerSelections,
    runs: context.runs,
    selection_rate: Math.round(selectionRate * 1000) / 1000,
    pass: selectionRate >= context.threshold,
  };
};

export const main = async (
  argv: readonly string[],
  dependencies: RoutingDependencies = defaultDependencies,
): Promise<number> => {
  const {values, positionals} = parseCliArgs(cliDefinition, argv);
  if (positionals.length > 1) {
    throw new Error(
      `unrecognized arguments: ${positionals.slice(1).join(" ")}`,
    );
  }
  const catalogRoot = resolve(positionals[0] ?? "packages/skill");
  const harnessInput =
    (values.harness as string | undefined) ??
    process.env.SKILL_EVAL_HARNESS ??
    "claude";
  if (harnessInput !== "claude" && harnessInput !== "codex") {
    throw new Error(
      `invalid harness '${harnessInput}'; expected claude or codex`,
    );
  }
  const splitResult = parseQuerySplit(values.split ?? "all");
  if (!splitResult.success)
    throw new Error(`invalid split: ${splitResult.error}`);
  const optionsResult = parseTriggerOptions({
    runs: (values.runs as string | undefined) ?? "1",
    threshold: (values.threshold as string | undefined) ?? "0.5",
    budget: (values["max-budget-usd"] as string | undefined) ?? "0.05",
    batchSize: (values["batch-size"] as string | undefined) ?? "8",
  });
  if (!optionsResult.success) {
    throw new Error(`invalid evaluator options: ${optionsResult.error}`);
  }
  const offset = parseNonNegativeInteger(
    "offset",
    (values.offset as string | undefined) ?? "0",
  );
  const limitRaw = values.limit as string | undefined;
  const limit =
    limitRaw === undefined
      ? undefined
      : parseNonNegativeInteger("limit", limitRaw);
  const catalogScenarios = buildRoutingScenarios(catalogRoot, dependencies.fs);
  const owners = new Set(
    ((values.owners as string | undefined) ?? "")
      .split(",")
      .map((owner) => owner.trim())
      .filter(Boolean),
  );
  const splitScenarios = catalogScenarios.filter(
    (scenario) =>
      (splitResult.data === "all" || scenario.split === splitResult.data) &&
      (owners.size === 0 || owners.has(scenario.expectedSkill)),
  );
  const scenarios =
    limit === undefined
      ? splitScenarios.slice(offset)
      : splitScenarios.slice(offset, offset + limit);
  if (scenarios.length === 0) throw new Error("no routing scenarios selected");

  const {batchSize, budget, runs, threshold} = optionsResult.data;
  const batches = boundedBatches(scenarios, batchSize ?? 8);
  const maxBatchRuns = Math.max(
    ...batches.map((batch) => triggerModelRunCount(batch.length, runs)),
  );
  if (maxBatchRuns > MAX_TRIGGER_MODEL_RUNS) {
    throw new Error(
      `routing eval batch would launch ${String(maxBatchRuns)} model runs; maximum is ${String(MAX_TRIGGER_MODEL_RUNS)}`,
    );
  }

  const harness = harnessInput;
  const executable = dependencies.resolveExecutable(harness);
  const defaultModel = skillEvalModelDefaults(harness).generation;
  const modelInput =
    (values.model as string | undefined) ??
    (harness === "claude"
      ? process.env.CLAUDE_MODEL
      : process.env.CODEX_MODEL) ??
    defaultModel;
  const model = modelInput.trim().length > 0 ? modelInput : defaultModel;
  const workspace = resolve(
    (values.workspace as string | undefined) ??
      `.skill-eval-results/routing/${harness}`,
  );
  dependencies.fs.makeDirectory(workspace);
  const resultsPath = join(workspace, "results.jsonl");
  const summaryPath = join(workspace, "summary.json");
  const invalidRunPath = join(workspace, "invalid-run.json");
  dependencies.discard(resultsPath);
  dependencies.discard(summaryPath);
  dependencies.discard(invalidRunPath);

  const context: RoutingContext = {
    budget,
    executable,
    harness,
    model,
    runs,
    threshold,
    dependencies,
  };
  const records: RoutingRecord[] = [];
  for (const [batchIndex, batch] of batches.entries()) {
    process.stderr.write(
      `batch ${String(batchIndex + 1)}/${String(batches.length)}: ${String(batch.length)} routing scenarios\n`,
    );
    for (const scenario of batch) {
      const record = await evaluateScenario(scenario, context);
      if ("error" in record) {
        dependencies.fs.writeFile(
          invalidRunPath,
          `${JSON.stringify({query: scenario.query, error: record.error}, null, 2)}\n`,
        );
        process.stderr.write(
          `Error: routing infrastructure failure: ${record.error}\n`,
        );
        return 2;
      }
      records.push(record);
      process.stdout.write(`${JSON.stringify(record)}\n`);
    }
  }

  dependencies.fs.writeFile(
    resultsPath,
    `${records.map((record) => JSON.stringify(record)).join("\n")}\n`,
  );
  const passed = records.filter(({pass}) => pass).length;
  dependencies.fs.writeFile(
    summaryPath,
    `${JSON.stringify(
      {
        harness,
        model,
        split: splitResult.data,
        selected_owners: [
          ...new Set(records.map(({expected_skill}) => expected_skill)),
        ].toSorted(),
        catalog_scenarios: catalogScenarios.length,
        selected_scenarios: records.length,
        runs_per_scenario: runs,
        passed,
        failed: records.length - passed,
      },
      null,
      2,
    )}\n`,
  );
  return passed === records.length ? 0 : 1;
};
