import {describe, expect, it} from "vitest";
import {
  evalEntries,
  normalizeEval,
  parseJudgeResults,
  parseOutputOptions,
} from "./validation.js";

describe("output eval validation", () => {
  it("normalizes a valid eval without coercing field types", () => {
    expect(
      normalizeEval(
        {
          id: "case-1",
          prompt: "Explain the result",
          assertions: ["Includes evidence"],
          files: ["fixtures/input.txt"],
        },
        1,
      ),
    ).toEqual({
      success: true,
      data: {
        id: "case-1",
        prompt: "Explain the result",
        expected_output: "",
        assertions: ["Includes evidence"],
        files: ["fixtures/input.txt"],
      },
    });
  });

  it.each([
    {prompt: false, assertions: ["valid"]},
    {prompt: "valid", assertions: [false]},
    {id: "../escape", prompt: "valid", assertions: ["valid"]},
    {prompt: "valid", assertions: ["valid"], files: ["../secret"]},
  ])("rejects malformed eval input: %o", (input) => {
    expect(normalizeEval(input, 1).success).toBe(false);
  });

  it("accepts both supported eval file envelopes", () => {
    expect(evalEntries([{}])).toEqual({success: true, data: [{}]});
    expect(evalEntries({evals: [{}]})).toEqual({success: true, data: [{}]});
    expect(evalEntries({evals: "wrong"}).success).toBe(false);
  });
});

describe("judge validation", () => {
  it("requires a boolean passed verdict", () => {
    expect(
      parseJudgeResults({
        assertion_results: [{passed: false, evidence: "missing"}],
      }),
    ).toEqual([{passed: false, evidence: "missing"}]);
    expect(
      parseJudgeResults({assertion_results: [{passed: "false"}]}),
    ).toBeUndefined();
  });
});

describe("numeric option validation", () => {
  it("accepts positive finite values and a safe iteration", () => {
    expect(
      parseOutputOptions({
        runs: "2",
        concurrency: "4",
        timeout: "600",
        budget: "0.25",
        iteration: "iteration-3",
      }),
    ).toEqual({
      success: true,
      data: {
        runs: 2,
        concurrency: 4,
        timeout: 600,
        budget: 0.25,
        iteration: "iteration-3",
      },
    });
  });

  it.each([
    {runs: "0", concurrency: "1", timeout: "1"},
    {runs: "1", concurrency: "NaN", timeout: "1"},
    {runs: "1", concurrency: "1", timeout: "-1"},
    {runs: "1", concurrency: "1", timeout: "1", budget: "Infinity"},
    {
      runs: "1",
      concurrency: "1",
      timeout: "1",
      iteration: "../escape",
    },
  ])("rejects invalid options: %o", (input) => {
    expect(parseOutputOptions(input).success).toBe(false);
  });
});
