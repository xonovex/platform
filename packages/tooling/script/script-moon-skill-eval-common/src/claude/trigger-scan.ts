import {isRecord} from "@xonovex/script-moon-common/records";
import {
  processFailure,
  TRIGGER_OUTPUT_LIMIT,
  type TriggerExit,
  type TriggerOutcome,
} from "../trigger.js";
import {streamTextDeltaLength} from "./stream-events.js";

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
export interface ClaudeTriggerScan {
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

export const initialClaudeTriggerScan: ClaudeTriggerScan = {
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
export const isClaudeTriggerScanSettled = (scan: ClaudeTriggerScan): boolean =>
  scan.matched || scan.competingSkill !== null || scan.outputLimitExceeded;

/** Folds one stdout line into the scan, leaving a settled scan untouched. */
export const scanClaudeTriggerLine = (
  scan: ClaudeTriggerScan,
  raw: string,
  target: string,
  short: string,
): ClaudeTriggerScan => {
  if (isClaudeTriggerScanSettled(scan)) {
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
export const resolveClaudeTriggerOutcome = (
  scan: ClaudeTriggerScan,
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
  const failure = processFailure("claude", exit);
  if (failure !== null) {
    return {triggered: false, error: failure};
  }
  if (!scan.targetAvailable) {
    return {triggered: false, error: "target skill unavailable"};
  }
  return {triggered: false, error: null, selected: "none"};
};
