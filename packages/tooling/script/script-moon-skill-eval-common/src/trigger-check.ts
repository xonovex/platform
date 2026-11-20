import type {SkillEvalHarness} from "@xonovex/script-moon-common/skill-eval-models";
import {checkClaudeTrigger} from "./claude/trigger.js";
import {checkCodexTrigger} from "./codex/trigger.js";
import type {HarnessPorts} from "./harness-ports.js";
import type {TriggerOutcome, TriggerRequest} from "./trigger.js";

/** Runs one trigger check through the selected harness. */
export const checkTrigger = (
  harness: SkillEvalHarness,
  request: TriggerRequest,
  ports: HarnessPorts,
): Promise<TriggerOutcome> =>
  harness === "claude"
    ? checkClaudeTrigger(request, ports)
    : checkCodexTrigger(request, ports);
