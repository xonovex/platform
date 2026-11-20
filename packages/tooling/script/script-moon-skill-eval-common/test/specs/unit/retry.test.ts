import {describe, expect, it} from "vitest";
import {
  runWithTransientRetry,
  TRANSIENT_RUN_ATTEMPTS,
  transientRetryNotice,
} from "../../../src/retry.js";

const scripted = (errors: readonly (string | null)[]) => {
  let calls = 0;
  return {
    run: () => {
      const error = errors[calls] ?? null;
      calls += 1;
      return Promise.resolve({error, call: calls});
    },
    calls: () => calls,
  };
};

describe("runWithTransientRetry", () => {
  it("retries a failed run until one succeeds", async () => {
    const run = scripted(["flaky", null]);
    const notices: string[] = [];

    const result = await runWithTransientRetry(run.run, (attempt, error) => {
      notices.push(transientRetryNotice(attempt, error));
    });

    expect(result).toEqual({error: null, call: 2});
    expect(notices).toEqual(["retrying after transient failure (1/3): flaky"]);
  });

  it("returns the last failure once every attempt failed", async () => {
    const run = scripted(["a", "b", "c", null]);

    const retried: number[] = [];

    const result = await runWithTransientRetry(run.run, (attempt) => {
      retried.push(attempt);
    });

    expect(result).toEqual({error: "c", call: TRANSIENT_RUN_ATTEMPTS});
    expect(retried).toEqual([1, 2]);
    expect(run.calls()).toBe(TRANSIENT_RUN_ATTEMPTS);
  });
});
