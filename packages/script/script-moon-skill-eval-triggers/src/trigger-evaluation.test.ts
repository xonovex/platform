import {describe, expect, it} from "vitest";
import {
  runTriggerEvaluation,
  type TriggerCheck,
  type TriggerEvaluationOptions,
} from "./trigger-evaluation.js";

const optionsWith = (
  check: TriggerCheck,
  queryBatches: TriggerEvaluationOptions["queryBatches"],
): TriggerEvaluationOptions => ({
  queryBatches,
  runs: 2,
  threshold: 0.5,
  skillName: "test-skill",
  workspace: undefined,
  check,
});

const matchPositiveQuery: TriggerCheck = (query) =>
  Promise.resolve({triggered: query === "positive", error: null});

const failInfrastructure: TriggerCheck = () =>
  Promise.resolve({triggered: false, error: "unavailable"});

describe("trigger evaluation", () => {
  it("summarizes matching positive and negative trigger decisions", async () => {
    const options = optionsWith(matchPositiveQuery, [
      [
        {
          query: "positive",
          should_trigger: true,
          rationale: "matches",
          split: "train",
        },
        {
          query: "negative",
          should_trigger: false,
          rationale: "does not match",
          split: "validation",
        },
      ],
    ]);

    const result = await runTriggerEvaluation(options);

    expect(result).toMatchObject({
      success: true,
      passed: 2,
      failed: 0,
      total: 2,
    });
  });

  it("retries a transient failure instead of discarding the batch", async () => {
    let calls = 0;
    const flakyOnce: TriggerCheck = (query) => {
      calls += 1;
      // Fail the very first attempt, then behave normally.
      if (calls === 1)
        return Promise.resolve({triggered: false, error: "flaky"});
      return Promise.resolve({triggered: query === "positive", error: null});
    };
    const options = optionsWith(flakyOnce, [
      [
        {
          query: "positive",
          should_trigger: true,
          rationale: "matches",
          split: "train",
        },
      ],
    ]);

    const result = await runTriggerEvaluation(options);

    expect(result).toMatchObject({success: true, passed: 1, total: 1});
    expect(calls).toBe(3); // one wasted attempt, then the two real runs
  });

  it("invalidates evidence when every attempt fails", async () => {
    const options = optionsWith(failInfrastructure, [
      [
        {
          query: "positive",
          should_trigger: true,
          rationale: "matches",
          split: "train",
        },
      ],
    ]);

    const result = await runTriggerEvaluation(options);

    expect(result).toMatchObject({success: false, results: []});
  });
});
