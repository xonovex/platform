import {describe, expect, it} from "vitest";
import {
  initialCodexTriggerScan,
  initialTriggerScan,
  invokedSkillNames,
  isCodexTriggerScanSettled,
  isTriggerScanSettled,
  resolveCodexTriggerOutcome,
  resolveTriggerOutcome,
  scanCodexTriggerLine,
  scanTriggerLine,
  toolResultOutcome,
  TRIGGER_OUTPUT_LIMIT,
  type TriggerExit,
} from "../../../src/trigger-scan.js";

const TARGET = "plugin:test-skill";
const SHORT = "test-skill";

const CLEAN_EXIT: TriggerExit = {
  code: 0,
  stderr: "",
  timedOut: false,
  spawnError: null,
};

const init = (...skills: readonly string[]): string =>
  JSON.stringify({type: "system", subtype: "init", skills});

const toolUse = (...skills: readonly string[]): string =>
  JSON.stringify({
    message: {
      content: skills.map((skill) => ({
        type: "tool_use",
        name: "Skill",
        input: {skill},
      })),
    },
  });

const toolResult = (content: string): string =>
  JSON.stringify({message: {content: [{type: "tool_result", content}]}});

const toolUseError = (content: string): string =>
  JSON.stringify({
    message: {content: [{type: "tool_result", is_error: true, content}]},
  });

const denial = (skill: string): string =>
  JSON.stringify({
    permission_denials: [{tool_name: "Skill", tool_input: {skill}}],
  });

const textDelta = (characters: number): string =>
  JSON.stringify({
    type: "stream_event",
    event: {
      type: "content_block_delta",
      delta: {type: "text_delta", text: "x".repeat(characters)},
    },
  });

const score = (lines: readonly string[], exit: TriggerExit = CLEAN_EXIT) =>
  resolveTriggerOutcome(
    lines.reduce(
      (scan, line) => scanTriggerLine(scan, line, TARGET, SHORT),
      initialTriggerScan,
    ),
    exit,
  );

describe("scanTriggerLine", () => {
  it("settles as soon as the target is invoked, so the caller can end the run", () => {
    const scan = [init(TARGET), toolUse(SHORT)].reduce(
      (state, line) => scanTriggerLine(state, line, TARGET, SHORT),
      initialTriggerScan,
    );

    expect(isTriggerScanSettled(scan)).toBe(true);
    expect(scan.matched).toBe(true);
  });

  it("leaves a settled scan untouched when further lines arrive", () => {
    const settled = scanTriggerLine(
      initialTriggerScan,
      toolUse(SHORT),
      TARGET,
      SHORT,
    );

    expect(
      scanTriggerLine(settled, toolUse("other-skill"), TARGET, SHORT),
    ).toBe(settled);
  });

  it("ignores blank lines", () => {
    expect(
      scanTriggerLine(initialTriggerScan, " ".repeat(3), TARGET, SHORT),
    ).toBe(initialTriggerScan);
  });

  it("matches the target by full name, short name, and plugin suffix", () => {
    for (const name of [TARGET, SHORT, `other-plugin:${SHORT}`]) {
      expect(
        scanTriggerLine(initialTriggerScan, toolUse(name), TARGET, SHORT)
          .matched,
      ).toBe(true);
    }
  });
});

describe("resolveTriggerOutcome", () => {
  it("detects a matching Skill tool invocation", () => {
    expect(score([init(TARGET), toolUse(SHORT)])).toEqual({
      triggered: true,
      error: null,
      selected: "target",
    });
  });

  it("scores a competing skill that the harness launched as a non-trigger", () => {
    expect(
      score([
        init(TARGET),
        toolUse("other-skill"),
        toolResult("Launching skill: other-skill"),
      ]),
    ).toEqual({
      triggered: false,
      error: null,
      selected: "competitor:other-skill",
    });
  });

  it("does not settle the run when the harness cannot resolve the skill name", () => {
    expect(
      score([
        init(TARGET),
        toolUse("plugin"),
        toolUseError("<tool_use_error>Unknown skill: plugin</tool_use_error>"),
        toolUse(SHORT),
      ]),
    ).toEqual({triggered: true, error: null, selected: "target"});
  });

  it("scores an unresolved skill name that is never corrected as a clean negative", () => {
    expect(
      score([
        init(TARGET),
        toolUse("plugin"),
        toolUseError("<tool_use_error>Unknown skill: plugin</tool_use_error>"),
      ]),
    ).toEqual({triggered: false, error: null, selected: "none"});
  });

  it("scores a denied competing skill invocation as a non-trigger", () => {
    expect(
      score([
        init(TARGET),
        denial("other-skill"),
        toolResult("Launching skill: other-skill"),
      ]),
    ).toEqual({
      triggered: false,
      error: null,
      selected: "competitor:other-skill",
    });
  });

  it("prefers the target when it is invoked alongside a competing skill", () => {
    expect(score([init(TARGET), toolUse("other-skill", SHORT)])).toEqual({
      triggered: true,
      error: null,
      selected: "target",
    });
  });

  it("returns a clean negative result when the skill is available", () => {
    expect(score([init(SHORT), "not json"])).toEqual({
      triggered: false,
      error: null,
      selected: "none",
    });
  });

  it("reports when the target skill is unavailable", () => {
    expect(score([init("other-skill")])).toEqual({
      triggered: false,
      error: "target skill unavailable",
    });
  });

  it("includes process stderr in nonzero-exit failures", () => {
    expect(
      score([], {
        code: 3,
        stderr: "authentication failed\n",
        timedOut: false,
        spawnError: null,
      }),
    ).toEqual({
      triggered: false,
      error: "claude exited 3: authentication failed",
    });
  });

  it("omits the detail suffix when a nonzero exit wrote no stderr", () => {
    expect(
      score([], {code: 3, stderr: "  ", timedOut: false, spawnError: null}),
    ).toEqual({triggered: false, error: "claude exited 3"});
  });

  it("reports a timeout ahead of every other ending", () => {
    expect(
      score([init("other-skill")], {
        code: null,
        stderr: "",
        timedOut: true,
        spawnError: null,
      }),
    ).toEqual({triggered: false, error: "timeout"});
  });

  it("reports a spawn failure when nothing streamed", () => {
    expect(
      score([], {
        code: null,
        stderr: "",
        timedOut: false,
        spawnError: "spawn claude ENOENT",
      }),
    ).toEqual({triggered: false, error: "spawn claude ENOENT"});
  });

  it("scores a response that exceeds the output limit as a non-trigger", () => {
    expect(score([init(SHORT), textDelta(TRIGGER_OUTPUT_LIMIT + 1)])).toEqual({
      triggered: false,
      error: null,
      selected: "output-limit",
    });
  });

  it("still reports an unavailable skill when the output limit is exceeded", () => {
    expect(
      score([init("other-skill"), textDelta(TRIGGER_OUTPUT_LIMIT + 1)]),
    ).toEqual({triggered: false, error: "target skill unavailable"});
  });

  it("keeps reading while streamed output stays within the limit", () => {
    expect(
      score([init(SHORT), textDelta(TRIGGER_OUTPUT_LIMIT), toolUse(SHORT)]),
    ).toEqual({triggered: true, error: null, selected: "target"});
  });
});

describe("invokedSkillNames", () => {
  it("returns nothing for malformed and non-object lines", () => {
    for (const line of ["not json", "[]", '"text"']) {
      expect(invokedSkillNames(line)).toEqual([]);
    }
  });

  it("collects both invoked and denied skill names", () => {
    expect(
      invokedSkillNames(
        JSON.stringify({
          message: {
            content: [
              {type: "tool_use", name: "Skill", input: {skill: "first"}},
              {type: "tool_use", name: "Read", input: {skill: "ignored"}},
            ],
          },
          permission_denials: [
            {tool_name: "Skill", tool_input: {skill: "second"}},
            {tool_name: "Bash", tool_input: {skill: "ignored"}},
          ],
        }),
      ),
    ).toEqual(["first", "second"]);
  });
});

describe("toolResultOutcome", () => {
  it("returns null for lines that carry no tool result", () => {
    for (const line of [
      "not json",
      "[]",
      JSON.stringify({message: 1}),
      init(),
    ]) {
      expect(toolResultOutcome(line)).toBeNull();
    }
  });

  it("reads a launch, an error flag, and an embedded tool_use_error", () => {
    expect(toolResultOutcome(toolResult("Launching skill: other"))).toBe(
      "launched",
    );
    expect(toolResultOutcome(toolUseError("denied"))).toBe("rejected");
    expect(toolResultOutcome(toolResult("<tool_use_error>x"))).toBe("rejected");
  });
});

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
