import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {describe, expect, it} from "vitest";
import {
  EVAL_EXIT,
  reporter,
  runEvaluator,
  verdictExit,
} from "../../../src/report.js";

describe("reporter", () => {
  it("writes progress and errors to stderr and records to stdout", () => {
    const output = memoryOutput();
    const report = reporter(output);

    report.info("batch 1/1");
    report.error("broken");
    report.result({pass: true});

    expect(output.stderrLines()).toEqual(["batch 1/1", "Error: broken"]);
    expect(output.stdoutLines()).toEqual(['{"pass":true}']);
  });
});

describe("verdictExit", () => {
  it("maps a verdict to its exit code", () => {
    expect([verdictExit(true), verdictExit(false)]).toEqual([0, 1]);
  });
});

describe("runEvaluator", () => {
  it("returns the exit code of a run that completes", async () => {
    await expect(
      runEvaluator(() => Promise.resolve(1), memoryOutput()),
    ).resolves.toBe(1);
  });

  it("reports a thrown failure as an invalid run", async () => {
    const output = memoryOutput();

    await expect(
      runEvaluator(() => Promise.reject(new Error("no catalog")), output),
    ).resolves.toBe(EVAL_EXIT.invalid);
    expect(output.stderrLines()).toEqual(["Error: no catalog"]);
  });
});
