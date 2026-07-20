import {
  chmodSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {runJob, type RunContext} from "./output-process.js";
import type {NormalizedEval} from "./validation.js";

describe("runJob", () => {
  const originalPath = process.env.PATH;
  let temporaryDirectory: string;

  beforeEach(() => {
    temporaryDirectory = mkdtempSync(join(tmpdir(), "skill-eval-output-"));
    const executable = join(temporaryDirectory, "claude");
    writeFileSync(
      executable,
      `#!/usr/bin/env node
const args = process.argv.slice(2);
const finalArgument = args.at(-1) ?? "";
if (!args.includes("--json-schema") && finalArgument.includes("fail-generation")) {
  process.stderr.write("generation failed");
  process.exit(3);
}
if (args.includes("--json-schema")) {
  console.log(JSON.stringify({structured_output: {assertion_results: [{passed: true, evidence: "response contains the answer"}]}}));
} else {
  console.log(JSON.stringify({type: "system", subtype: "init", skills: ["test-skill"]}));
  console.log(JSON.stringify({message: {content: [{type: "tool_use", name: "Skill", input: {skill: "test-skill"}}]}}));
  console.log(JSON.stringify({type: "result", result: "the answer", usage: {input_tokens: 3, output_tokens: 2}, duration_ms: 12}));
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

  const evaluation = (prompt: string): NormalizedEval => ({
    id: 1,
    prompt,
    expected_output: "the answer",
    assertions: ["response contains the answer"],
    files: [],
  });

  const context = (): RunContext => ({
    withArgs: [],
    withoutArgs: [],
    cwd: undefined,
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
});
