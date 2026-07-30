import {
  chmodSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {main} from "./evaluate.js";
import {runJob, type RunContext} from "./output-process.js";
import type {NormalizedEval} from "./validation.js";

const evaluation = (prompt: string): NormalizedEval => ({
  id: 1,
  prompt,
  expected_output: "the answer",
  assertions: ["response contains the answer"],
  files: [],
});

describe("runJob", () => {
  const originalPath = process.env.PATH;
  let temporaryDirectory: string;

  beforeEach(() => {
    temporaryDirectory = mkdtempSync(join(tmpdir(), "skill-eval-output-"));
    const executable = join(temporaryDirectory, "claude");
    writeFileSync(
      executable,
      String.raw`#!/usr/bin/env node
const args = process.argv.slice(2);
const finalArgument = args.at(-1) ?? "";
if (!args.includes("--json-schema") && finalArgument.includes("fail-generation")) {
  process.stderr.write("generation failed");
  process.exit(3);
}
if (args.includes("--json-schema")) {
  const passed = finalArgument.includes("ASSISTANT RESPONSE:\nthe answer");
  console.log(JSON.stringify({structured_output: {assertion_results: [{passed, evidence: passed ? "response contains the answer" : "answer missing"}]}}));
} else {
  const result = finalArgument.startsWith("/test-skill ") ? "the answer" : "wrong";
  console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
  console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
  console.log(JSON.stringify({type: "result", result, usage: {input_tokens: 3, output_tokens: 2}, duration_ms: 12}));
}
`,
    );
    chmodSync(executable, 0o755);
    process.env.PATH = `${temporaryDirectory}:${originalPath ?? ""}`;
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
  });

  afterEach(() => {
    process.env.PATH = originalPath;
    vi.restoreAllMocks();
    rmSync(temporaryDirectory, {recursive: true, force: true});
  });

  const context = (): RunContext => ({
    withArgs: [],
    withoutArgs: [],
    cwd: undefined,
    guideDirectory: temporaryDirectory,
    harness: "claude",
    timeout: 2,
    target: "test-skill",
    shortName: "test-skill",
    judgeModel: "judge-model",
    judgeBudget: 0.01,
    iterationDirectory: temporaryDirectory,
    runs: 1,
    buildPrompt: (value) => value.prompt,
  });

  it("generates, grades, and persists a healthy run", async () => {
    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      context(),
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
    expect(
      readFileSync(
        join(
          temporaryDirectory,
          "eval-1",
          "with_skill",
          "outputs",
          "response.md",
        ),
        "utf8",
      ),
    ).toBe("the answer");
  });

  it("records a generation process failure without grading", async () => {
    const record = await runJob(
      evaluation("fail-generation"),
      "without_skill",
      0,
      context(),
    );

    expect(record).toMatchObject({
      pass_rate: 0,
      skill_triggered: false,
      error: "claude exited 3: generation failed",
    });
    expect(
      readFileSync(
        join(temporaryDirectory, "eval-1", "without_skill", "grading.json"),
        "utf8",
      ),
    ).toContain("not graded because generation evidence is invalid");
  });

  it("runs generation and judging through the Codex harness", async () => {
    const codex = join(temporaryDirectory, "codex");
    const guideDirectory = join(temporaryDirectory, "guide");
    mkdirSync(guideDirectory);
    writeFileSync(
      join(guideDirectory, "SKILL.md"),
      "---\nname: test-skill\ndescription: Use for test prompts.\n---\n",
    );
    writeFileSync(
      codex,
      String.raw`#!/usr/bin/env node
const prompt = process.argv.at(-1) ?? "";
const text = prompt.includes("ASSISTANT RESPONSE:")
  ? '{"assertion_results":[{"passed":true,"evidence":"answer present"}]}'
  : "XONOVEX_SKILL_USED\nthe answer";
console.log(JSON.stringify({type: "item.completed", item: {type: "agent_message", text}}));
console.log(JSON.stringify({type: "turn.completed", usage: {input_tokens: 4, output_tokens: 2}}));
`,
    );
    chmodSync(codex, 0o755);
    const codexContext = {
      ...context(),
      guideDirectory,
      harness: "codex" as const,
    };

    const record = await runJob(
      evaluation("answer the question"),
      "with_skill",
      0,
      codexContext,
    );

    expect(record).toMatchObject({
      pass_rate: 1,
      tokens: 6,
      skill_triggered: true,
      error: null,
    });
  });

  it("runs the complete evaluator and writes benchmark evidence", async () => {
    const evaluationsFile = join(temporaryDirectory, "evals.json");
    const workspace = join(temporaryDirectory, "workspace");
    writeFileSync(
      evaluationsFile,
      JSON.stringify({
        skill_name: "test-skill",
        tier: "moderate",
        evals: [evaluation("question")],
      }),
    );

    const exitCode = await main([
      "--workspace",
      workspace,
      evaluationsFile,
      "test-skill",
      "iteration-1",
    ]);

    expect(exitCode).toBe(0);
    expect(
      JSON.parse(
        readFileSync(join(workspace, "iteration-1", "benchmark.json"), "utf8"),
      ),
    ).toMatchObject({
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

  it("fails a valid positive delta below the tier's absolute quality floor", async () => {
    const evaluationsFile = join(temporaryDirectory, "low-quality-evals.json");
    const workspace = join(temporaryDirectory, "low-quality-workspace");
    writeFileSync(
      evaluationsFile,
      JSON.stringify({
        skill_name: "test-skill",
        tier: "moderate",
        evals: [
          {
            ...evaluation("question"),
            assertions: ["one", "two", "three", "four"],
          },
        ],
      }),
    );

    const exitCode = await main([
      "--workspace",
      workspace,
      evaluationsFile,
      "test-skill",
      "iteration-1",
    ]);

    expect(exitCode).toBe(1);
    expect(
      JSON.parse(
        readFileSync(join(workspace, "iteration-1", "benchmark.json"), "utf8"),
      ),
    ).toMatchObject({
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
});
