import type {Output} from "@xonovex/script-moon-common/output";

// An infrastructure failure invalidates the whole run, so it never shares the code
// of a failed verdict: a caller can tell missing evidence from a skill that missed.
export const EVAL_EXIT = {passed: 0, failed: 1, invalid: 2} as const;

export type EvalExit = (typeof EVAL_EXIT)[keyof typeof EVAL_EXIT];

export const verdictExit = (passed: boolean): EvalExit =>
  passed ? EVAL_EXIT.passed : EVAL_EXIT.failed;

/**
 * The lines an evaluator writes: progress and errors go to stderr, so stdout holds
 * only the JSON records a caller can pipe.
 */
export interface Reporter {
  readonly info: (line: string) => void;
  readonly error: (message: string) => void;
  readonly result: (record: unknown) => void;
}

export const reporter = (output: Output): Reporter => ({
  info: (line) => {
    output.stderr(line);
  },
  error: (message) => {
    output.stderr(`Error: ${message}`);
  },
  result: (record) => {
    output.stdout(JSON.stringify(record));
  },
});

/** Runs an evaluator and reports an unexpected failure as an invalid run. */
export const runEvaluator = async (
  evaluate: () => Promise<number>,
  output: Output,
): Promise<number> => {
  try {
    return await evaluate();
  } catch (error: unknown) {
    reporter(output).error(
      error instanceof Error ? error.message : String(error),
    );
    return EVAL_EXIT.invalid;
  }
};
