import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {beforeEach, describe, expect, it} from "vitest";
import {
  buildEvaluationPrompt,
  resolveEvaluationConfig,
  type EvaluationConfig,
} from "../../../src/evaluation-config.js";
import type {NormalizedEval} from "../../../src/validation.js";

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

const DIRECTORY = "/guide";

// The guide a run resolves from: a SKILL.md naming the skill, plus the file the
// evaluations are read out of. A case that needs one of them absent builds the
// tree without it rather than deleting it.
const guideTree = ({skill = true}: {skill?: boolean} = {}): FileSystem =>
  memoryFileSystem({
    files: {
      ...(skill
        ? {
            [join(DIRECTORY, "SKILL.md")]:
              "---\nname: 'plugin:test-skill'\n---\n",
          }
        : {}),
      [join(DIRECTORY, "evals.json")]: JSON.stringify(
        evalFile([validEvaluation()]),
      ),
    },
  });

describe("evaluation configuration", () => {
  let directory: string;
  let fs: FileSystem;

  beforeEach(() => {
    directory = DIRECTORY;
    fs = guideTree();
  });

  const resolveConfig = (
    argv: readonly string[] = [],
    environment: Readonly<Record<string, string | undefined>> = {},
  ) =>
    resolveEvaluationConfig(argv, {
      environment,
      executableAvailable: () => true,
      workingDirectory: directory,
      fs,
    });

  it("resolves defaults, environment options, batches, and the next iteration", () => {
    fs.makeDirectory(join(directory, "workspace", "iteration-2"));
    fs.makeDirectory(join(directory, "workspace", "iteration-invalid"));

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
    fs.writeFile(
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
      judgeModel: "claude-sonnet-5",
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
    fs.writeFile(
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
        fs,
      }),
    ).toMatchObject({
      success: false,
      error: "'claude' CLI not found in PATH",
    });

    fs = guideTree({skill: false});
    expect(failureMessage(resolveConfig())).toContain("no skill_name given");
  });

  it("rejects malformed, empty, ungradable, and duplicate evaluations", () => {
    const evalsPath = join(directory, "evals.json");
    fs.writeFile(evalsPath, "not-json");
    expect(failureMessage(resolveConfig())).toContain("invalid JSON");

    fs.writeFile(evalsPath, JSON.stringify({unknown: []}));
    expect(failureMessage(resolveConfig())).toContain("invalid eval structure");

    fs.writeFile(evalsPath, JSON.stringify(evalFile([])));
    expect(failureMessage(resolveConfig())).toContain("has no evals");

    fs.writeFile(evalsPath, JSON.stringify(evalFile([{prompt: ""}])));
    expect(resolveConfig()).toMatchObject({
      success: false,
      error: "no gradable evals",
    });

    fs.writeFile(
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

    fs.writeFile(
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
