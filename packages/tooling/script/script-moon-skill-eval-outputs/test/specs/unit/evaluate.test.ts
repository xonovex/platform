import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryProcessRunner} from "@xonovex/script-moon-skill-eval-common/adapters/process-runner-memory";
import {memoryScratch} from "@xonovex/script-moon-skill-eval-common/adapters/scratch-memory";
import type {ProcessInvocation} from "@xonovex/script-moon-skill-eval-common/process-runner";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {main, type EvaluatorDependencies} from "../../../src/evaluate.js";
import type {NormalizedEval} from "../../../src/validation.js";

const evaluation = (prompt: string): NormalizedEval => ({
  id: 1,
  prompt,
  expected_output: "the answer",
  assertions: ["response contains the answer"],
  files: [],
});

const finalArgument = (request: ProcessInvocation): string =>
  request.args.at(-1) ?? "";

const isJudgeRequest = (request: ProcessInvocation): boolean =>
  finalArgument(request).includes("ASSISTANT RESPONSE:");

// The with_skill arm prefixes the prompt with the skill's slash command, which is
// how a recorded harness tells the two arms apart and answers only the one that
// was given the skill.
const usedSkill = (request: ProcessInvocation): boolean =>
  finalArgument(request).startsWith("/test-skill ");

const generationOutput = (result: string): string =>
  [
    JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}),
    JSON.stringify({
      message: {
        content: [
          {type: "tool_use", name: "Skill", input: {skill: "test-skill"}},
        ],
      },
    }),
    JSON.stringify({
      type: "result",
      result,
      usage: {input_tokens: 3, output_tokens: 2},
      duration_ms: 12,
    }),
  ].join("\n");

// Grades an assertion as passed when the response the judge was handed is the
// answer, which is what separates the two arms in the aggregate.
const judgementOutput = (request: ProcessInvocation, count: number): string => {
  const passed = finalArgument(request).includes(
    "ASSISTANT RESPONSE:\nthe answer",
  );
  return JSON.stringify({
    structured_output: {
      assertion_results: Array.from({length: count}, () => ({
        passed,
        evidence: passed ? "answer present" : "answer missing",
      })),
    },
  });
};

// The arm that was given the skill answers correctly; the arm without it does not,
// which is the delta every case below reads.
const armGeneration = (request: ProcessInvocation): string =>
  generationOutput(usedSkill(request) ? "the answer" : "wrong");

// A tree the evaluator reads its inputs from and writes its evidence into, so a
// whole sweep can be scored and asserted on without a disk.
const dependencies = (
  fs: FileSystem,
  assertionCount = 1,
  failGeneration = false,
): EvaluatorDependencies => ({
  fs,
  scratch: memoryScratch(fs),
  environment: {},
  workingDirectory: ROOT,
  now: () => 0,
  // The harness is answered from this tree, so its location must not be read
  // from the host's PATH.
  locateExecutable: (command) => `/usr/bin/${command}`,
  runProcess: memoryProcessRunner((request) => {
    if (failGeneration && !isJudgeRequest(request)) {
      return {code: 3, stderr: "generation failed"};
    }
    return {
      stdout: isJudgeRequest(request)
        ? judgementOutput(request, assertionCount)
        : armGeneration(request),
    };
  }).run,
});

const ROOT = "/workspace";

const readWorkspaceJson = (
  fs: FileSystem,
  workspace: string,
  file: string,
): unknown => JSON.parse(fs.readText(join(workspace, "iteration-1", file)));

const runEvaluator = async (
  name: string,
  evaluations: readonly NormalizedEval[],
  makeDependencies: (fs: FileSystem) => EvaluatorDependencies,
): Promise<{exitCode: number; workspace: string; fs: FileSystem}> => {
  const evaluationsFile = join(ROOT, `${name}.json`);
  const workspace = join(ROOT, name);
  // resolveEvaluationConfig locates the skill by its SKILL.md, so the tree has
  // to carry one alongside the evaluations file.
  const fs = memoryFileSystem({
    files: {
      [evaluationsFile]: JSON.stringify({
        skill_name: "test-skill",
        tier: "moderate",
        evals: evaluations,
      }),
      [join(ROOT, "SKILL.md")]:
        "---\nname: test-skill\ndescription: Use for tests.\n---\n",
    },
  });

  return {
    fs,
    workspace,
    exitCode: await main(
      ["--workspace", workspace, evaluationsFile, "test-skill", "iteration-1"],
      makeDependencies(fs),
    ),
  };
};

describe("the output evaluator", () => {
  beforeEach(() => {
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    vi.spyOn(process.stdout, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("runs the complete evaluator and writes benchmark evidence", async () => {
    const {exitCode, fs, workspace} = await runEvaluator(
      "evals",
      [evaluation("question")],
      (fs) => dependencies(fs),
    );

    expect(exitCode).toBe(0);
    expect(readWorkspaceJson(fs, workspace, "benchmark.json")).toMatchObject({
      skill: "test-skill",
      tier: "moderate",
      iteration: "iteration-1",
      run_summary: {
        with_skill: {pass_rate: {mean: 1}},
        without_skill: {pass_rate: {mean: 0}},
        delta: {pass_rate: 1},
      },
      quality_gate: {passed: true},
    });
  });

  it("reports each skipped evaluation on stderr and evaluates the rest", async () => {
    const {exitCode} = await runEvaluator(
      "skipped",
      [{...evaluation("blank"), prompt: ""}, evaluation("question")],
      (fs) => dependencies(fs),
    );

    expect(exitCode).toBe(0);
    expect(process.stderr.write).toHaveBeenCalledWith(
      expect.stringContaining("Skipping eval #1"),
    );
  });

  it("fails a valid positive delta below the tier's absolute quality floor", async () => {
    const {exitCode, fs, workspace} = await runEvaluator(
      "low-quality",
      [
        {
          ...evaluation("question"),
          assertions: ["one", "two", "three", "four"],
        },
      ],
      // One of four assertions passes, which clears the delta check and fails the
      // tier's absolute floor.
      (fs) => ({
        ...dependencies(fs),
        runProcess: memoryProcessRunner((request) => ({
          stdout: isJudgeRequest(request)
            ? JSON.stringify({
                structured_output: {
                  assertion_results: [
                    {
                      passed: finalArgument(request).includes(
                        "ASSISTANT RESPONSE:\nthe answer",
                      ),
                      evidence: "first only",
                    },
                    {passed: false, evidence: "no"},
                    {passed: false, evidence: "no"},
                    {passed: false, evidence: "no"},
                  ],
                },
              })
            : armGeneration(request),
        })).run,
      }),
    );

    expect(exitCode).toBe(1);
    expect(readWorkspaceJson(fs, workspace, "benchmark.json")).toMatchObject({
      run_summary: {
        with_skill: {pass_rate: {mean: 0.25}},
        without_skill: {pass_rate: {mean: 0}},
      },
      quality_gate: {
        passed: false,
        checks: {withSkillPassRate: false, deltaPassRate: true},
      },
    });
  });

  it("invalidates the whole run when a generation fails as infrastructure", async () => {
    const {exitCode, fs, workspace} = await runEvaluator(
      "failing",
      [evaluation("question")],
      (fs) => dependencies(fs, 1, true),
    );

    expect(exitCode).toBe(2);
    expect(readWorkspaceJson(fs, workspace, "invalid-run.json")).toMatchObject({
      skill: "test-skill",
      iteration: "iteration-1",
      status: "invalid",
    });
  });
});
