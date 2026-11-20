import {join, resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveClaudePluginDirectories} from "@xonovex/script-moon-common/claude-plugin";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {skillEvalModelDefaults} from "@xonovex/script-moon-common/skill-eval-models";
import {nodeProcessRunner} from "@xonovex/script-moon-skill-eval-common/adapters/process-runner-node";
import {nodeScratch} from "@xonovex/script-moon-skill-eval-common/adapters/scratch-node";
import {buildTriggerClaudeArgs} from "@xonovex/script-moon-skill-eval-common/claude/args";
import {buildIsolatedCodexArgs} from "@xonovex/script-moon-skill-eval-common/codex/args";
import type {HarnessPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import {
  buildRoutingScenarios,
  type RoutingScenario,
} from "@xonovex/script-moon-skill-eval-common/routing-catalog";
import {checkTrigger} from "@xonovex/script-moon-skill-eval-common/trigger-check";
import {
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

export interface RoutingDependencies extends HarnessPorts {
  readonly resolveExecutable: (command: string) => string;
}

export const defaultDependencies: RoutingDependencies = {
  fs: nodeFileSystem,
  scratch: nodeScratch,
  runProcess: nodeProcessRunner,
  environment: process.env,
  resolveExecutable,
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
    const outcome = await checkTrigger(
      context.harness,
      {
        executable: context.executable,
        args,
        query: scenario.query,
        target,
        shortName: target,
        candidates: scenario.candidates,
      },
      context.dependencies,
    );
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
  const args = parseCliArgs(cliDefinition, argv);
  if (args.kind === "help") {
    console.log(args.text);
    return 0;
  }
  const {values, positionals} = args;
  if (positionals.length > 1) {
    throw new Error(
      `unrecognized arguments: ${positionals.slice(1).join(" ")}`,
    );
  }
  const catalogRoot = resolve(positionals[0] ?? "packages/runtime/plugin");
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

  const {batchSize = 8, budget, runs, threshold} = optionsResult.data;
  const batches = boundedBatches(scenarios, batchSize);
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
  dependencies.scratch.remove(resultsPath);
  dependencies.scratch.remove(summaryPath);
  dependencies.scratch.remove(invalidRunPath);

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
