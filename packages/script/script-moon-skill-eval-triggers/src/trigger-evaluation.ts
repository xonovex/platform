import {writeFileSync} from "node:fs";
import {join} from "node:path";
import {type TriggerOutcome} from "./trigger-process.js";
import {type Query} from "./validation.js";

interface ResultRecord {
  readonly query: string;
  readonly should_trigger: boolean;
  readonly triggers: number;
  readonly runs: number;
  readonly trigger_rate: number;
  readonly pass: boolean;
  readonly rationale: string;
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
}

// A harness error is transient (a dropped stream, a timeout, a killed process), so
// retry the run before discarding the whole batch's evidence. A run that fails every
// attempt still invalidates: partial evidence must never be reported as a pass.
export const TRIGGER_RUN_ATTEMPTS = 3;

const runOnce = async (
  query: string,
  options: TriggerEvaluationOptions,
): Promise<{readonly triggered: boolean} | {readonly error: string}> => {
  let lastError = "unknown failure";
  for (let attempt = 1; attempt <= TRIGGER_RUN_ATTEMPTS; attempt += 1) {
    const outcome = await options.check(query);
    if (outcome.error === null) return {triggered: outcome.triggered};
    lastError = outcome.error;
    if (attempt < TRIGGER_RUN_ATTEMPTS) {
      process.stderr.write(
        `retrying after transient failure (${String(attempt)}/${String(TRIGGER_RUN_ATTEMPTS)}): ${outcome.error}\n`,
      );
    }
  }
  return {error: lastError};
};

const evaluateQuery = async (
  entry: Query,
  options: TriggerEvaluationOptions,
): Promise<ResultRecord | undefined> => {
  let triggers = 0;
  for (let index = 0; index < options.runs; index += 1) {
    const outcome = await runOnce(entry.query, options);
    if ("error" in outcome) {
      if (options.workspace !== undefined) {
        writeFileSync(
          join(options.workspace, "invalid-run.json"),
          `${JSON.stringify({query: entry.query, error: outcome.error}, null, 2)}\n`,
          "utf8",
        );
      }
      process.stderr.write(
        `Error: trigger infrastructure failure for query ${JSON.stringify(entry.query)}: ${outcome.error}\n`,
      );
      return undefined;
    }
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
  };
};

export const runTriggerEvaluation = async (
  options: TriggerEvaluationOptions,
): Promise<TriggerEvaluationResult> => {
  const results: ResultRecord[] = [];
  for (const [batchIndex, batch] of options.queryBatches.entries()) {
    process.stderr.write(
      `batch ${String(batchIndex + 1)}/${String(options.queryBatches.length)}: ` +
        `${String(batch.length)} queries\n`,
    );
    for (const entry of batch) {
      const result = await evaluateQuery(entry, options);
      if (result === undefined) {
        return {success: false, results, passed: 0, failed: 0, total: 0};
      }
      results.push(result);
      process.stdout.write(`${JSON.stringify(result)}\n`);
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
