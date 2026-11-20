// A harness error is transient (a dropped stream, a timeout, a killed process), so
// a run is retried before its batch is discarded. A run that fails every attempt
// still invalidates the run: partial evidence must never be reported as a pass.
export const TRANSIENT_RUN_ATTEMPTS = 3;

export const transientRetryNotice = (attempt: number, error: string): string =>
  `retrying after transient failure (${String(attempt)}/${String(TRANSIENT_RUN_ATTEMPTS)}): ${error}`;

export const runWithTransientRetry = async <
  R extends {readonly error: string | null},
>(
  run: () => Promise<R>,
  onRetry: (attempt: number, error: string) => void,
): Promise<R> => {
  let result = await run();
  for (
    let attempt = 1;
    attempt < TRANSIENT_RUN_ATTEMPTS && result.error !== null;
    attempt += 1
  ) {
    onRetry(attempt, result.error);
    result = await run();
  }
  return result;
};
