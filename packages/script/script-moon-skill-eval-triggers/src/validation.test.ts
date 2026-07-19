import {describe, expect, it} from "vitest";
import {parseQueries, parseTriggerOptions} from "./validation.js";

describe("trigger query validation", () => {
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
    [{query: "query", should_trigger: "false"}],
    [{query: "", should_trigger: true}],
    [{query: "query", should_trigger: true, split: "production"}],
    {query: "query", should_trigger: true},
  ])("rejects malformed query data: %o", (input) => {
    expect(parseQueries(input).success).toBe(false);
  });
});

describe("trigger numeric options", () => {
  it("accepts positive runs, a probability threshold, and zero budget", () => {
    expect(
      parseTriggerOptions({runs: "3", threshold: "0.5", budget: "0"}),
    ).toEqual({
      success: true,
      data: {runs: 3, threshold: 0.5, budget: 0},
    });
  });

  it.each([
    {runs: "0", threshold: "0.5", budget: "0.1"},
    {runs: "2.5", threshold: "0.5", budget: "0.1"},
    {runs: "1", threshold: "1.1", budget: "0.1"},
    {runs: "1", threshold: "NaN", budget: "0.1"},
    {runs: "1", threshold: "0.5", budget: "-1"},
  ])("rejects invalid options: %o", (input) => {
    expect(parseTriggerOptions(input).success).toBe(false);
  });
});
