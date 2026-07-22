import {describe, expect, it} from "vitest";
import {
  buildGenerationClaudeArgs,
  buildGenerationPrompt,
  buildJudgeClaudeArgs,
  evalEntries,
  evaluateOutputGate,
  findEvaluationInfrastructureFailures,
  normalizeEval,
  outputModelCallCount,
  parseJudgeResults,
  parseOutputOptions,
  runFailFastPool,
  streamTextDeltaLength,
  validateUniqueEvaluationIds,
} from "./validation.js";

describe("generation prompt isolation", () => {
  it("requires the target skill only in the with-skill arm", () => {
    const prompt = "Assess this release candidate.";
    const skill = "xonovex-skill-accessibility:accessibility-guide";

    expect(buildGenerationPrompt(prompt, "with_skill", skill)).toBe(
      "/xonovex-skill-accessibility:accessibility-guide Assess this release candidate.",
    );
    expect(buildGenerationPrompt(prompt, "without_skill", skill)).toBe(prompt);
  });
});

describe("Claude process isolation", () => {
  it("allows only Skill and Read in the activated generation arm", () => {
    const args = buildGenerationClaudeArgs({
      arm: "with_skill",
      model: "haiku",
      budget: 0.1,
      disallowedTools: "Bash,Edit,Write",
      pluginDirectories: ["dependency-directory", "plugin-directory"],
    });

    expect(args).toContain("Skill,Read");
    expect(args).toContain("0.1");
    expect(args).toContain("6");
    expect(args).toContain("--include-partial-messages");
    expect(args.some((arg) => arg.includes("under 1,000 words"))).toBe(true);
    expect(args).toContain("");
    expect(args).not.toContain("Bash");
    expect(args.filter((arg) => arg === "--plugin-dir")).toHaveLength(2);
  });

  it("allows only Read in baseline and no tools in the judge", () => {
    const baseline = buildGenerationClaudeArgs({
      arm: "without_skill",
      model: "haiku",
      budget: 0.1,
      disallowedTools: "Bash,Edit,Write",
    });
    const judge = buildJudgeClaudeArgs({
      model: "sonnet",
      budget: 0.1,
      assertionCount: 3,
    });

    expect(
      baseline.slice(
        baseline.indexOf("--tools"),
        baseline.indexOf("--tools") + 2,
      ),
    ).toEqual(["--tools", "Read"]);
    expect(baseline).toContain("--disallowedTools=Skill,Bash,Edit,Write");
    expect(
      judge.slice(judge.indexOf("--tools"), judge.indexOf("--tools") + 2),
    ).toEqual(["--tools", ""]);
    expect(judge).toContain("0.1");
    expect(judge).toContain("1");
    expect(judge).toContain("--json-schema");
  });
});

describe("stream output accounting", () => {
  it("counts only assistant text deltas", () => {
    expect(
      streamTextDeltaLength({
        type: "stream_event",
        event: {
          type: "content_block_delta",
          delta: {type: "text_delta", text: "answer"},
        },
      }),
    ).toBe(6);
    expect(
      streamTextDeltaLength({
        type: "stream_event",
        event: {
          type: "content_block_delta",
          delta: {type: "input_json_delta", partial_json: "tool input"},
        },
      }),
    ).toBe(0);
  });
});

describe("fail-fast scheduling", () => {
  it("does not start another job after a failure", async () => {
    const started: number[] = [];
    const results = await runFailFastPool(
      [1, 2, 3, 4],
      1,
      (item) => {
        started.push(item);
        return Promise.resolve(item);
      },
      (item) => item === 2,
    );

    expect(results).toEqual([1, 2]);
    expect(started).toEqual([1, 2]);
  });
});

describe("benchmark evidence health", () => {
  it("rejects process errors and missing target activation", () => {
    expect(
      findEvaluationInfrastructureFailures([
        {
          id: 1,
          arm: "with_skill",
          skill_triggered: true,
          error: null,
        },
        {
          id: 1,
          arm: "without_skill",
          skill_triggered: false,
          error: null,
        },
      ]),
    ).toEqual([]);

    expect(
      findEvaluationInfrastructureFailures([
        {
          id: 2,
          arm: "with_skill",
          skill_triggered: false,
          error: null,
        },
        {
          id: 3,
          arm: "without_skill",
          skill_triggered: false,
          error: "timeout",
        },
      ]),
    ).toEqual([
      {id: 2, arm: "with_skill", reason: "target skill did not activate"},
      {id: 3, arm: "without_skill", reason: "timeout"},
    ]);
  });
});

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

  it("requires the skill name and optimization tier in the eval envelope", () => {
    expect(
      evalEntries({
        skill_name: "test-guide",
        tier: "moderate",
        evals: [{}],
      }),
    ).toEqual({
      success: true,
      data: {skillName: "test-guide", tier: "moderate", evals: [{}]},
    });
    expect(evalEntries([{}]).success).toBe(false);
    expect(evalEntries({evals: [{}]}).success).toBe(false);
    expect(evalEntries({evals: "wrong"}).success).toBe(false);
  });

  it("applies tiered absolute, delta, and activation quality gates", () => {
    expect(evaluateOutputGate("moderate", 0.9, 0.8, 1)).toMatchObject({
      passed: true,
      policy: {minimumWithSkillPassRate: 0.8, minimumDeltaPassRate: 0.05},
    });
    expect(evaluateOutputGate("moderate", 0.333, 0, 1)).toMatchObject({
      passed: false,
      checks: {withSkillPassRate: false},
    });
    expect(evaluateOutputGate("conservative", 0.9, 0.85, 1)).toMatchObject({
      passed: false,
      checks: {deltaPassRate: false},
    });
    expect(evaluateOutputGate("aggressive", 1, 0, 0.5)).toMatchObject({
      passed: false,
      checks: {skillTriggerRate: false},
    });
  });

  it("rejects evaluation IDs that resolve to the same output directory", () => {
    const evaluations = [
      {
        id: 1,
        prompt: "First",
        expected_output: "",
        assertions: ["First assertion"],
        files: [],
      },
      {
        id: "1",
        prompt: "Second",
        expected_output: "",
        assertions: ["Second assertion"],
        files: [],
      },
    ];

    expect(validateUniqueEvaluationIds(evaluations)).toEqual({
      success: false,
      error: "duplicate eval id: 1",
    });
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
        concurrency: "2",
        timeout: "600",
        budget: "0.1",
        judgeBudget: "0.1",
        batchSize: "6",
        iteration: "iteration-3",
      }),
    ).toEqual({
      success: true,
      data: {
        runs: 2,
        concurrency: 2,
        timeout: 600,
        budget: 0.1,
        judgeBudget: 0.1,
        batchSize: 6,
        iteration: "iteration-3",
      },
    });
  });

  it.each([
    {runs: "0", concurrency: "1", timeout: "1"},
    {runs: "4", concurrency: "1", timeout: "1"},
    {runs: "1", concurrency: "3", timeout: "1"},
    {runs: "1", concurrency: "NaN", timeout: "1"},
    {runs: "1", concurrency: "1", timeout: "-1"},
    {runs: "1", concurrency: "1", timeout: "1", budget: "Infinity"},
    {runs: "1", concurrency: "1", timeout: "1", budget: "0.11"},
    {runs: "1", concurrency: "1", timeout: "1", judgeBudget: "0.11"},
    {runs: "1", concurrency: "1", timeout: "1", batchSize: "0"},
    {
      runs: "1",
      concurrency: "1",
      timeout: "1",
      iteration: "../escape",
    },
  ])("rejects invalid options: %o", (input) => {
    expect(parseOutputOptions(input).success).toBe(false);
  });

  it("counts both generation arms and their judges", () => {
    expect(outputModelCallCount(6, 1)).toBe(24);
    expect(outputModelCallCount(7, 1)).toBe(28);
  });
});
