import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {nodeProcessRunner} from "../../../../src/adapters/process-runner-node.js";
import {ignoreLines, type LineFold} from "../../../../src/process-runner.js";

const collectUntil = (stop: string): LineFold<readonly string[]> => ({
  initial: [],
  step: (lines, line) => [...lines, line],
  settled: (lines) => lines.includes(stop),
});

const runNode = <State>(
  script: string,
  lines: LineFold<State>,
  timeoutMs = 10_000,
) =>
  nodeProcessRunner({
    executable: process.execPath,
    args: ["--eval", script],
    cwd: undefined,
    env: {PATH: process.env.PATH},
    timeoutMs,
    lines,
  });

describe("nodeProcessRunner", () => {
  it("returns stdout, stderr and the exit code", async () => {
    const result = await runNode(
      `process.stdout.write("out"); process.stderr.write("err"); process.exit(3);`,
      ignoreLines,
    );

    expect(result).toMatchObject({
      code: 3,
      stdout: "out",
      stderr: "err",
      timedOut: false,
      spawnError: null,
    });
  });

  it("folds each line, the final partial line included", async () => {
    const result = await runNode(
      `process.stdout.write("one\\r\\ntwo\\nthree");`,
      collectUntil("never"),
    );

    expect(result.state).toEqual(["one", "two", "three"]);
  });

  it("ends a process once the fold settles", async () => {
    const result = await runNode(
      `console.log("stop"); setInterval(() => {}, 1000);`,
      collectUntil("stop"),
    );

    expect(result.state).toEqual(["stop"]);
    expect(result.timedOut).toBe(false);
  });

  it("ends a process that outlives its timeout", async () => {
    const result = await runNode(
      `setInterval(() => {}, 1000);`,
      ignoreLines,
      200,
    );

    expect(result.timedOut).toBe(true);
  });

  it("passes the working directory and the environment", async () => {
    const directory = tmpdir();
    const result = await nodeProcessRunner({
      executable: process.execPath,
      args: [
        "--eval",
        `process.stdout.write(process.cwd() + "|" + process.env.MARKER);`,
      ],
      cwd: directory,
      env: {MARKER: "set"},
      timeoutMs: 10_000,
      lines: ignoreLines,
    });

    expect(result.stdout).toBe(`${directory}|set`);
  });

  it("reports an executable that cannot start", async () => {
    const result = await nodeProcessRunner({
      executable: join(tmpdir(), "xonovex-missing-harness"),
      args: [],
      cwd: undefined,
      env: {},
      timeoutMs: 10_000,
      lines: ignoreLines,
    });

    expect(result.spawnError).toContain("ENOENT");
  });
});
