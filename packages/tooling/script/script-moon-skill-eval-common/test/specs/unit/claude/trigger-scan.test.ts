import {describe, expect, it} from "vitest";
import {
  initialClaudeTriggerScan,
  invokedSkillNames,
  isClaudeTriggerScanSettled,
  resolveClaudeTriggerOutcome,
  scanClaudeTriggerLine,
  toolResultOutcome,
} from "../../../../src/claude/trigger-scan.js";
import {
  TRIGGER_OUTPUT_LIMIT,
  type TriggerExit,
} from "../../../../src/trigger.js";

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
  resolveClaudeTriggerOutcome(
    lines.reduce(
      (scan, line) => scanClaudeTriggerLine(scan, line, TARGET, SHORT),
      initialClaudeTriggerScan,
    ),
    exit,
  );

describe("scanClaudeTriggerLine", () => {
  it("settles as soon as the target is invoked, so the caller can end the run", () => {
    const scan = [init(TARGET), toolUse(SHORT)].reduce(
      (state, line) => scanClaudeTriggerLine(state, line, TARGET, SHORT),
      initialClaudeTriggerScan,
    );

    expect(isClaudeTriggerScanSettled(scan)).toBe(true);
    expect(scan.matched).toBe(true);
  });

  it("leaves a settled scan untouched when further lines arrive", () => {
    const settled = scanClaudeTriggerLine(
      initialClaudeTriggerScan,
      toolUse(SHORT),
      TARGET,
      SHORT,
    );

    expect(
      scanClaudeTriggerLine(settled, toolUse("other-skill"), TARGET, SHORT),
    ).toBe(settled);
  });

  it("ignores blank lines", () => {
    expect(
      scanClaudeTriggerLine(
        initialClaudeTriggerScan,
        " ".repeat(3),
        TARGET,
        SHORT,
      ),
    ).toBe(initialClaudeTriggerScan);
  });

  it("matches the target by full name, short name, and plugin suffix", () => {
    for (const name of [TARGET, SHORT, `other-plugin:${SHORT}`]) {
      expect(
        scanClaudeTriggerLine(
          initialClaudeTriggerScan,
          toolUse(name),
          TARGET,
          SHORT,
        ).matched,
      ).toBe(true);
    }
  });
});

describe("resolveClaudeTriggerOutcome", () => {
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
