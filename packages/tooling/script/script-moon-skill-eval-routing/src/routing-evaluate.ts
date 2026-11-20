import {resolveClaudePluginDirectories} from "@xonovex/script-moon-common/claude-plugin";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {buildTriggerClaudeArgs} from "@xonovex/script-moon-skill-eval-common/claude/args";
import {buildIsolatedCodexArgs} from "@xonovex/script-moon-skill-eval-common/codex/args";
import {
  clearEvidence,
  writeJsonEvidence,
  writeJsonLinesEvidence,
} from "@xonovex/script-moon-skill-eval-common/evidence";
import type {
  EvaluatorPorts,
  HarnessPorts,
} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import {
  EVAL_EXIT,
  reporter,
  verdictExit,
} from "@xonovex/script-moon-skill-eval-common/report";
import type {RoutingScenario} from "@xonovex/script-moon-skill-eval-common/routing-catalog";
import {checkTrigger} from "@xonovex/script-moon-skill-eval-common/trigger-check";
import {
  cliDefinition,
  resolveRoutingConfig,
  type RoutingConfig,
} from "./routing-config.js";

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

interface RoutingFailure {
  readonly error: string;
}

const scenarioArgs = (
  scenario: RoutingScenario,
  config: RoutingConfig,
  ports: HarnessPorts,
): readonly string[] =>
  config.harness === "claude"
    ? buildTriggerClaudeArgs({
        model: config.model,
        budget: config.budget,
        pluginDirectories: unique(
          scenario.candidates.flatMap(({pluginDirectory}) =>
            resolveClaudePluginDirectories(pluginDirectory, ports.fs),
          ),
        ),
      })
    : buildIsolatedCodexArgs({model: config.model});

const evaluateScenario = async (
  scenario: RoutingScenario,
  config: RoutingConfig,
  ports: HarnessPorts,
): Promise<RoutingRecord | RoutingFailure> => {
  const target = scenario.expectedSkill;
  const args = scenarioArgs(scenario, config, ports);
  let ownerSelections = 0;
  for (let runIndex = 0; runIndex < config.runs; runIndex += 1) {
    const outcome = await checkTrigger(
      config.harness,
      {
        executable: config.executable,
        args,
        query: scenario.query,
        target,
        shortName: target,
        candidates: scenario.candidates,
      },
      ports,
    );
    if (outcome.error !== null) return {error: outcome.error};
    if (outcome.triggered) ownerSelections += 1;
  }

  const selectionRate = ownerSelections / config.runs;
  return {
    query: scenario.query,
    split: scenario.split,
    expected_skill: target,
    candidate_skills: scenario.candidates.map(({shortName}) => shortName),
    owner_selections: ownerSelections,
    runs: config.runs,
    selection_rate: Math.round(selectionRate * 1000) / 1000,
    pass: selectionRate >= config.threshold,
  };
};

export const main = async (
  argv: readonly string[],
  workingDirectory: string,
  ports: EvaluatorPorts,
): Promise<number> => {
  const report = reporter(ports.output);
  const args = parseCliArgs(cliDefinition, argv);
  if (args.kind === "help") {
    ports.output.stdout(args.text);
    return EVAL_EXIT.passed;
  }
  const configResult = resolveRoutingConfig(args, {
    environment: ports.environment,
    workingDirectory,
    platform: ports.platform,
    fs: ports.fs,
  });
  if (!configResult.success) {
    report.error(configResult.error);
    return EVAL_EXIT.invalid;
  }
  const config = configResult.data;
  ports.fs.makeDirectory(config.workspace);
  clearEvidence(ports.scratch, [
    config.resultsPath,
    config.summaryPath,
    config.invalidRunPath,
  ]);

  const records: RoutingRecord[] = [];
  for (const [batchIndex, batch] of config.batches.entries()) {
    report.info(
      `batch ${String(batchIndex + 1)}/${String(config.batches.length)}: ${String(batch.length)} routing scenarios`,
    );
    for (const scenario of batch) {
      const record = await evaluateScenario(scenario, config, ports);
      if ("error" in record) {
        writeJsonEvidence(ports.fs, config.invalidRunPath, {
          query: scenario.query,
          error: record.error,
        });
        report.error(`routing infrastructure failure: ${record.error}`);
        return EVAL_EXIT.invalid;
      }
      records.push(record);
      report.result(record);
    }
  }

  writeJsonLinesEvidence(ports.fs, config.resultsPath, records);
  const passed = records.filter(({pass}) => pass).length;
  writeJsonEvidence(ports.fs, config.summaryPath, {
    harness: config.harness,
    model: config.model,
    split: config.split,
    selected_owners: [
      ...new Set(records.map(({expected_skill}) => expected_skill)),
    ].toSorted(),
    catalog_scenarios: config.catalogScenarioCount,
    selected_scenarios: records.length,
    runs_per_scenario: config.runs,
    passed,
    failed: records.length - passed,
  });
  return verdictExit(passed === records.length);
};
