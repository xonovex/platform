import {isRecord} from "@xonovex/script-moon-common/records";
import {
  processFailure,
  TRIGGER_OUTPUT_LIMIT,
  type TriggerExit,
  type TriggerOutcome,
} from "../trigger.js";

export const CODEX_TRIGGER_SIGNAL = "XONOVEX_SKILL_TRIGGERED";

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
  const failure = processFailure("codex", exit);
  if (failure !== null) {
    return {triggered: false, error: failure};
  }
  return {triggered: false, error: null, selected: "none"};
};
