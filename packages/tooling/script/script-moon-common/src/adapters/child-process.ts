import {type ChildProcess} from "node:child_process";

export const TERMINATE_ESCALATION_MS = 5000;

// Ends a harness run whose outcome the caller has already settled. SIGTERM first, so the
// harness flushes its stream and reaps the processes it spawned; the escalation forces
// the exit when it does not. The timer is unreferenced because a settled run must never
// hold the caller's event loop open waiting for a process that has already gone.
//
// The returned handle is the caller's to clear once the child closes; a caller that may
// terminate the same child twice clears the previous handle before arming another.
export const terminateProcess = (
  proc: ChildProcess,
  escalationMs = TERMINATE_ESCALATION_MS,
): NodeJS.Timeout | undefined => {
  if (proc.exitCode !== null || proc.signalCode !== null) {
    return undefined;
  }
  proc.kill("SIGTERM");
  return setTimeout(() => proc.kill("SIGKILL"), escalationMs).unref();
};
