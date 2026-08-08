import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import type {
  HarnessRequest,
  HarnessRunner,
  ProcessOutput,
} from "../../../src/output-harness.js";
import {CODEX_SKILL_SIGNAL} from "../../../src/output-parse.js";
import {runJob, type RunContext} from "../../../src/output-process.js";
import type {NormalizedEval} from "../../../src/validation.js";

const evaluation = (prompt: string): NormalizedEval => ({
  id: 1,
  prompt,
  expected_output: "the answer",
  assertions: ["response contains the answer"],
  files: [],
});

const output = (overrides: Partial<ProcessOutput> = {}): ProcessOutput => ({
  stdout: "",
  timedOut: false,
  outputLimitExceeded: false,
  error: null,
  ...overrides,
});

const claudeGeneration = (result: string): string =>
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

const claudeJudgement = (passed: boolean): string =>
  JSON.stringify({
    structured_output: {
      assertion_results: [
        {passed, evidence: passed ? "answer present" : "answer missing"},
      ],
    },
  });

// A runner that answers from recorded output and records what it was asked to run,
// which is what lets these cases score a job without a harness process.
const recordingRunner = (
  reply: (request: HarnessRequest) => ProcessOutput,
): {runner: HarnessRunner; requests: HarnessRequest[]} => {
  const requests: HarnessRequest[] = [];
  return {
    requests,
    runner: (request) => {
      requests.push(request);
      return Promise.resolve(reply(request));
    },
  };
};

const isJudgeRequest = (request: HarnessRequest): boolean =>
  request.finalArgument.includes("ASSISTANT RESPONSE:");

const ITERATION_DIRECTORY = "/workspace/iteration-1";

describe("runJob", () => {
  let tree: FileSystem;

  beforeEach(() => {
    tree = memoryFileSystem();
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const context = (overrides: Partial<RunContext> = {}): RunContext => ({
    withArgs: ["--with"],
    withoutArgs: ["--without"],
    cwd: undefined,
    guideDirectory: join(ITERATION_DIRECTORY, "guide"),
    harness: "claude",
    timeout: 2,
    target: "test-skill",
    shortName: "test-skill",
    judgeModel: "judge-model",
    judgeBudget: 0.01,
    iterationDirectory: ITERATION_DIRECTORY,
    fs: tree,
    runs: 1,
    buildPrompt: (value) => value.prompt,
    runHarness: () => Promise.resolve(output()),
    now: () => 0,
    ...overrides,
  });

  const readArm = (arm: string, file: string, run?: string): string =>
    tree.readText(
      join(
        ITERATION_DIRECTORY,
        "eval-1",
        arm,
        ...(run === undefined ? [] : [run]),
        file,
      ),
    );

  it("generates, grades, and persists a healthy run", async () => {
    const {runner} = recordingRunner((request) =>
      output({
        stdout: isJudgeRequest(request)
          ? claudeJudgement(true)
          : claudeGeneration("the answer"),
      }),
    );

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({runHarness: runner}),
    );

    expect(record).toEqual({
      id: 1,
      arm: "with_skill",
      pass_rate: 1,
      tokens: 5,
      duration_ms: 12,
      skill_triggered: true,
      error: null,
    });
    expect(readArm("with_skill", join("outputs", "response.md"))).toBe(
      "the answer",
    );
    expect(JSON.parse(readArm("with_skill", "timing.json"))).toMatchObject({
      total_tokens: 5,
      skill_triggered: true,
    });
  });

  it("sends the arm's own harness arguments", async () => {
    const {runner, requests} = recordingRunner((request) =>
      output({
        stdout: isJudgeRequest(request)
          ? claudeJudgement(true)
          : claudeGeneration("the answer"),
      }),
    );

    await runJob(
      evaluation("answer the question"),
      "without_skill",
      0,
      context({runHarness: runner}),
    );

    expect(requests[0]?.args).toEqual(["--without"]);
    expect(requests[0]?.skillGuide).toBeUndefined();
  });

  it("stages the guide for a Codex run that is expected to use the skill", async () => {
    const guideDirectory = join(ITERATION_DIRECTORY, "guide");
    const {runner, requests} = recordingRunner((request) =>
      output({
        stdout: isJudgeRequest(request)
          ? JSON.stringify({
              type: "item.completed",
              item: {
                type: "agent_message",
                text: '{"assertion_results":[{"passed":true,"evidence":"present"}]}',
              },
            })
          : [
              JSON.stringify({
                type: "item.completed",
                item: {
                  type: "agent_message",
                  text: `${CODEX_SKILL_SIGNAL}\nthe answer`,
                },
              }),
              JSON.stringify({
                type: "turn.completed",
                usage: {input_tokens: 4, output_tokens: 2},
              }),
            ].join("\n"),
      }),
    );

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({harness: "codex", guideDirectory, runHarness: runner}),
    );

    expect(requests[0]?.skillGuide).toBe(guideDirectory);
    expect(requests[0]?.maxOutputCharacters).toBe(10_000);
    expect(record).toMatchObject({
      pass_rate: 1,
      tokens: 6,
      skill_triggered: true,
      error: null,
    });
  });

  it("records a generation process failure without grading it", async () => {
    const {runner, requests} = recordingRunner(() =>
      output({error: "claude exited 3: generation failed"}),
    );

    const record = await runJob(
      evaluation("answer the question"),
      "without_skill",
      0,
      context({runHarness: runner}),
    );

    expect(record).toMatchObject({
      pass_rate: 0,
      skill_triggered: false,
      error: "claude exited 3: generation failed",
    });
    expect(requests).toHaveLength(1);
    expect(readArm("without_skill", "grading.json")).toContain(
      "not graded because generation evidence is invalid",
    );
  });

  it("reports a generation that hit the output ceiling or the clock", async () => {
    for (const [reply, error, durationMs] of [
      [output({outputLimitExceeded: true}), "output-limit", 0],
      [output({timedOut: true}), "timeout", 2000],
    ] as const) {
      const record = await runJob(
        evaluation("answer the question"),
        "without_skill",
        0,
        context({runHarness: () => Promise.resolve(reply)}),
      );

      expect(record).toMatchObject({error, duration_ms: durationMs});
    }
  });

  it("treats a with_skill run that never used the skill as invalid evidence", async () => {
    const {runner, requests} = recordingRunner(() =>
      output({
        stdout: JSON.stringify({
          type: "result",
          result: "the answer",
          usage: {},
          duration_ms: 1,
        }),
      }),
    );

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({runHarness: runner}),
    );

    expect(record).toMatchObject({skill_triggered: false, pass_rate: 0});
    expect(requests).toHaveLength(1);
  });

  it("fails every assertion when the judge output cannot be read", async () => {
    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({
        runHarness: (request) =>
          Promise.resolve(
            output({
              stdout: isJudgeRequest(request)
                ? "not json"
                : claudeGeneration("the answer"),
            }),
          ),
      }),
    );

    expect(record).toMatchObject({
      pass_rate: 0,
      error: "unparseable judge output",
    });
  });

  it("fails every assertion when the judge process fails or times out", async () => {
    for (const [reply, error] of [
      [output({error: "boom"}), "judge process error: boom"],
      [output({timedOut: true}), "judge timeout"],
    ] as const) {
      const record = await runJob(
        evaluation("answer the question"),
        "with_skill",
        0,
        context({
          runHarness: (request) =>
            Promise.resolve(
              isJudgeRequest(request)
                ? reply
                : output({stdout: claudeGeneration("the answer")}),
            ),
        }),
      );

      expect(record).toMatchObject({pass_rate: 0, error});
    }
  });

  it("keeps each run of a multi-run sweep in its own directory", async () => {
    await runJob(
      evaluation("answer the question"),
      "with_skill",
      1,
      context({
        runs: 2,
        runHarness: (request) =>
          Promise.resolve(
            output({
              stdout: isJudgeRequest(request)
                ? claudeJudgement(true)
                : claudeGeneration("the answer"),
            }),
          ),
      }),
    );

    expect(readArm("with_skill", join("outputs", "response.md"), "run-2")).toBe(
      "the answer",
    );
  });
});
