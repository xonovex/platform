import {describe, expect, it} from "vitest";
import {memoryProcessRunner} from "../../../../src/adapters/process-runner-memory.js";
import {ignoreLines, type LineFold} from "../../../../src/process-runner.js";

const collect: LineFold<readonly string[]> = {
  initial: [],
  step: (lines, line) => [...lines, line],
  settled: (lines) => lines.includes("stop"),
};

const request = {
  executable: "/bin/harness",
  args: ["--flag"],
  cwd: "/work",
  env: {HOME: "/home"},
  timeoutMs: 1000,
};

describe("memoryProcessRunner", () => {
  it("defaults to a clean exit with no output", async () => {
    const runner = memoryProcessRunner(() => ({}));

    await expect(runner.run({...request, lines: ignoreLines})).resolves.toEqual(
      {
        code: 0,
        stdout: "",
        stderr: "",
        timedOut: false,
        spawnError: null,
        state: null,
      },
    );
  });

  it("folds stdout up to the settled line, a final partial line included", async () => {
    const runner = memoryProcessRunner(() => ({stdout: "a\nstop\nb"}));
    const partial = memoryProcessRunner(() => ({stdout: "a\nb"}));

    await expect(
      runner.run({...request, lines: collect}),
    ).resolves.toMatchObject({state: ["a", "stop"], stdout: "a\nstop\nb"});
    await expect(
      partial.run({...request, lines: collect}),
    ).resolves.toMatchObject({state: ["a", "b"]});
  });

  it("returns the scripted failure fields", async () => {
    const runner = memoryProcessRunner(() => ({
      code: null,
      stderr: "boom",
      timedOut: true,
      spawnError: "spawn ENOENT",
    }));

    await expect(
      runner.run({...request, lines: ignoreLines}),
    ).resolves.toMatchObject({
      code: null,
      stderr: "boom",
      timedOut: true,
      spawnError: "spawn ENOENT",
    });
  });

  it("lists every invocation without its line fold", async () => {
    const runner = memoryProcessRunner(() => ({}));

    await runner.run({...request, lines: ignoreLines});

    expect(runner.invocations()).toEqual([request]);
  });
});
