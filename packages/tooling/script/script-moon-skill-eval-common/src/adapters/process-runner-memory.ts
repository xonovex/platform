import {
  foldLines,
  splitCompleteLines,
  type ProcessInvocation,
  type ProcessRunner,
} from "../process-runner.js";

/** How a scripted process ends; an absent field takes a clean exit's value. */
export interface ScriptedProcess {
  readonly stdout?: string;
  readonly stderr?: string;
  readonly code?: number | null;
  readonly timedOut?: boolean;
  readonly spawnError?: string | null;
}

/** A ProcessRunner that also lists every process it was asked to start. */
export interface MemoryProcessRunner {
  readonly run: ProcessRunner;
  readonly invocations: () => readonly ProcessInvocation[];
}

/**
 * A ProcessRunner that answers each request from a script instead of spawning.
 * It folds the scripted stdout through the request's line fold the way a live
 * stream is read, so a case drives the same early settlement a real run does.
 */
export const memoryProcessRunner = (
  answer: (invocation: ProcessInvocation) => ScriptedProcess,
): MemoryProcessRunner => {
  const invocations: ProcessInvocation[] = [];
  const run: ProcessRunner = (request) => {
    const invocation: ProcessInvocation = {
      executable: request.executable,
      args: request.args,
      cwd: request.cwd,
      env: request.env,
      timeoutMs: request.timeoutMs,
    };
    invocations.push(invocation);
    const scripted = answer(invocation);
    const stdout = scripted.stdout ?? "";
    const {lines, rest} = splitCompleteLines(stdout);
    const completeLines = rest.length > 0 ? [...lines, rest] : lines;
    return Promise.resolve({
      code: scripted.code === undefined ? 0 : scripted.code,
      stdout,
      stderr: scripted.stderr ?? "",
      timedOut: scripted.timedOut ?? false,
      spawnError: scripted.spawnError ?? null,
      state: foldLines(request.lines, request.lines.initial, completeLines),
    });
  };
  return {run, invocations: () => [...invocations]};
};
