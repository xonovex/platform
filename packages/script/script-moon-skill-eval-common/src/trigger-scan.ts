import {isRecord} from "@xonovex/script-moon-common/records";
import {streamTextDeltaLength} from "./validation.js";

// A run that never invokes the skill would otherwise answer the whole query, which
// dominates trigger-eval spend. Once this many answer characters have streamed with
// no invocation, the routing outcome is settled — the response was written without
// the skill — so the run ends early and scores as a non-trigger. It is conclusive
// evidence, not an infrastructure failure: a query whose plain answer is long (any
// "write this code" negative) would otherwise invalidate its skill's whole sweep on
// every attempt, since the length is a property of the query rather than a flake.
export const TRIGGER_OUTPUT_LIMIT = 2000;

export const CODEX_TRIGGER_SIGNAL = "XONOVEX_SKILL_TRIGGERED";

export interface TriggerOutcome {
  readonly triggered: boolean;
  readonly error: string | null;
  // How the run ended, recorded per run so a later triage reads the evidence instead
  // of re-probing the model to recover which skill answered.
  readonly selected?: string;
}

/** How a harness process ended, as the outcome resolvers read it. */
export interface TriggerExit {
  readonly code: number | null;
  readonly stderr: string;
  readonly timedOut: boolean;
  readonly spawnError: string | null;
}

const matchSkill = (
  skillField: unknown,
  target: string,
  short: string,
): boolean => {
  if (typeof skillField !== "string") {
    return false;
  }
  return (
    skillField === target ||
    skillField === short ||
    skillField.endsWith(`:${short}`)
  );
};

// Every skill the line invokes, whether the harness allowed the call or denied it.
// The names matter beyond the target's own: a line that invokes some other skill
// settles the routing outcome just as firmly as one that invokes the target.
export const invokedSkillNames = (line: string): readonly string[] => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return [];
  }
  if (!isRecord(obj)) {
    return [];
  }

  const names: string[] = [];
  const message = obj.message;
  if (isRecord(message)) {
    const content = Array.isArray(message.content) ? message.content : [];
    for (const item of content) {
      if (isRecord(item) && item.type === "tool_use" && item.name === "Skill") {
        const inputField = item.input;
        if (isRecord(inputField) && typeof inputField.skill === "string") {
          names.push(inputField.skill);
        }
      }
    }
  }

  const denials = Array.isArray(obj.permission_denials)
    ? obj.permission_denials
    : [];
  for (const denial of denials) {
    if (isRecord(denial) && denial.tool_name === "Skill") {
      const toolInput = denial.tool_input;
      if (isRecord(toolInput) && typeof toolInput.skill === "string") {
        names.push(toolInput.skill);
      }
    }
  }

  return names;
};

const skillAvailableLine = (
  line: string,
  target: string,
  short: string,
): boolean => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return false;
  }
  if (!isRecord(obj) || obj.type !== "system" || obj.subtype !== "init") {
    return false;
  }
  const skills = Array.isArray(obj.skills) ? obj.skills : [];
  return skills.some((skill) => matchSkill(skill, target, short));
};

const textDeltaLength = (line: string): number => {
  try {
    return streamTextDeltaLength(JSON.parse(line));
  } catch {
    return 0;
  }
};

// What the harness did with a Skill call. An invocation naming something the harness
// cannot resolve launches nothing, so it says nothing about which skill the request
// belongs to; only a launch settles that. The two are indistinguishable from the
// tool_use line alone, which carries whatever name the model wrote.
export type ToolResultOutcome = "launched" | "rejected";

export const toolResultOutcome = (line: string): ToolResultOutcome | null => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return null;
  }
  if (!isRecord(obj)) {
    return null;
  }
  const message = obj.message;
  if (!isRecord(message)) {
    return null;
  }
  const content = Array.isArray(message.content) ? message.content : [];
  for (const item of content) {
    if (!isRecord(item) || item.type !== "tool_result") {
      continue;
    }
    if (item.is_error === true) {
      return "rejected";
    }
    const text =
      typeof item.content === "string"
        ? item.content
        : JSON.stringify(item.content);
    return text.includes("tool_use_error") ? "rejected" : "launched";
  }
  return null;
};

/** What a Claude run has established so far, carried line to line. */
export interface TriggerScan {
  readonly matched: boolean;
  readonly targetAvailable: boolean;
  // A skill some other name invoked, held until a later line says whether the
  // harness launched it. Holding it is what separates a lost request from a name
  // the harness could not resolve.
  readonly pendingSkill: string | null;
  readonly competingSkill: string | null;
  readonly outputLimitExceeded: boolean;
  readonly outputCharacters: number;
}

export const initialTriggerScan: TriggerScan = {
  matched: false,
  targetAvailable: false,
  pendingSkill: null,
  competingSkill: null,
  outputLimitExceeded: false,
  outputCharacters: 0,
};

/**
 * Whether the scan has decided the routing outcome. A settled run needs no further
 * turns, so the caller stops reading and ends the process.
 */
export const isTriggerScanSettled = (scan: TriggerScan): boolean =>
  scan.matched || scan.competingSkill !== null || scan.outputLimitExceeded;

/** Folds one stdout line into the scan, leaving a settled scan untouched. */
export const scanTriggerLine = (
  scan: TriggerScan,
  raw: string,
  target: string,
  short: string,
): TriggerScan => {
  if (isTriggerScanSettled(scan)) {
    return scan;
  }
  const line = raw.trim();
  if (!line) {
    return scan;
  }
  const targetAvailable =
    scan.targetAvailable || skillAvailableLine(line, target, short);
  const invoked = invokedSkillNames(line);
  if (invoked.some((skill) => matchSkill(skill, target, short))) {
    return {...scan, targetAvailable, matched: true};
  }
  // Some other name was invoked. Whether it settles the run depends on what the
  // harness did with it, which arrives on a later line, so hold it and read on.
  if (invoked.length > 0) {
    return {...scan, targetAvailable, pendingSkill: invoked[0] ?? ""};
  }
  if (scan.pendingSkill !== null) {
    const outcome = toolResultOutcome(line);
    if (outcome === "launched") {
      // Another skill answered the request, so the target lost it. Ending here
      // spends no further turns on a settled result.
      return {...scan, targetAvailable, competingSkill: scan.pendingSkill};
    }
    // A rejected call resolved nothing, so no skill answered and the run is not
    // settled. The model still has a turn to name a skill that exists.
    return {
      ...scan,
      targetAvailable,
      pendingSkill: outcome === "rejected" ? null : scan.pendingSkill,
    };
  }
  const outputCharacters = scan.outputCharacters + textDeltaLength(line);
  return {
    ...scan,
    targetAvailable,
    outputCharacters,
    outputLimitExceeded: outputCharacters > TRIGGER_OUTPUT_LIMIT,
  };
};

/** Scores a finished Claude run from what it streamed and how it exited. */
export const resolveTriggerOutcome = (
  scan: TriggerScan,
  exit: TriggerExit,
): TriggerOutcome => {
  if (scan.matched) {
    return {triggered: true, error: null, selected: "target"};
  }
  if (exit.timedOut) {
    return {triggered: false, error: "timeout"};
  }
  if (scan.competingSkill !== null || scan.outputLimitExceeded) {
    const selected =
      scan.competingSkill === null
        ? "output-limit"
        : `competitor:${scan.competingSkill}`;
    return scan.targetAvailable
      ? {triggered: false, error: null, selected}
      : {triggered: false, error: "target skill unavailable"};
  }
  if (exit.spawnError !== null) {
    return {triggered: false, error: exit.spawnError};
  }
  if (exit.code !== 0) {
    const detail = exit.stderr.trim();
    const detailSuffix = detail.length > 0 ? `: ${detail}` : "";
    return {
      triggered: false,
      error: `claude exited ${String(exit.code)}${detailSuffix}`,
    };
  }
  if (!scan.targetAvailable) {
    return {triggered: false, error: "target skill unavailable"};
  }
  return {triggered: false, error: null, selected: "none"};
};

const codexAgentMessage = (line: string): string => {
  try {
    const event: unknown = JSON.parse(line);
    if (!isRecord(event) || event.type !== "item.completed") return "";
    const item = event.item;
    return isRecord(item) &&
      item.type === "agent_message" &&
      typeof item.text === "string"
      ? item.text
      : "";
  } catch {
    return "";
  }
};

/**
 * What a Codex run has established so far. Codex signals applicability in its agent
 * message rather than by invoking a skill, so it has no competing-skill state.
 */
export interface CodexTriggerScan {
  readonly matched: boolean;
  readonly outputLimitExceeded: boolean;
  readonly outputCharacters: number;
}

export const initialCodexTriggerScan: CodexTriggerScan = {
  matched: false,
  outputLimitExceeded: false,
  outputCharacters: 0,
};

export const isCodexTriggerScanSettled = (scan: CodexTriggerScan): boolean =>
  scan.matched || scan.outputLimitExceeded;

/** Folds one Codex JSONL line into the scan, leaving a settled scan untouched. */
export const scanCodexTriggerLine = (
  scan: CodexTriggerScan,
  raw: string,
): CodexTriggerScan => {
  if (isCodexTriggerScanSettled(scan)) {
    return scan;
  }
  const message = codexAgentMessage(raw.trim());
  const outputCharacters = scan.outputCharacters + message.length;
  if (message.includes(CODEX_TRIGGER_SIGNAL)) {
    return {...scan, outputCharacters, matched: true};
  }
  return {
    ...scan,
    outputCharacters,
    outputLimitExceeded: outputCharacters > TRIGGER_OUTPUT_LIMIT,
  };
};

// The outcomes carry the same `selected` evidence the Claude path records, so a
// triage reads how a Codex run ended from its result rather than re-probing the
// model. Codex has no competitor outcome: it signals applicability in the agent
// message instead of invoking a skill, so no other skill can win the request.
export const resolveCodexTriggerOutcome = (
  scan: CodexTriggerScan,
  exit: TriggerExit,
): TriggerOutcome => {
  if (scan.matched) {
    return {triggered: true, error: null, selected: "target"};
  }
  if (exit.timedOut) {
    return {triggered: false, error: "timeout"};
  }
  if (scan.outputLimitExceeded) {
    return {triggered: false, error: null, selected: "output-limit"};
  }
  if (exit.spawnError !== null) {
    return {triggered: false, error: exit.spawnError};
  }
  if (exit.code !== 0) {
    const detail = exit.stderr.trim();
    const suffix = detail.length > 0 ? `: ${detail}` : "";
    return {
      triggered: false,
      error: `codex exited ${String(exit.code)}${suffix}`,
    };
  }
  return {triggered: false, error: null, selected: "none"};
};
