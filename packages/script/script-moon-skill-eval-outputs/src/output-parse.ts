import {isRecord} from "@xonovex/script-moon-common/records";
import {streamTextDeltaLength} from "@xonovex/script-moon-skill-eval-common/validation";
import {extractJson} from "./output-results.js";

// Codex bounds a generation by wall clock alone, so a character ceiling is what
// stops a runaway answer there. Claude already carries a per-run spend cap and a
// turn cap, which bound it sooner.
export const GENERATION_OUTPUT_LIMIT = 10_000;

// Codex signals that it used the staged skill by opening its final response with
// this marker, which the parser then strips from the answer it returns.
export const CODEX_SKILL_SIGNAL = "XONOVEX_SKILL_USED";

const matchSkill = (
  skillField: unknown,
  target: string,
  shortName: string,
): boolean =>
  typeof skillField === "string" &&
  (skillField === target ||
    skillField === shortName ||
    skillField.endsWith(`:${shortName}`));

export const skillInvoked = (
  value: Record<string, unknown>,
  target: string,
  shortName: string,
): boolean => {
  const message = value.message;
  if (!isRecord(message)) return false;
  const content = Array.isArray(message.content) ? message.content : [];
  return content.some((item) => {
    if (!isRecord(item) || item.type !== "tool_use" || item.name !== "Skill") {
      return false;
    }
    return (
      isRecord(item.input) && matchSkill(item.input.skill, target, shortName)
    );
  });
};

export const skillAvailable = (
  value: Record<string, unknown>,
  target: string,
  shortName: string,
): boolean =>
  value.type === "system" &&
  value.subtype === "init" &&
  Array.isArray(value.skills) &&
  value.skills.some((skill) => matchSkill(skill, target, shortName));

export const codexAgentMessage = (value: unknown): string => {
  if (!isRecord(value) || value.type !== "item.completed") return "";
  const item = value.item;
  return isRecord(item) &&
    item.type === "agent_message" &&
    typeof item.text === "string"
    ? item.text
    : "";
};

/** How many answer characters a streamed JSONL line contributes for its harness. */
export const streamedAnswerLength = (
  harness: "claude" | "codex",
  event: unknown,
): number =>
  harness === "claude"
    ? streamTextDeltaLength(event)
    : codexAgentMessage(event).length;

export const jsonLines = (stdout: string): readonly Record<string, unknown>[] =>
  stdout.split(/\r?\n/).flatMap((rawLine) => {
    const line = rawLine.trim();
    if (line.length === 0) return [];
    try {
      const value: unknown = JSON.parse(line);
      return isRecord(value) ? [value] : [];
    } catch {
      return [];
    }
  });

export interface ParsedGeneration {
  readonly available: boolean;
  readonly durationMs: number;
  readonly invoked: boolean;
  readonly text: string;
  readonly usage: unknown;
}

export const parseClaudeGeneration = (
  stdout: string,
  target: string,
  shortName: string,
): ParsedGeneration => {
  let text = "";
  let usage: unknown = {};
  let durationMs = 0;
  let invoked = false;
  let available = false;
  for (const value of jsonLines(stdout)) {
    available ||= skillAvailable(value, target, shortName);
    invoked ||= skillInvoked(value, target, shortName);
    if (value.type !== "result") continue;
    text = typeof value.result === "string" ? value.result : "";
    usage = value.usage ?? {};
    durationMs = typeof value.duration_ms === "number" ? value.duration_ms : 0;
  }
  return {available, durationMs, invoked, text, usage};
};

export const parseCodexGeneration = (
  stdout: string,
  durationMs: number,
): ParsedGeneration => {
  let text = "";
  let usage: unknown = {};
  for (const value of jsonLines(stdout)) {
    const message = codexAgentMessage(value);
    if (message.length > 0) text = message;
    if (value.type === "turn.completed") usage = value.usage ?? {};
  }
  const invoked = text.startsWith(CODEX_SKILL_SIGNAL);
  return {
    available: false,
    durationMs,
    invoked,
    text: text.replace(new RegExp(String.raw`^${CODEX_SKILL_SIGNAL}\s*`), ""),
    usage,
  };
};

/**
 * The judge verdict a harness wrote, before it is checked against the assertions.
 * Claude answers with one structured object; Codex streams agent messages and the
 * last one that parses carries the verdict.
 */
export const parseJudgeVerdict = (
  harness: "claude" | "codex",
  stdout: string,
): unknown => {
  if (harness === "claude") {
    try {
      const outer: unknown = JSON.parse(stdout);
      const structured = isRecord(outer) ? outer.structured_output : null;
      const resultText =
        isRecord(outer) && typeof outer.result === "string" ? outer.result : "";
      return isRecord(structured) ? structured : extractJson(resultText);
    } catch {
      return null;
    }
  }
  let verdict: unknown = null;
  for (const line of stdout.trim().split(/\r?\n/)) {
    try {
      const message = codexAgentMessage(JSON.parse(line));
      if (message.length > 0) verdict = extractJson(message);
    } catch {
      verdict = null;
    }
  }
  return verdict;
};
