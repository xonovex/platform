import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {
  memoryOutput,
  type MemoryOutput,
} from "@xonovex/script-moon-common/adapters/output-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  memoryProcessRunner,
  type MemoryProcessRunner,
  type ScriptedProcess,
} from "@xonovex/script-moon-skill-eval-common/adapters/process-runner-memory";
import {
  memoryScratch,
  type MemoryScratch,
} from "@xonovex/script-moon-skill-eval-common/adapters/scratch-memory";
import type {ProcessInvocation} from "@xonovex/script-moon-skill-eval-common/process-runner";
import {reporter} from "@xonovex/script-moon-skill-eval-common/report";
import {beforeEach, describe, expect, it} from "vitest";
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

const finalArgument = (invocation: ProcessInvocation): string =>
  invocation.args.at(-1) ?? "";

const isJudgeRequest = (invocation: ProcessInvocation): boolean =>
  finalArgument(invocation).includes("ASSISTANT RESPONSE:");

const codexMessage = (text: string): string =>
  JSON.stringify({type: "item.completed", item: {type: "agent_message", text}});

const ITERATION_DIRECTORY = "/workspace/iteration-1";

// A Claude run whose judge passes and whose generation answers correctly.
const healthyClaude = (invocation: ProcessInvocation): ScriptedProcess => ({
  stdout: isJudgeRequest(invocation)
    ? claudeJudgement(true)
    : claudeGeneration("the answer"),
});

const SKILL_DIRECTORY = join(ITERATION_DIRECTORY, "guide");

describe("runJob", () => {
  let tree: FileSystem;
  let scratch: MemoryScratch;
  let runner: MemoryProcessRunner;
  let output: MemoryOutput;

  // Answers every harness invocation from a script and records what it was asked
  // to run, which is what lets these cases score a job without a harness process.
  const answer = (
    reply: (invocation: ProcessInvocation) => ScriptedProcess,
  ): MemoryProcessRunner => {
    runner = memoryProcessRunner(reply);
    return runner;
  };

  beforeEach(() => {
    tree = memoryFileSystem({
      files: {[join(SKILL_DIRECTORY, "SKILL.md")]: "# Guide\n"},
    });
    scratch = memoryScratch(tree);
    output = memoryOutput();
    answer(() => ({}));
  });

  const context = (overrides: Partial<RunContext> = {}): RunContext => ({
    withArgs: ["--with"],
    withoutArgs: ["--without"],
    cwd: undefined,
    skillDirectory: SKILL_DIRECTORY,
    harness: "claude",
    executable: "/bin/harness",
    timeout: 2,
    target: "test-skill",
    shortName: "test-skill",
    judgeModel: "judge-model",
    judgeBudget: 0.01,
    iterationDirectory: ITERATION_DIRECTORY,
    fs: tree,
    scratch,
    runProcess: runner.run,
    environment: {},
    runs: 1,
    buildPrompt: (value) => value.prompt,
    now: () => 0,
    report: reporter(output),
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
    const {run} = answer(healthyClaude);

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({runProcess: run}),
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
    expect(output.stderrLines()).toEqual([
      "  [1/with_skill] pass_rate=1 tokens=5 (skill fired)",
    ]);
    expect(JSON.parse(readArm("with_skill", "timing.json"))).toMatchObject({
      total_tokens: 5,
      skill_triggered: true,
    });
  });

  it("sends the arm's own harness arguments with the prompt last", async () => {
    const {run, invocations} = answer(healthyClaude);

    await runJob(
      evaluation("answer the question"),
      "without_skill",
      0,
      context({runProcess: run}),
    );

    expect(invocations()[0]).toEqual({
      executable: "/bin/harness",
      args: ["--without", "answer the question"],
      cwd: undefined,
      env: {},
      timeoutMs: 2000,
    });
    expect(scratch.removed()).toEqual([]);
  });

  it.each([
    {answer: " ".repeat(3), error: "empty response"},
    {answer: "", error: "no-result"},
  ])(
    "fails an activated run with $error without calling the judge",
    async ({answer: text, error}) => {
      const {run, invocations} = answer(() => ({
        stdout: claudeGeneration(text),
      }));

      const record = await runJob(
        evaluation("answer the question"),
        "with_skill",
        0,
        context({runProcess: run}),
      );

      expect(invocations()).toHaveLength(1);
      expect(record).toMatchObject({
        pass_rate: 0,
        error,
        skill_triggered: true,
      });
    },
  );

  it("stages the skill for a Codex run that is expected to use the skill", async () => {
    const {run, invocations} = answer((invocation) => ({
      stdout: isJudgeRequest(invocation)
        ? codexMessage(
            '{"assertion_results":[{"passed":true,"evidence":"present"}]}',
          )
        : [
            codexMessage(`${CODEX_SKILL_SIGNAL}\nthe answer`),
            JSON.stringify({
              type: "turn.completed",
              usage: {input_tokens: 4, output_tokens: 2},
            }),
          ].join("\n"),
    }));

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({harness: "codex", runProcess: run}),
    );

    const workspace = "/scratch/xonovex-codex-output-1";
    expect(
      tree.readText(
        join(
          workspace,
          "home",
          ".agents",
          "skills",
          "target-skill",
          "SKILL.md",
        ),
      ),
    ).toContain(CODEX_SKILL_SIGNAL);
    expect(invocations()[0]).toMatchObject({
      cwd: workspace,
      env: {
        HOME: join(workspace, "home"),
        CODEX_HOME: join(workspace, ".codex"),
      },
    });
    expect(scratch.removed()).toEqual([
      workspace,
      "/scratch/xonovex-codex-output-2",
    ]);
    expect(record).toMatchObject({
      pass_rate: 1,
      tokens: 6,
      skill_triggered: true,
      error: null,
    });
  });

  it("records a generation process failure without grading it", async () => {
    const {run, invocations} = answer(() => ({
      code: 3,
      stderr: "generation failed",
    }));

    const record = await runJob(
      evaluation("answer the question"),
      "without_skill",
      0,
      context({runProcess: run}),
    );

    expect(record).toMatchObject({
      pass_rate: 0,
      skill_triggered: false,
      error: "claude exited 3: generation failed",
    });
    expect(invocations()).toHaveLength(1);
    expect(readArm("without_skill", "grading.json")).toContain(
      "not graded because generation evidence is invalid",
    );
  });

  it("reads a failed Claude exit's detail from its result when stderr is empty", async () => {
    const {run} = answer(() => ({
      code: 1,
      stdout: JSON.stringify({subtype: "error_max_turns"}),
    }));

    const record = await runJob(
      evaluation("answer the question"),
      "without_skill",
      0,
      context({runProcess: run}),
    );

    expect(record.error).toBe("claude exited 1: error_max_turns");
  });

  it("reports a generation that hit the output ceiling or the clock", async () => {
    for (const [harness, reply, error, durationMs] of [
      ["codex", {stdout: codexMessage("x".repeat(10_001))}, "output-limit", 0],
      ["claude", {timedOut: true, code: null}, "timeout", 2000],
    ] as const) {
      const {run} = answer(() => reply);
      const record = await runJob(
        evaluation("answer the question"),
        "without_skill",
        0,
        context({harness, runProcess: run}),
      );

      expect(record).toMatchObject({error, duration_ms: durationMs});
    }
  });

  it("treats a with_skill run that never used the skill as invalid evidence", async () => {
    const {run, invocations} = answer(() => ({
      stdout: JSON.stringify({
        type: "result",
        result: "the answer",
        usage: {},
        duration_ms: 1,
      }),
    }));

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({runProcess: run}),
    );

    expect(record).toMatchObject({skill_triggered: false, pass_rate: 0});
    expect(invocations()).toHaveLength(1);
  });

  it("fails every assertion when the judge output cannot be read", async () => {
    const {run} = answer((invocation) => ({
      stdout: isJudgeRequest(invocation)
        ? "not json"
        : claudeGeneration("the answer"),
    }));

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({runProcess: run}),
    );

    expect(record).toMatchObject({
      pass_rate: 0,
      error: "unparseable judge output",
    });
  });

  it("fails every assertion when the judge process fails or times out", async () => {
    for (const [reply, error] of [
      [{spawnError: "boom", code: null}, "judge process error: boom"],
      [{timedOut: true, code: null}, "judge timeout"],
    ] as const) {
      const {run} = answer((invocation) =>
        isJudgeRequest(invocation)
          ? reply
          : {stdout: claudeGeneration("the answer")},
      );
      const record = await runJob(
        evaluation("answer the question"),
        "with_skill",
        0,
        context({runProcess: run}),
      );

      expect(record).toMatchObject({pass_rate: 0, error});
    }
  });

  it("keeps each run of a multi-run sweep in its own directory", async () => {
    const {run} = answer(healthyClaude);

    await runJob(
      evaluation("answer the question"),
      "with_skill",
      1,
      context({runs: 2, runProcess: run}),
    );

    expect(readArm("with_skill", join("outputs", "response.md"), "run-2")).toBe(
      "the answer",
    );
  });

  it("stages the local plugin of a Claude run in its own workspace", async () => {
    const plugin = "/plugins/plugin-guide";
    tree.writeFile(
      join(plugin, ".claude-plugin", "plugin.json"),
      JSON.stringify({name: "xonovex-guide", skills: "./skills/"}),
    );
    tree.writeFile(join(plugin, "skills", "guide", "SKILL.md"), "# Guide\n");
    const {run, invocations} = answer(healthyClaude);

    await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context({withArgs: ["--plugin-dir", plugin], runProcess: run}),
    );

    const workspace = "/scratch/xonovex-claude-output-1";
    expect(invocations()[0]?.args).toEqual([
      "--plugin-dir",
      join(workspace, "plugin-1"),
      "/test-skill answer the question",
    ]);
    expect(scratch.removed()).toContain(workspace);
  });
});
