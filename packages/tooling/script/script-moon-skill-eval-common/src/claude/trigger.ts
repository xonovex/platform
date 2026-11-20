import {stageClaudePluginArguments} from "@xonovex/script-moon-common/claude-plugin";
import type {HarnessPorts} from "../harness-ports.js";
import {
  TRIGGER_TIMEOUT_MS,
  type TriggerOutcome,
  type TriggerRequest,
} from "../trigger.js";
import {
  initialClaudeTriggerScan,
  isClaudeTriggerScanSettled,
  resolveClaudeTriggerOutcome,
  scanClaudeTriggerLine,
} from "./trigger-scan.js";

/**
 * Runs one Claude trigger check. The plugins on offer are staged in a private
 * directory with only the candidate skills, so the run sees no other skill.
 */
export const checkClaudeTrigger = async (
  request: TriggerRequest,
  ports: HarnessPorts,
): Promise<TriggerOutcome> => {
  const workspace = ports.scratch.makeTemporaryDirectory(
    "xonovex-claude-trigger-",
  );
  try {
    const args = stageClaudePluginArguments(
      request.args,
      workspace,
      request.candidates.map(({shortName}) => shortName),
      ports.fs,
    );
    const result = await ports.runProcess({
      executable: request.executable,
      args: [...args, request.query],
      cwd: undefined,
      env: ports.environment,
      timeoutMs: TRIGGER_TIMEOUT_MS,
      lines: {
        initial: initialClaudeTriggerScan,
        step: (scan, line) =>
          scanClaudeTriggerLine(scan, line, request.target, request.shortName),
        settled: isClaudeTriggerScanSettled,
      },
    });
    return resolveClaudeTriggerOutcome(result.state, result);
  } finally {
    ports.scratch.remove(workspace);
  }
};
