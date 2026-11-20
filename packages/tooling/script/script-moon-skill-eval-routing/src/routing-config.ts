import {join, resolve} from "node:path";
import type {ParsedArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {SkillEvalHarness} from "@xonovex/script-moon-common/skill-eval-models";
import {planBatches} from "@xonovex/script-moon-skill-eval-common/batch-plan";
import type {Environment} from "@xonovex/script-moon-skill-eval-common/environment";
import {
  harnessInput,
  locateHarness,
  parseHarness,
  resolveHarnessModel,
} from "@xonovex/script-moon-skill-eval-common/harness-config";
import type {ValidationResult} from "@xonovex/script-moon-skill-eval-common/option-schemas";
import {
  buildRoutingScenarios,
  type RoutingScenario,
} from "@xonovex/script-moon-skill-eval-common/routing-catalog";
import {
  MAX_TRIGGER_MODEL_RUNS,
  parseQuerySplit,
  parseTriggerOptions,
  triggerModelRunCount,
} from "@xonovex/script-moon-skill-eval-common/validation";

export const cliDefinition = {
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

/** Where the configuration reads its defaults, the catalog and the harness from. */
export interface RoutingConfigContext {
  readonly environment: Environment;
  readonly workingDirectory: string;
  readonly platform: NodeJS.Platform;
  readonly fs: FileSystem;
}

export interface RoutingConfig {
  readonly budget: number;
  readonly batches: readonly (readonly RoutingScenario[])[];
  readonly catalogScenarioCount: number;
  readonly executable: string;
  readonly harness: SkillEvalHarness;
  readonly invalidRunPath: string;
  readonly model: string;
  readonly resultsPath: string;
  readonly runs: number;
  readonly split: "train" | "validation" | "all";
  readonly summaryPath: string;
  readonly threshold: number;
  readonly workspace: string;
}

const failure = (error: string): ValidationResult<RoutingConfig> => ({
  success: false,
  error,
});

const parseNonNegativeInteger = (
  name: string,
  raw: string,
): ValidationResult<number> =>
  /^\d+$/.test(raw)
    ? {success: true, data: Number(raw)}
    : {success: false, error: `${name} must be a non-negative integer`};

const textValue = (
  values: ParsedArgs["values"],
  name: string,
): string | undefined => {
  const value = values[name];
  return typeof value === "string" ? value : undefined;
};

const selectScenarios = (
  scenarios: readonly RoutingScenario[],
  split: "train" | "validation" | "all",
  owners: ReadonlySet<string>,
  offset: number,
  limit: number | undefined,
): readonly RoutingScenario[] => {
  const selected = scenarios.filter(
    (scenario) =>
      (split === "all" || scenario.split === split) &&
      (owners.size === 0 || owners.has(scenario.expectedSkill)),
  );
  return limit === undefined
    ? selected.slice(offset)
    : selected.slice(offset, offset + limit);
};

export const resolveRoutingConfig = (
  args: ParsedArgs,
  context: RoutingConfigContext,
): ValidationResult<RoutingConfig> => {
  const {values, positionals} = args;
  const {environment, workingDirectory, fs} = context;
  if (positionals.length > 1) {
    return failure(`unrecognized arguments: ${positionals.slice(1).join(" ")}`);
  }
  const catalogRoot = resolve(
    workingDirectory,
    positionals[0] ?? "packages/runtime/plugin",
  );
  const requestedHarness = harnessInput(
    textValue(values, "harness"),
    environment,
  );
  const harness = parseHarness(requestedHarness);
  if (harness === undefined) {
    return failure(
      `invalid harness '${requestedHarness}'; expected claude or codex`,
    );
  }
  const splitResult = parseQuerySplit(textValue(values, "split") ?? "all");
  if (!splitResult.success) {
    return failure(`invalid split: ${splitResult.error}`);
  }
  const optionsResult = parseTriggerOptions({
    runs: textValue(values, "runs") ?? "1",
    threshold: textValue(values, "threshold") ?? "0.5",
    budget: textValue(values, "max-budget-usd") ?? "0.05",
    batchSize: textValue(values, "batch-size") ?? "8",
  });
  if (!optionsResult.success) {
    return failure(`invalid evaluator options: ${optionsResult.error}`);
  }
  const offset = parseNonNegativeInteger(
    "offset",
    textValue(values, "offset") ?? "0",
  );
  if (!offset.success) return offset;
  const limitText = textValue(values, "limit");
  const limit =
    limitText === undefined
      ? undefined
      : parseNonNegativeInteger("limit", limitText);
  if (limit !== undefined && !limit.success) return limit;

  let catalogScenarios: readonly RoutingScenario[];
  try {
    catalogScenarios = buildRoutingScenarios(catalogRoot, fs);
  } catch (error: unknown) {
    return failure(error instanceof Error ? error.message : String(error));
  }
  const owners = new Set(
    (textValue(values, "owners") ?? "")
      .split(",")
      .map((owner) => owner.trim())
      .filter(Boolean),
  );
  const scenarios = selectScenarios(
    catalogScenarios,
    splitResult.data,
    owners,
    offset.data,
    limit?.data,
  );
  if (scenarios.length === 0) return failure("no routing scenarios selected");

  const {batchSize, budget, runs, threshold} = optionsResult.data;
  const {batches, maxBatchRuns} = planBatches(scenarios, batchSize, (count) =>
    triggerModelRunCount(count, runs),
  );
  if (maxBatchRuns > MAX_TRIGGER_MODEL_RUNS) {
    return failure(
      `routing eval batch would launch ${String(maxBatchRuns)} model runs; maximum is ${String(MAX_TRIGGER_MODEL_RUNS)}`,
    );
  }

  const executable = locateHarness(harness, environment, context.platform, fs);
  if (executable === undefined) {
    return failure(`'${harness}' CLI not found in PATH`);
  }
  const workspace = resolve(
    workingDirectory,
    textValue(values, "workspace") ?? `.skill-eval-results/routing/${harness}`,
  );
  return {
    success: true,
    data: {
      budget,
      batches,
      catalogScenarioCount: catalogScenarios.length,
      executable,
      harness,
      invalidRunPath: join(workspace, "invalid-run.json"),
      model: resolveHarnessModel(
        harness,
        "generation",
        textValue(values, "model"),
        environment,
      ),
      resultsPath: join(workspace, "results.jsonl"),
      runs,
      split: splitResult.data,
      summaryPath: join(workspace, "summary.json"),
      threshold,
      workspace,
    },
  };
};
