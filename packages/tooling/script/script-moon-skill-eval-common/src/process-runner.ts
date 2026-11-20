import type {Environment} from "./environment.js";

/**
 * Reads stdout one line at a time while the process runs. A harness decides a
 * routing or output-limit outcome part way through its stream, and a settled
 * state ends the process there instead of paying for the rest of the run.
 */
export interface LineFold<State> {
  readonly initial: State;
  readonly step: (state: State, line: string) => State;
  readonly settled: (state: State) => boolean;
}

/** What to start, without the line fold, as a test double reads it. */
export interface ProcessInvocation {
  readonly executable: string;
  readonly args: readonly string[];
  readonly cwd: string | undefined;
  readonly env: Environment;
  readonly timeoutMs: number;
}

export interface ProcessRequest<State> extends ProcessInvocation {
  readonly lines: LineFold<State>;
}

/** How a process ended. spawnError is set when it never started. */
export interface ProcessResult<State> {
  readonly code: number | null;
  readonly stdout: string;
  readonly stderr: string;
  readonly timedOut: boolean;
  readonly spawnError: string | null;
  readonly state: State;
}

/**
 * Runs one process to its end. adapters/process-runner-node.ts spawns it, and
 * adapters/process-runner-memory.ts answers from a script, so the evaluators score
 * a whole sweep without a process.
 */
export type ProcessRunner = <State>(
  request: ProcessRequest<State>,
) => Promise<ProcessResult<State>>;

/** A fold for a process whose stream nothing reads before it ends. */
export const ignoreLines: LineFold<null> = {
  initial: null,
  step: (state) => state,
  settled: () => false,
};

/** Splits buffered text into its complete lines and the partial line after them. */
export const splitCompleteLines = (
  buffered: string,
): {readonly lines: readonly string[]; readonly rest: string} => {
  const parts = buffered.split(/\r?\n/);
  const rest = parts.pop() ?? "";
  return {lines: parts, rest};
};

/** Folds lines into the state, stopping at the first settled state. */
export const foldLines = <State>(
  fold: LineFold<State>,
  state: State,
  lines: readonly string[],
): State => {
  let current = state;
  for (const line of lines) {
    if (fold.settled(current)) return current;
    current = fold.step(current, line);
  }
  return current;
};
