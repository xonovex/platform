import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {
  buildEvaluationPrompt,
  resolveEvaluationConfig,
  type EvaluationConfig,
} from "./evaluation-config.js";
import type {NormalizedEval} from "./validation.js";

const validEvaluation = (id: string | number = 1) => ({
  id,
  prompt: "Answer the question",
  expected_output: "the answer",
  assertions: ["response contains the answer"],
  files: [],
});

const evalFile = (evals: readonly unknown[], skillName = "test-skill") => ({
  skill_name: skillName,
  tier: "moderate",
  evals,
});

type EvaluationConfigResult = ReturnType<typeof resolveEvaluationConfig>;

const successfulConfig = (result: EvaluationConfigResult): EvaluationConfig => {
  expect(result.success).toBe(true);
  if (!result.success) throw new Error(result.error);
  return result.data;
};

const failureMessage = (result: EvaluationConfigResult): string => {
  expect(result.success).toBe(false);
  if (result.success) throw new Error("expected configuration failure");
  return result.error;
};

const normalizedEvaluation = (files: readonly string[]): NormalizedEval => ({
  ...validEvaluation(),
  files,
});

describe("evaluation configuration", () => {
  let directory: string;

  beforeEach(() => {
    directory = mkdtempSync(join(tmpdir(), "skill-eval-config-"));
    writeFileSync(
      join(directory, "SKILL.md"),
      "---\nname: 'plugin:test-skill'\n---\n",
    );
    writeFileSync(
      join(directory, "evals.json"),
      JSON.stringify(evalFile([validEvaluation()])),
    );
  });

  afterEach(() => {
    rmSync(directory, {recursive: true, force: true});
  });

  const resolveConfig = (
    argv: readonly string[] = [],
    environment: Readonly<Record<string, string | undefined>> = {},
  ) =>
    resolveEvaluationConfig(argv, {
      environment,
      executableAvailable: () => true,
      workingDirectory: directory,
    });

  it("resolves defaults, environment options, batches, and the next iteration", () => {
    mkdirSync(join(directory, "workspace", "iteration-2"), {recursive: true});
    mkdirSync(join(directory, "workspace", "iteration-invalid"), {
      recursive: true,
    });

    const config = successfulConfig(
      resolveConfig([], {
        WORKSPACE: "workspace",
        RUNS: "2",
        CONCURRENCY: "1",
        GEN_TIMEOUT: "30",
        CLAUDE_MODEL: "generation-model",
        JUDGE_MODEL: "judge-model",
        MAX_BUDGET_USD: "0.05",
        JUDGE_MAX_BUDGET_USD: "0.04",
        EVAL_CWD: "",
      }),
    );

    expect(config).toMatchObject({
      skillName: "plugin:test-skill",
      shortName: "test-skill",
      iteration: "iteration-3",
      runs: 2,
      concurrency: 1,
      timeout: 30,
      budget: 0.05,
      judgeBudget: 0.04,
      harness: "claude",
      model: "generation-model",
      judgeModel: "judge-model",
      cwd: undefined,
      maxBatchModelCalls: 8,
    });
    expect(config.evaluationBatches).toHaveLength(1);
    expect(config.withArgs).toContain("generation-model");
    expect(config.withoutArgs).toContain(
      "--disallowedTools=Skill,Bash,Edit,Write,NotebookEdit,WebFetch",
    );
    expect(config.benchmarkPath).toBe(
      join(directory, "workspace", "iteration-3", "benchmark.json"),
    );
  });

  it("honors CLI options and explicit positional values", () => {
    writeFileSync(
      join(directory, "evals.json"),
      JSON.stringify(evalFile([validEvaluation()], "explicit-skill")),
    );
    const config = successfulConfig(
      resolveConfig([
        "--runs",
        "1",
        "--batch-size",
        "1",
        "--eval-cwd",
        "run-here",
        "--model",
        "cli-model",
        "--max-turns",
        "18",
        "evals.json",
        "explicit-skill",
        "iteration-4",
      ]),
    );

    expect(config).toMatchObject({
      skillName: "explicit-skill",
      tier: "moderate",
      iteration: "iteration-4",
      cwd: "run-here",
      model: "cli-model",
      maxTurns: 18,
    });
    expect(config.withArgs[config.withArgs.indexOf("--max-turns") + 1]).toBe(
      "18",
    );
    expect(
      config.withoutArgs[config.withoutArgs.indexOf("--max-turns") + 1],
    ).toBe("18");
  });

  it("defaults the turn cap and rejects one above the ceiling", () => {
    expect(successfulConfig(resolveConfig())).toMatchObject({maxTurns: 12});
    expect(
      successfulConfig(resolveConfig([], {MAX_TURNS: "20"})),
    ).toMatchObject({maxTurns: 20});
    expect(failureMessage(resolveConfig(["--max-turns", "25"]))).toContain(
      "invalid evaluator options",
    );
  });

  it("uses pinned Claude generation and judge models by default", () => {
    const config = successfulConfig(resolveConfig());

    expect(config).toMatchObject({
      model: "claude-haiku-4-5-20251001",
      judgeModel: "claude-sonnet-4-6",
    });
  });

  it("selects Codex for generation and judging", () => {
    const config = successfulConfig(resolveConfig(["--harness", "codex"]));

    expect(config).toMatchObject({
      harness: "codex",
      model: "gpt-5.3-codex",
      judgeModel: "gpt-5.3-codex",
    });
    expect(config.withArgs).toEqual(
      expect.arrayContaining([
        "exec",
        "--json",
        "--ephemeral",
        "gpt-5.3-codex",
      ]),
    );
  });

  it("keeps valid evaluations and reports skipped invalid entries", () => {
    writeFileSync(
      join(directory, "evals.json"),
      JSON.stringify(evalFile([{prompt: ""}, validEvaluation("valid")])),
    );

    const result = resolveConfig();

    expect(result.warnings).toHaveLength(1);
    expect(result.warnings[0]).toContain("Skipping eval #1");
    const config = successfulConfig(result);
    expect(config.evaluations.map(({id}) => id)).toEqual(["valid"]);
  });

  it("rejects missing inputs and invalid evaluator state", () => {
    expect(resolveConfig(["missing.json"])).toMatchObject({
      success: false,
      error: "evals file not found: missing.json",
    });
    expect(
      resolveEvaluationConfig([], {
        environment: {},
        executableAvailable: () => false,
        workingDirectory: directory,
      }),
    ).toMatchObject({
      success: false,
      error: "'claude' CLI not found in PATH",
    });

    rmSync(join(directory, "SKILL.md"));
    expect(failureMessage(resolveConfig())).toContain("no skill_name given");
  });

  it("rejects malformed, empty, ungradable, and duplicate evaluations", () => {
    const evalsPath = join(directory, "evals.json");
    writeFileSync(evalsPath, "not-json");
    expect(failureMessage(resolveConfig())).toContain("invalid JSON");

    writeFileSync(evalsPath, JSON.stringify({unknown: []}));
    expect(failureMessage(resolveConfig())).toContain("invalid eval structure");

    writeFileSync(evalsPath, JSON.stringify(evalFile([])));
    expect(failureMessage(resolveConfig())).toContain("has no evals");

    writeFileSync(evalsPath, JSON.stringify(evalFile([{prompt: ""}])));
    expect(resolveConfig()).toMatchObject({
      success: false,
      error: "no gradable evals",
    });

    writeFileSync(
      evalsPath,
      JSON.stringify(evalFile([validEvaluation(1), validEvaluation(1)])),
    );
    expect(failureMessage(resolveConfig())).toContain("duplicate eval id");
  });

  it("rejects invalid directories, options, and oversized batches", () => {
    expect(resolveConfig(["--plugin-dir", "missing-plugin"])).toMatchObject({
      success: false,
      error: "local plugin directory not found: missing-plugin",
    });
    expect(failureMessage(resolveConfig(["--runs", "4"]))).toContain(
      "invalid evaluator options",
    );

    writeFileSync(
      join(directory, "evals.json"),
      JSON.stringify(
        evalFile(Array.from({length: 7}, (_, index) => validEvaluation(index))),
      ),
    );
    expect(failureMessage(resolveConfig())).toContain(
      "would launch 28 model calls",
    );
    expect(resolveConfig(["--batch-size", "6"])).toMatchObject({
      success: true,
    });
  });
});

describe("buildEvaluationPrompt", () => {
  it("returns the original prompt when no files are present", () => {
    expect(buildEvaluationPrompt("/evals", normalizedEvaluation([]))).toBe(
      "Answer the question",
    );
  });

  it("appends resolved input file paths", () => {
    expect(
      buildEvaluationPrompt("/evals", normalizedEvaluation(["a.md", "b.txt"])),
    ).toBe(`Answer the question

Relevant input files (read them as needed):
- /evals/a.md
- /evals/b.txt`);
  });
});
