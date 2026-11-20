import type {HarnessPorts} from "../harness-ports.js";
import {
  TRIGGER_TIMEOUT_MS,
  type TriggerOutcome,
  type TriggerRequest,
} from "../trigger.js";
import {
  CODEX_TRIGGER_SIGNAL,
  initialCodexTriggerScan,
  isCodexTriggerScanSettled,
  resolveCodexTriggerOutcome,
  scanCodexTriggerLine,
} from "./trigger-scan.js";
import {stageCodexWorkspace} from "./workspace.js";

// Only the target's staged copy carries the signal, so a competing skill that
// Codex follows instead reads as a non-trigger.
const SIGNAL_APPENDIX = `\n## Evaluation signal\n\nWhen this skill applies, reply with exactly ${CODEX_TRIGGER_SIGNAL}.\n`;

export const codexTriggerPrompt = (query: string): string =>
  `${query}\n\nDecide only whether an available skill applies. ` +
  "Do not perform the task. If no skill applies, reply exactly NOT_APPLICABLE.";

/**
 * Runs one Codex trigger check in a private workspace that holds the candidate
 * skills as project skills.
 */
export const checkCodexTrigger = async (
  request: TriggerRequest,
  ports: HarnessPorts,
): Promise<TriggerOutcome> => {
  const workspace = ports.scratch.makeTemporaryDirectory(
    "xonovex-codex-trigger-",
  );
  try {
    const staged = stageCodexWorkspace(
      workspace,
      "workspace",
      request.candidates.map((candidate) => ({
        source: candidate.skillDirectory,
        name: candidate.shortName,
        appendix:
          candidate.shortName === request.shortName
            ? SIGNAL_APPENDIX
            : undefined,
      })),
      ports.environment,
      ports.fs,
    );
    const result = await ports.runProcess({
      executable: request.executable,
      args: [...request.args, codexTriggerPrompt(request.query)],
      cwd: staged.cwd,
      env: staged.env,
      timeoutMs: TRIGGER_TIMEOUT_MS,
      lines: {
        initial: initialCodexTriggerScan,
        step: scanCodexTriggerLine,
        settled: isCodexTriggerScanSettled,
      },
    });
    return resolveCodexTriggerOutcome(result.state, result);
  } finally {
    ports.scratch.remove(workspace);
  }
};
