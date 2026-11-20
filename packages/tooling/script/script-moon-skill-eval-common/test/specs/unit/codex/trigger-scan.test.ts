import {describe, expect, it} from "vitest";
import {
  initialCodexTriggerScan,
  isCodexTriggerScanSettled,
  resolveCodexTriggerOutcome,
  scanCodexTriggerLine,
} from "../../../../src/codex/trigger-scan.js";
import {
  TRIGGER_OUTPUT_LIMIT,
  type TriggerExit,
} from "../../../../src/trigger.js";

const CLEAN_EXIT: TriggerExit = {
  code: 0,
  stderr: "",
  timedOut: false,
  spawnError: null,
};

const codexMessage = (text: string): string =>
  JSON.stringify({type: "item.completed", item: {type: "agent_message", text}});

const scoreCodex = (lines: readonly string[], exit: TriggerExit = CLEAN_EXIT) =>
  resolveCodexTriggerOutcome(
    lines.reduce(scanCodexTriggerLine, initialCodexTriggerScan),
    exit,
  );

describe("scanCodexTriggerLine", () => {
  it("settles on the isolated skill signal", () => {
    const scan = scanCodexTriggerLine(
      initialCodexTriggerScan,
      codexMessage("XONOVEX_SKILL_TRIGGERED"),
    );

    expect(isCodexTriggerScanSettled(scan)).toBe(true);
  });

  it("leaves a settled scan untouched when further lines arrive", () => {
    const settled = scanCodexTriggerLine(
      initialCodexTriggerScan,
      codexMessage("XONOVEX_SKILL_TRIGGERED"),
    );

    expect(scanCodexTriggerLine(settled, codexMessage("more"))).toBe(settled);
  });

  it("counts only completed agent messages toward the output limit", () => {
    const scan = [
      "not json",
      JSON.stringify({type: "item.started"}),
      JSON.stringify({type: "item.completed", item: {type: "reasoning"}}),
    ].reduce(scanCodexTriggerLine, initialCodexTriggerScan);

    expect(scan.outputCharacters).toBe(0);
  });
});

describe("resolveCodexTriggerOutcome", () => {
  it("detects the isolated skill signal in a Codex JSONL agent message", () => {
    expect(scoreCodex([codexMessage("XONOVEX_SKILL_TRIGGERED")])).toEqual({
      triggered: true,
      error: null,
      selected: "target",
    });
  });

  it("returns a clean negative result for malformed and nonmatching output", () => {
    expect(scoreCodex(["not json", codexMessage("NOT_APPLICABLE")])).toEqual({
      triggered: false,
      error: null,
      selected: "none",
    });
  });

  it("includes process stderr in nonzero-exit failures", () => {
    expect(
      scoreCodex([], {
        code: 3,
        stderr: "authentication failed",
        timedOut: false,
        spawnError: null,
      }),
    ).toEqual({
      triggered: false,
      error: "codex exited 3: authentication failed",
    });
  });

  it("omits the detail suffix when a nonzero exit wrote no stderr", () => {
    expect(
      scoreCodex([], {code: 3, stderr: "", timedOut: false, spawnError: null}),
    ).toEqual({triggered: false, error: "codex exited 3"});
  });

  it("reports a timeout and a spawn failure", () => {
    expect(
      scoreCodex([], {
        code: null,
        stderr: "",
        timedOut: true,
        spawnError: null,
      }),
    ).toEqual({triggered: false, error: "timeout"});
    expect(
      scoreCodex([], {
        code: null,
        stderr: "",
        timedOut: false,
        spawnError: "spawn codex ENOENT",
      }),
    ).toEqual({triggered: false, error: "spawn codex ENOENT"});
  });

  it("scores a response that exceeds the output limit as a non-trigger", () => {
    expect(
      scoreCodex([codexMessage("x".repeat(TRIGGER_OUTPUT_LIMIT + 1))]),
    ).toEqual({triggered: false, error: null, selected: "output-limit"});
  });
});
