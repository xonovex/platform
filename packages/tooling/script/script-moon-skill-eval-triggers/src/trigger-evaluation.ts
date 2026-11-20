import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {writeJsonEvidence} from "@xonovex/script-moon-skill-eval-common/evidence";
import type {Reporter} from "@xonovex/script-moon-skill-eval-common/report";
import {
  runWithTransientRetry,
  transientRetryNotice,
} from "@xonovex/script-moon-skill-eval-common/retry";
import {type TriggerOutcome} from "@xonovex/script-moon-skill-eval-common/trigger";
import {type Query} from "@xonovex/script-moon-skill-eval-common/validation";

interface ResultRecord {
  readonly query: string;
  readonly should_trigger: boolean;
  readonly triggers: number;
  readonly runs: number;
  readonly trigger_rate: number;
  readonly pass: boolean;
  readonly rationale: string;
  // One entry per run naming how it ended — the target, a named competitor, the
  // output limit, or nothing. A rate on its own cannot say which skill answered, and
  // recovering that after the fact costs a re-probe per run.
  readonly selected_skills: readonly string[];
}

export interface TriggerEvaluationResult {
  readonly success: boolean;
  readonly results: readonly ResultRecord[];
  readonly passed: number;
  readonly failed: number;
  readonly total: number;
}

export type TriggerCheck = (query: string) => Promise<TriggerOutcome>;

export interface TriggerEvaluationOptions {
  readonly queryBatches: readonly (readonly Query[])[];
  readonly runs: number;
  readonly threshold: number;
  readonly skillName: string;
  readonly workspace: string | undefined;
  readonly check: TriggerCheck;
  readonly fs: FileSystem;
  readonly report: Reporter;
}

const runOnce = async (
  query: string,
  options: TriggerEvaluationOptions,
): Promise<
  | {readonly triggered: boolean; readonly selected: string}
  | {readonly error: string}
> => {
  const outcome = await runWithTransientRetry(
    () => options.check(query),
    (attempt, error) => {
      options.report.info(transientRetryNotice(attempt, error));
    },
  );
  return outcome.error === null
    ? {triggered: outcome.triggered, selected: outcome.selected ?? "unrecorded"}
    : {error: outcome.error};
};

const evaluateQuery = async (
  entry: Query,
  options: TriggerEvaluationOptions,
): Promise<ResultRecord | undefined> => {
  let triggers = 0;
  const selected: string[] = [];
  for (let index = 0; index < options.runs; index += 1) {
    const outcome = await runOnce(entry.query, options);
    if ("error" in outcome) {
      if (options.workspace !== undefined) {
        writeJsonEvidence(
          options.fs,
          join(options.workspace, "invalid-run.json"),
          {query: entry.query, error: outcome.error},
        );
      }
      options.report.error(
        `trigger infrastructure failure for query ${JSON.stringify(entry.query)}: ${outcome.error}`,
      );
      return undefined;
    }
    selected.push(outcome.selected);
    if (outcome.triggered) {
      triggers += 1;
    }
  }

  const rate = triggers / options.runs;
  return {
    query: entry.query,
    should_trigger: entry.should_trigger,
    triggers,
    runs: options.runs,
    trigger_rate: Math.round(rate * 1000) / 1000,
    pass: rate >= options.threshold === entry.should_trigger,
    rationale: entry.rationale,
    selected_skills: selected,
  };
};

export const runTriggerEvaluation = async (
  options: TriggerEvaluationOptions,
): Promise<TriggerEvaluationResult> => {
  const results: ResultRecord[] = [];
  for (const [batchIndex, batch] of options.queryBatches.entries()) {
    options.report.info(
      `batch ${String(batchIndex + 1)}/${String(options.queryBatches.length)}: ` +
        `${String(batch.length)} queries`,
    );
    for (const entry of batch) {
      const result = await evaluateQuery(entry, options);
      if (result === undefined) {
        return {success: false, results, passed: 0, failed: 0, total: 0};
      }
      results.push(result);
      options.report.result(result);
    }
  }

  const passed = results.filter((result) => result.pass).length;
  return {
    success: true,
    results,
    passed,
    failed: results.length - passed,
    total: results.length,
  };
};
