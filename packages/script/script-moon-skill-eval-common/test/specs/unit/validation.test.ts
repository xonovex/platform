import {describe, expect, it} from "vitest";
import {
  buildIsolatedClaudeArgs,
  buildIsolatedCodexArgs,
  buildTriggerClaudeArgs,
  parseQueries,
  parseQuerySplit,
  parseTriggerOptions,
  selectQueries,
  streamTextDeltaLength,
  TRIGGER_MAX_TURNS,
  triggerModelRunCount,
} from "../../../src/validation.js";

describe("trigger query validation", () => {
  it.each(["train", "validation", "all"])(
    "accepts the supported query split %s",
    (split) => {
      expect(parseQuerySplit(split)).toEqual({success: true, data: split});
    },
  );

  it("rejects an unsupported query split", () => {
    expect(parseQuerySplit("production").success).toBe(false);
  });

  it("accepts exact booleans", () => {
    expect(
      parseQueries([
        {
          query: "Should this skill run?",
          should_trigger: false,
          rationale: "No relevant task",
          split: "validation",
        },
      ]),
    ).toEqual({
      success: true,
      data: [
        {
          query: "Should this skill run?",
          should_trigger: false,
          rationale: "No relevant task",
          split: "validation",
        },
      ],
    });
  });

  it.each([
    [],
    [{query: "query", should_trigger: "false"}],
    [{query: "", should_trigger: true}],
    [{query: "query", should_trigger: true, split: "production"}],
    {query: "query", should_trigger: true},
  ])("rejects malformed query data: %o", (input) => {
    expect(parseQueries(input).success).toBe(false);
  });

  it("rejects a split with no matching queries", () => {
    const queries = [
      {
        query: "Training example",
        should_trigger: true,
        rationale: "Relevant",
        split: "train" as const,
      },
    ];

    expect(selectQueries(queries, "validation")).toEqual({
      success: false,
      error: "split 'validation' has no queries",
    });
  });
});

describe("trigger numeric options", () => {
  it("accepts bounded runs, batching, a probability threshold, and a positive budget", () => {
    expect(
      parseTriggerOptions({
        runs: "3",
        threshold: "0.5",
        budget: "0.05",
        batchSize: "8",
      }),
    ).toEqual({
      success: true,
      data: {runs: 3, threshold: 0.5, budget: 0.05, batchSize: 8},
    });
  });

  it.each([
    {runs: "0", threshold: "0.5", budget: "0.05"},
    {runs: "4", threshold: "0.5", budget: "0.05"},
    {runs: "2.5", threshold: "0.5", budget: "0.05"},
    {runs: "1", threshold: "1.1", budget: "0.05"},
    {runs: "1", threshold: "NaN", budget: "0.05"},
    {runs: "1", threshold: "0.5", budget: "-1"},
    {runs: "1", threshold: "0.5", budget: "0"},
    {runs: "1", threshold: "0.5", budget: "0.06"},
    {runs: "1", threshold: "0.5", budget: "0.05", batchSize: "0"},
  ])("rejects invalid options: %o", (input) => {
    expect(parseTriggerOptions(input).success).toBe(false);
  });

  it("counts every query repetition as one model run", () => {
    expect(triggerModelRunCount(8, 3)).toBe(24);
    expect(triggerModelRunCount(9, 3)).toBe(27);
  });
});

// Every Claude and Codex eval invocation across the trigger, routing, and output
// harnesses is built from these two argument sets. A flag dropped here silently
// lets a run read the machine's real settings, so the whole set is pinned.
describe("shared harness isolation", () => {
  it("denies a Claude run the machine's settings, MCP servers, and session", () => {
    const args = buildIsolatedClaudeArgs("stream-json");

    expect(args).toEqual([
      "-p",
      "--output-format",
      "stream-json",
      "--setting-sources",
      "",
      "--strict-mcp-config",
      "--mcp-config",
      '{"mcpServers":{}}',
      "--no-session-persistence",
      "--no-chrome",
    ]);
  });

  it("carries the requested output format", () => {
    const args = buildIsolatedClaudeArgs("json");

    expect(args.slice(0, 3)).toEqual(["-p", "--output-format", "json"]);
  });

  it("denies a Codex run user config, rules, and write access", () => {
    expect(buildIsolatedCodexArgs({model: ""})).toEqual([
      "exec",
      "--json",
      "--ephemeral",
      "--sandbox",
      "read-only",
      "--ignore-user-config",
      "--ignore-rules",
      "--skip-git-repo-check",
    ]);
  });

  it("appends only a non-empty Codex model", () => {
    expect(buildIsolatedCodexArgs({model: "gpt-5"}).slice(-2)).toEqual([
      "--model",
      "gpt-5",
    ]);
  });
});

describe("trigger process isolation", () => {
  it("exposes only Skill with bounded isolated execution", () => {
    const args = buildTriggerClaudeArgs({
      model: "haiku",
      budget: 0.05,
      pluginDirectories: ["dependency-directory", "plugin-directory"],
    });
    const toolIndex = args.indexOf("--tools");
    const settingIndex = args.indexOf("--setting-sources");

    expect(args.slice(toolIndex, toolIndex + 2)).toEqual(["--tools", "Skill"]);
    expect(args.slice(settingIndex, settingIndex + 2)).toEqual([
      "--setting-sources",
      "",
    ]);
    expect(args).toContain("--strict-mcp-config");
    expect(args).toContain("--no-session-persistence");
    expect(args).toContain("--max-budget-usd");
    expect(args).toContain("0.05");
    const turnsIndex = args.indexOf("--max-turns");
    expect(args.slice(turnsIndex, turnsIndex + 2)).toEqual([
      "--max-turns",
      String(TRIGGER_MAX_TURNS),
    ]);
    expect(args).not.toContain("Read");
    expect(args).not.toContain("Bash");
    expect(args.filter((argument) => argument === "--plugin-dir")).toHaveLength(
      2,
    );
  });

  it("counts only streamed response text", () => {
    expect(
      streamTextDeltaLength({
        type: "stream_event",
        event: {
          type: "content_block_delta",
          delta: {type: "text_delta", text: "no skill"},
        },
      }),
    ).toBe(8);
    expect(streamTextDeltaLength({type: "result"})).toBe(0);
  });
});
