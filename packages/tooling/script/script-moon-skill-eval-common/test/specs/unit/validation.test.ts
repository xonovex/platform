import {describe, expect, it} from "vitest";
import {
  parseQueries,
  parseQuerySplit,
  parseTriggerOptions,
  selectQueries,
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
