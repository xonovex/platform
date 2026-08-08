import {describe, expect, it} from "vitest";
import {
  CODEX_SKILL_SIGNAL,
  codexAgentMessage,
  jsonLines,
  parseClaudeGeneration,
  parseCodexGeneration,
  parseJudgeVerdict,
  skillAvailable,
  skillInvoked,
  streamedAnswerLength,
} from "../../../src/output-parse.js";

const lines = (...values: readonly unknown[]): string =>
  values.map((value) => JSON.stringify(value)).join("\n");

describe("jsonLines", () => {
  it("keeps object lines and drops blank, malformed, and non-object ones", () => {
    expect(
      jsonLines(
        ['{"a":1}', "", "  ", "not json", "[1,2]", '"text"'].join("\n"),
      ),
    ).toEqual([{a: 1}]);
  });

  it("reads output that uses carriage returns", () => {
    expect(jsonLines('{"a":1}\r\n{"b":2}')).toEqual([{a: 1}, {b: 2}]);
  });
});

const invocation = (skill: unknown): Record<string, unknown> => ({
  message: {content: [{type: "tool_use", name: "Skill", input: {skill}}]},
});

describe("skillInvoked", () => {
  it("matches the target by full name, short name, and plugin suffix", () => {
    for (const name of [
      "plugin:test-skill",
      "test-skill",
      "other:test-skill",
    ]) {
      expect(
        skillInvoked(invocation(name), "plugin:test-skill", "test-skill"),
      ).toBe(true);
    }
  });

  it("ignores another skill, another tool, and a malformed message", () => {
    expect(
      skillInvoked(invocation("other"), "plugin:test-skill", "test-skill"),
    ).toBe(false);
    expect(
      skillInvoked(
        {message: {content: [{type: "tool_use", name: "Read"}]}},
        "plugin:test-skill",
        "test-skill",
      ),
    ).toBe(false);
    expect(skillInvoked({message: 1}, "plugin:test-skill", "test-skill")).toBe(
      false,
    );
  });
});

describe("skillAvailable", () => {
  it("reads the target out of an init line only", () => {
    expect(
      skillAvailable(
        {type: "system", subtype: "init", skills: ["test-skill"]},
        "plugin:test-skill",
        "test-skill",
      ),
    ).toBe(true);
    expect(
      skillAvailable(
        {type: "system", subtype: "other", skills: ["test-skill"]},
        "plugin:test-skill",
        "test-skill",
      ),
    ).toBe(false);
    expect(
      skillAvailable({type: "system", subtype: "init"}, "x", "test-skill"),
    ).toBe(false);
  });
});

describe("codexAgentMessage", () => {
  it("returns the text of a completed agent message and nothing else", () => {
    expect(
      codexAgentMessage({
        type: "item.completed",
        item: {type: "agent_message", text: "answer"},
      }),
    ).toBe("answer");
    for (const value of [
      null,
      "text",
      {type: "item.started"},
      {type: "item.completed", item: {type: "reasoning"}},
      {type: "item.completed", item: {type: "agent_message", text: 1}},
    ]) {
      expect(codexAgentMessage(value)).toBe("");
    }
  });
});

describe("streamedAnswerLength", () => {
  it("counts a Claude text delta and a Codex agent message", () => {
    expect(
      streamedAnswerLength("claude", {
        type: "stream_event",
        event: {
          type: "content_block_delta",
          delta: {type: "text_delta", text: "abcd"},
        },
      }),
    ).toBe(4);
    expect(
      streamedAnswerLength("codex", {
        type: "item.completed",
        item: {type: "agent_message", text: "abc"},
      }),
    ).toBe(3);
  });
});

describe("parseClaudeGeneration", () => {
  it("reads the result, usage, duration, availability, and invocation", () => {
    const stdout = lines(
      {type: "system", subtype: "init", skills: ["test-skill"]},
      {
        message: {
          content: [
            {type: "tool_use", name: "Skill", input: {skill: "test-skill"}},
          ],
        },
      },
      {
        type: "result",
        result: "the answer",
        usage: {input_tokens: 3, output_tokens: 2},
        duration_ms: 12,
      },
    );

    expect(
      parseClaudeGeneration(stdout, "plugin:test-skill", "test-skill"),
    ).toEqual({
      available: true,
      durationMs: 12,
      invoked: true,
      text: "the answer",
      usage: {input_tokens: 3, output_tokens: 2},
    });
  });

  it("falls back to empty evidence when the result line is malformed", () => {
    expect(
      parseClaudeGeneration(lines({type: "result"}), "target", "short"),
    ).toEqual({
      available: false,
      durationMs: 0,
      invoked: false,
      text: "",
      usage: {},
    });
  });

  it("keeps the last result line when a run emits several", () => {
    expect(
      parseClaudeGeneration(
        lines(
          {type: "result", result: "first", duration_ms: 1},
          {type: "result", result: "second", duration_ms: 2},
        ),
        "target",
        "short",
      ),
    ).toMatchObject({text: "second", durationMs: 2});
  });
});

describe("parseCodexGeneration", () => {
  it("strips the skill signal and reports the run as skill-using", () => {
    const stdout = lines(
      {
        type: "item.completed",
        item: {
          type: "agent_message",
          text: `${CODEX_SKILL_SIGNAL}\nthe answer`,
        },
      },
      {type: "turn.completed", usage: {input_tokens: 4, output_tokens: 2}},
    );

    expect(parseCodexGeneration(stdout, 40)).toEqual({
      available: false,
      durationMs: 40,
      invoked: true,
      text: "the answer",
      usage: {input_tokens: 4, output_tokens: 2},
    });
  });

  it("reports an answer written without the signal as skill-free", () => {
    expect(
      parseCodexGeneration(
        lines({
          type: "item.completed",
          item: {type: "agent_message", text: "the answer"},
        }),
        5,
      ),
    ).toMatchObject({invoked: false, text: "the answer", usage: {}});
  });
});

describe("parseJudgeVerdict", () => {
  it("prefers Claude's structured output over its result text", () => {
    expect(
      parseJudgeVerdict(
        "claude",
        JSON.stringify({
          structured_output: {assertion_results: [{passed: true}]},
          result: '{"assertion_results":[{"passed":false}]}',
        }),
      ),
    ).toEqual({assertion_results: [{passed: true}]});
  });

  it("falls back to JSON embedded in Claude's result text", () => {
    expect(
      parseJudgeVerdict(
        "claude",
        JSON.stringify({
          result: 'noise {"assertion_results":[{"passed":true}]} more',
        }),
      ),
    ).toEqual({assertion_results: [{passed: true}]});
  });

  it("returns null when Claude wrote nothing parseable", () => {
    expect(parseJudgeVerdict("claude", "not json")).toBeNull();
  });

  it("takes the last parseable Codex agent message", () => {
    expect(
      parseJudgeVerdict(
        "codex",
        lines(
          {
            type: "item.completed",
            item: {
              type: "agent_message",
              text: '{"assertion_results":[{"passed":false}]}',
            },
          },
          {
            type: "item.completed",
            item: {
              type: "agent_message",
              text: '{"assertion_results":[{"passed":true}]}',
            },
          },
        ),
      ),
    ).toEqual({assertion_results: [{passed: true}]});
  });
});
