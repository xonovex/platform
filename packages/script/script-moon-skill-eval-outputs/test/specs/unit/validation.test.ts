import {describe, expect, it} from "vitest";
import {
  buildGenerationClaudeArgs,
  buildGenerationPrompt,
  buildJudgeClaudeArgs,
  evalEntries,
  evaluateOutputGate,
  findEvaluationInfrastructureFailures,
  GENERATION_MAX_TURNS,
  normalizeEval,
  outputModelCallCount,
  parseJudgeResults,
  parseOutputOptions,
  runFailFastPool,
  runWithTransientRetry,
  validateUniqueEvaluationIds,
} from "../../../src/validation.js";

// Both generation arms and the judge run through buildIsolatedClaudeArgs. These
// assertions fail if a builder stops routing through it, which is the only way the
// harness could inherit the running machine's settings or MCP servers.
const expectIsolated = (args: readonly string[]): void => {
  const settingIndex = args.indexOf("--setting-sources");
  expect(args.slice(settingIndex, settingIndex + 2)).toEqual([
    "--setting-sources",
    "",
  ]);
  const mcpIndex = args.indexOf("--mcp-config");
  expect(args.slice(mcpIndex, mcpIndex + 2)).toEqual([
    "--mcp-config",
    '{"mcpServers":{}}',
  ]);
  expect(args).toContain("--strict-mcp-config");
  expect(args).toContain("--no-session-persistence");
  expect(args).toContain("--no-chrome");
};

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
      maxTurns: GENERATION_MAX_TURNS,
      disallowedTools: "Bash,Edit,Write",
      pluginDirectories: ["dependency-directory", "plugin-directory"],
    });

    expect(args).toContain("Skill,Read");
    expect(args).toContain("0.1");
    // Enough turns for the arm to load a progressively-disclosed skill and still
    // answer; too few reports error_max_turns and invalidates the batch.
    expect(args[args.indexOf("--max-turns") + 1]).toBe("12");
    expect(args).toContain("--include-partial-messages");
    expect(args.some((arg) => arg.includes("under 1,000 words"))).toBe(true);
    expectIsolated(args);
    expect(args).not.toContain("Bash");
    expect(args.filter((arg) => arg === "--plugin-dir")).toHaveLength(2);
  });

  it("honours a raised per-generation turn cap", () => {
    const args = buildGenerationClaudeArgs({
      arm: "with_skill",
      model: "haiku",
      budget: 0.1,
      maxTurns: 24,
      disallowedTools: "",
    });

    expect(args[args.indexOf("--max-turns") + 1]).toBe("24");
  });

  it("allows only Read in baseline and no tools in the judge", () => {
    const baseline = buildGenerationClaudeArgs({
      arm: "without_skill",
      model: "haiku",
      budget: 0.1,
      maxTurns: GENERATION_MAX_TURNS,
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
    expectIsolated(baseline);
    expectIsolated(judge);
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

describe("transient failure retry", () => {
  it("returns the first healthy result without retrying", async () => {
    const attempts: number[] = [];
    const result = await runWithTransientRetry(
      () => Promise.resolve({error: null, value: "ok"}),
      3,
      (attempt) => {
        attempts.push(attempt);
      },
    );

    expect(result).toEqual({error: null, value: "ok"});
    expect(attempts).toEqual([]);
  });

  it("retries a transient failure and keeps the recovered result", async () => {
    const outcomes = [
      {error: "claude exited 1: error_max_turns"},
      {error: null},
    ];
    const retried: string[] = [];
    const result = await runWithTransientRetry(
      () => Promise.resolve(outcomes.shift() ?? {error: "exhausted"}),
      3,
      (_attempt, error) => {
        retried.push(error);
      },
    );

    expect(result).toEqual({error: null});
    expect(retried).toEqual(["claude exited 1: error_max_turns"]);
  });

  it("surfaces the last failure once every attempt is spent", async () => {
    let runs = 0;
    const result = await runWithTransientRetry(
      () => {
        runs += 1;
        return Promise.resolve({error: `timeout ${String(runs)}`});
      },
      3,
      () => {},
    );

    expect(runs).toBe(3);
    expect(result).toEqual({error: "timeout 3"});
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
        maxTurns: "24",
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
        maxTurns: 24,
        batchSize: 6,
        iteration: "iteration-3",
      },
    });
  });

  it("defaults the turn cap when none is supplied", () => {
    const result = parseOutputOptions({
      runs: "1",
      concurrency: "1",
      timeout: "1",
    });

    expect(result).toMatchObject({
      data: {maxTurns: GENERATION_MAX_TURNS},
      success: true,
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
    {runs: "1", concurrency: "1", timeout: "1", maxTurns: "0"},
    {runs: "1", concurrency: "1", timeout: "1", maxTurns: "25"},
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
