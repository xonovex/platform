import {
  chmodSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {nodeEvaluatorPorts} from "@xonovex/script-moon-skill-eval-common/adapters/evaluator-ports-node";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {main, type OutputEvaluatorPorts} from "../../../src/evaluate.js";
import type {NormalizedEval} from "../../../src/validation.js";

// These drive the evaluator end to end against a harness binary on PATH. The
// scoring and persistence rules are covered without a process in test/specs/unit;
// what is left here is the node adapters: locating the harness executable,
// spawning it, and turning a whole sweep into benchmark evidence on disk.
const evaluation = (prompt: string): NormalizedEval => ({
  id: 1,
  prompt,
  expected_output: "the answer",
  assertions: ["response contains the answer"],
  files: [],
});

describe("the output evaluator", () => {
  let temporaryDirectory: string;
  let ports: OutputEvaluatorPorts;

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
    ports = {
      ...nodeEvaluatorPorts,
      environment: {
        ...process.env,
        PATH: `${temporaryDirectory}:${process.env.PATH ?? ""}`,
        SKILL_EVAL_HARNESS: undefined,
      },
      output: memoryOutput(),
      now: () => Date.now(),
    };
  });

  afterEach(() => {
    rmSync(temporaryDirectory, {recursive: true, force: true});
  });

  const runEvaluator = async (
    name: string,
    evaluations: readonly NormalizedEval[],
  ): Promise<{exitCode: number; benchmark: unknown}> => {
    const evaluationsFile = join(temporaryDirectory, `${name}.json`);
    const workspace = join(temporaryDirectory, name);
    writeFileSync(
      evaluationsFile,
      JSON.stringify({
        skill_name: "test-skill",
        tier: "moderate",
        evals: evaluations,
      }),
    );

    const exitCode = await main(
      ["--workspace", workspace, evaluationsFile, "test-skill", "iteration-1"],
      temporaryDirectory,
      ports,
    );

    return {
      exitCode,
      benchmark: JSON.parse(
        readFileSync(join(workspace, "iteration-1", "benchmark.json"), "utf8"),
      ),
    };
  };

  it("runs the complete evaluator and writes benchmark evidence", async () => {
    const {exitCode, benchmark} = await runEvaluator("evals", [
      evaluation("question"),
    ]);

    expect(exitCode).toBe(0);
    expect(benchmark).toMatchObject({
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

  it("reports a harness that exited nonzero as invalid evidence", async () => {
    const evaluationsFile = join(temporaryDirectory, "failing.json");
    const workspace = join(temporaryDirectory, "failing");
    writeFileSync(
      evaluationsFile,
      JSON.stringify({
        skill_name: "test-skill",
        tier: "moderate",
        evals: [evaluation("fail-generation")],
      }),
    );

    const exitCode = await main(
      ["--workspace", workspace, evaluationsFile, "test-skill", "iteration-1"],
      temporaryDirectory,
      ports,
    );

    expect(exitCode).toBe(2);
    expect(
      JSON.parse(
        readFileSync(
          join(workspace, "iteration-1", "invalid-run.json"),
          "utf8",
        ),
      ),
    ).toMatchObject({skill: "test-skill", status: "invalid"});
  });
});
