// A run that never invokes the skill would otherwise answer the whole query, which
// dominates trigger-eval spend. Once this many answer characters have streamed with
// no invocation, the routing outcome is settled: the response was written without
// the skill, so the run ends early and scores as a non-trigger. It is conclusive
// evidence, not an infrastructure failure: a query whose plain answer is long (any
// "write this code" negative) would otherwise invalidate its skill's whole sweep on
// every attempt, since the length is a property of the query rather than a flake.
export const TRIGGER_OUTPUT_LIMIT = 2000;

export const TRIGGER_TIMEOUT_MS = 60_000;

export interface TriggerOutcome {
  readonly triggered: boolean;
  readonly error: string | null;
  // How the run ended, recorded per run so a later triage reads the evidence instead
  // of re-probing the model to recover which skill answered.
  readonly selected?: string;
}

/** How a harness process ended, as the outcome resolvers read it. */
export interface TriggerExit {
  readonly code: number | null;
  readonly stderr: string;
  readonly timedOut: boolean;
  readonly spawnError: string | null;
}

/**
 * The infrastructure failure a finished harness process reports, or null when it
 * started and exited cleanly. harnessName names the binary in the message.
 */
export const processFailure = (
  harnessName: string,
  exit: TriggerExit,
): string | null => {
  if (exit.spawnError !== null) return exit.spawnError;
  if (exit.code === 0) return null;
  const detail = exit.stderr.trim();
  const suffix = detail.length > 0 ? `: ${detail}` : "";
  return `${harnessName} exited ${String(exit.code)}${suffix}`;
};

/** A skill a trigger run offers the model. */
export interface TriggerCandidate {
  readonly skillDirectory: string;
  readonly shortName: string;
}

/** One trigger run: which query, which target, and which skills are on offer. */
export interface TriggerRequest {
  readonly executable: string;
  readonly args: readonly string[];
  readonly query: string;
  // The target as the evaluator names it, bare or plugin-namespaced.
  readonly target: string;
  readonly shortName: string;
  // Every skill the run loads, the target included.
  readonly candidates: readonly TriggerCandidate[];
}
