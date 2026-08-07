import {describe, expect, it, vi} from "vitest";
import {main, writeRetryNotice} from "../../../src/evaluate.js";
import {
  aggregateArm,
  claudeFailureDetail,
  extractJson,
  meanBlock,
  summarize,
  sumTokens,
} from "../../../src/output-results.js";

describe("output evaluation helpers", () => {
  it("sums supported token fields and ignores invalid values", () => {
    expect(
      sumTokens({
        input_tokens: 10.9,
        output_tokens: "4",
        cache_creation_input_tokens: "invalid",
        cache_read_input_tokens: 2,
      }),
    ).toBe(16);
    expect(sumTokens(null)).toBe(0);
  });

  it("extracts JSON objects from fenced and surrounding text", () => {
    expect(extractJson('```json\n{"passed":true}\n```')).toEqual({
      passed: true,
    });
    expect(extractJson('result: {"passed":false}')).toEqual({passed: false});
    expect(extractJson("no object")).toBeNull();
    expect(extractJson("{invalid}")).toBeNull();
    expect(extractJson("[1, 2]")).toBeNull();
  });

  it("finds the latest structured Claude error detail", () => {
    expect(
      claudeFailureDetail(
        [
          "not-json",
          '{"subtype":"error_max_turns"}',
          '{"result":"quota exceeded","is_error":true}',
        ].join("\n"),
      ),
    ).toBe("quota exceeded");
    expect(claudeFailureDetail('{"subtype":"error_max_turns"}')).toBe(
      "error_max_turns",
    );
    expect(claudeFailureDetail('{"type":"result"}')).toBe("");
  });

  it("summarizes assertion results including an empty result set", () => {
    expect(
      summarize([
        {text: "first", passed: true, evidence: "yes"},
        {text: "second", passed: false, evidence: "no"},
      ]).summary,
    ).toEqual({passed: 1, failed: 1, total: 2, pass_rate: 0.5});
    expect(summarize([], "empty")).toMatchObject({
      summary: {passed: 0, failed: 0, total: 0, pass_rate: 0},
      error: "empty",
    });
  });

  it("computes means and optional population deviation", () => {
    expect(meanBlock([], 1)).toEqual({mean: 0});
    expect(meanBlock([1, 3], 1)).toEqual({mean: 2});
    expect(meanBlock([1, 3], 2)).toEqual({mean: 2, stddev: 1});
  });

  it("aggregates repeated runs by evaluation before averaging the arm", () => {
    const records = [
      {
        id: 1,
        arm: "with_skill" as const,
        pass_rate: 1,
        tokens: 100,
        duration_ms: 10,
        skill_triggered: true,
        error: null,
      },
      {
        id: 1,
        arm: "with_skill" as const,
        pass_rate: 0,
        tokens: 200,
        duration_ms: 20,
        skill_triggered: false,
        error: null,
      },
    ];

    const result = aggregateArm(records, "with_skill", 2);

    expect(result).toEqual({
      pass_rate: {mean: 0.5},
      tokens: {mean: 150},
      duration_ms: {mean: 15},
      skill_trigger_rate: {mean: 0.5},
    });
  });
});

describe("main", () => {
  it("fails cleanly when the evaluation file is missing", async () => {
    await expect(main(["missing-evals.json"])).resolves.toBe(2);
  });
});

describe("writeRetryNotice", () => {
  it("names the attempt, the attempt cap, and the failure", () => {
    const write = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);

    writeRetryNotice(2, "claude exited 1: error_max_turns");

    expect(write).toHaveBeenCalledWith(
      "retrying after transient failure (2/3): claude exited 1: error_max_turns\n",
    );
    write.mockRestore();
  });
});
