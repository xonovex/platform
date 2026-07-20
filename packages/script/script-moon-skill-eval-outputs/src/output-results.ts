import {type EvaluationArm} from "./validation.js";

const TOKEN_KEYS = [
  "input_tokens",
  "output_tokens",
  "cache_creation_input_tokens",
  "cache_read_input_tokens",
] as const;

export interface AssertionResult {
  readonly text: string;
  readonly passed: boolean;
  readonly evidence: string;
}

export interface GradeSummary {
  readonly passed: number;
  readonly failed: number;
  readonly total: number;
  readonly pass_rate: number;
}

export interface Graded {
  readonly assertion_results: readonly AssertionResult[];
  readonly summary: GradeSummary;
  readonly error: string | null;
}

export interface JobRecord {
  readonly id: string | number;
  readonly arm: EvaluationArm;
  readonly pass_rate: number;
  readonly tokens: number;
  readonly duration_ms: number;
  readonly skill_triggered: boolean;
  readonly error: string | null;
}

export interface MeanBlock {
  mean: number;
  stddev?: number;
}

export interface ArmBlock {
  readonly pass_rate: MeanBlock;
  readonly tokens: MeanBlock;
  readonly duration_ms: MeanBlock;
  skill_trigger_rate?: {readonly mean: number};
}

export const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

export const round = (value: number, digits = 0): number => {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
};

export const fmean = (values: readonly number[]): number =>
  values.reduce((total, value) => total + value, 0) / values.length;

const populationDeviation = (values: readonly number[]): number => {
  const mean = fmean(values);
  const variance =
    values.reduce((total, value) => total + (value - mean) ** 2, 0) /
    values.length;
  return Math.sqrt(variance);
};

export const sumTokens = (usage: unknown): number => {
  if (!isRecord(usage)) return 0;
  let total = 0;
  for (const key of TOKEN_KEYS) {
    const raw = usage[key];
    const number = typeof raw === "number" ? raw : Number(raw);
    total += Number.isFinite(number) ? Math.trunc(number) : 0;
  }
  return total;
};

export const extractJson = (text: string): Record<string, unknown> | null => {
  if (!text) return null;
  const fenced = /```(?:json)?\s*(\{[\s\S]*?\})\s*```/.exec(text);
  let candidate: string | null = fenced?.[1] ?? null;
  if (candidate === null) {
    const start = text.indexOf("{");
    const end = text.lastIndexOf("}");
    candidate = start !== -1 && end > start ? text.slice(start, end + 1) : null;
  }
  if (candidate === null) return null;
  try {
    const value: unknown = JSON.parse(candidate);
    return isRecord(value) ? value : null;
  } catch {
    return null;
  }
};

export const claudeFailureDetail = (stdout: string): string => {
  const lines = stdout.trim().split(/\r?\n/).toReversed();
  for (const line of lines) {
    let value: unknown;
    try {
      value = JSON.parse(line);
    } catch {
      continue;
    }
    if (!isRecord(value)) continue;
    const result = typeof value.result === "string" ? value.result.trim() : "";
    const subtype =
      typeof value.subtype === "string" ? value.subtype.trim() : "";
    if (result && (value.is_error === true || subtype.startsWith("error"))) {
      return result.slice(0, 500);
    }
    if (subtype.startsWith("error")) return subtype.slice(0, 500);
  }
  return "";
};

export const summarize = (
  results: readonly AssertionResult[],
  error: string | null = null,
): Graded => {
  const passed = results.filter((result) => result.passed).length;
  const total = results.length;
  return {
    assertion_results: results,
    summary: {
      passed,
      failed: total - passed,
      total,
      pass_rate: total > 0 ? round(passed / total, 3) : 0,
    },
    error,
  };
};

export const meanBlock = (
  values: readonly number[],
  runs: number,
): MeanBlock => {
  const block: MeanBlock = {
    mean: values.length > 0 ? round(fmean(values), 3) : 0,
  };
  if (runs > 1 && values.length > 1) {
    block.stddev = round(populationDeviation(values), 3);
  }
  return block;
};

export const aggregateArm = (
  records: readonly JobRecord[],
  arm: EvaluationArm,
  runs: number,
): ArmBlock => {
  const armRecords = records.filter((record) => record.arm === arm);
  const byEvaluation = new Map<string | number, JobRecord[]>();
  for (const record of armRecords) {
    const group = byEvaluation.get(record.id) ?? [];
    group.push(record);
    byEvaluation.set(record.id, group);
  }
  const groups = [...byEvaluation.values()];
  const block: ArmBlock = {
    pass_rate: meanBlock(
      groups.map((group) => fmean(group.map((record) => record.pass_rate))),
      runs,
    ),
    tokens: meanBlock(
      groups.map((group) => fmean(group.map((record) => record.tokens))),
      runs,
    ),
    duration_ms: meanBlock(
      groups.map((group) => fmean(group.map((record) => record.duration_ms))),
      runs,
    ),
  };
  if (arm === "with_skill") {
    const fired = armRecords.map((record) => (record.skill_triggered ? 1 : 0));
    block.skill_trigger_rate = {
      mean: fired.length > 0 ? round(fmean(fired), 3) : 0,
    };
  }
  return block;
};
