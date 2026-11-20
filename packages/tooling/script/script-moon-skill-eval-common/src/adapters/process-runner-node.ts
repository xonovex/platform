import {spawn} from "node:child_process";
import {terminateProcess} from "@xonovex/script-moon-common/adapters/child-process";
import {
  foldLines,
  splitCompleteLines,
  type ProcessRunner,
} from "../process-runner.js";

/** The ProcessRunner that spawns a real process. */
export const nodeProcessRunner: ProcessRunner = (request) =>
  new Promise((resolvePromise) => {
    const child = spawn(request.executable, [...request.args], {
      stdio: ["ignore", "pipe", "pipe"],
      env: {...request.env},
      ...(request.cwd === undefined ? {} : {cwd: request.cwd}),
    });
    let state = request.lines.initial;
    let stdout = "";
    let stderr = "";
    let partialLine = "";
    let timedOut = false;
    let spawnError: string | null = null;
    let finished = false;
    let killTimer: NodeJS.Timeout | undefined;

    const terminate = (): void => {
      clearTimeout(killTimer);
      killTimer = terminateProcess(child);
    };
    const timeoutTimer = setTimeout(() => {
      timedOut = true;
      terminate();
    }, request.timeoutMs);

    const read = (lines: readonly string[]): void => {
      if (request.lines.settled(state)) return;
      state = foldLines(request.lines, state, lines);
      if (request.lines.settled(state)) terminate();
    };

    const finish = (code: number | null): void => {
      if (finished) return;
      finished = true;
      clearTimeout(timeoutTimer);
      clearTimeout(killTimer);
      resolvePromise({code, stdout, stderr, timedOut, spawnError, state});
    };

    child.stdout.setEncoding("utf8");
    child.stdout.on("data", (chunk: string) => {
      stdout += chunk;
      const {lines, rest} = splitCompleteLines(partialLine + chunk);
      partialLine = rest;
      read(lines);
    });
    child.stderr.setEncoding("utf8");
    child.stderr.on("data", (chunk: string) => {
      stderr += chunk;
    });
    child.on("error", (error) => {
      spawnError = error.message;
      finish(null);
    });
    child.on("close", (code) => {
      if (partialLine.length > 0) read([partialLine]);
      finish(code);
    });
  });
