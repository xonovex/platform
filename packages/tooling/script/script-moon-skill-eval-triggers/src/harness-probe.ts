import type {HarnessPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import {ignoreLines} from "@xonovex/script-moon-skill-eval-common/process-runner";

// A probe that hangs must not stall the run before its first query.
export const HARNESS_PROBE_TIMEOUT_MS = 30_000;

/**
 * Whether the harness CLI starts at all. A file on PATH can still fail to spawn, and
 * that failure would otherwise surface as an infrastructure error on every query.
 */
export const harnessStarts = async (
  executable: string,
  ports: Pick<HarnessPorts, "runProcess" | "environment">,
): Promise<boolean> => {
  const probe = await ports.runProcess({
    executable,
    args: ["--version"],
    cwd: undefined,
    env: ports.environment,
    timeoutMs: HARNESS_PROBE_TIMEOUT_MS,
    lines: ignoreLines,
  });
  return probe.spawnError === null;
};
