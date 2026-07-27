import {spawn} from "node:child_process";
import {appendFileSync, cpSync, mkdirSync, mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {createInterface} from "node:readline";
import {streamTextDeltaLength} from "./validation.js";

const TRIGGER_TIMEOUT_MS = 60_000;

// A run that never invokes the skill would otherwise answer the whole query, which
// dominates trigger-eval spend. Once this many answer characters have streamed with
// no invocation, the routing outcome is settled — the response was written without
// the skill — so the run ends early and scores as a non-trigger. It is conclusive
// evidence, not an infrastructure failure: a query whose plain answer is long (any
// "write this code" negative) would otherwise invalidate its skill's whole sweep on
// every attempt, since the length is a property of the query rather than a flake.
export const TRIGGER_OUTPUT_LIMIT = 2000;

export interface TriggerOutcome {
  readonly triggered: boolean;
  readonly error: string | null;
  // How the run ended, recorded per run so a later triage reads the evidence instead
  // of re-probing the model to recover which skill answered.
  readonly selected?: string;
}

interface CodexTriggerOptions {
  readonly args: readonly string[];
  readonly executable: string;
  readonly guideDirectory: string;
  readonly query: string;
  readonly shortName: string;
  readonly candidateGuides?: readonly CodexCandidateGuide[];
}

export interface CodexCandidateGuide {
  readonly guideDirectory: string;
  readonly shortName: string;
}

const CODEX_TRIGGER_SIGNAL = "XONOVEX_SKILL_TRIGGERED";

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const matchSkill = (
  skillField: unknown,
  target: string,
  short: string,
): boolean => {
  if (typeof skillField !== "string") {
    return false;
  }
  return (
    skillField === target ||
    skillField === short ||
    skillField.endsWith(`:${short}`)
  );
};

// Every skill the line invokes, whether the harness allowed the call or denied it.
// The names matter beyond the target's own: a line that invokes some other skill
// settles the routing outcome just as firmly as one that invokes the target.
export const invokedSkillNames = (line: string): readonly string[] => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return [];
  }
  if (!isRecord(obj)) {
    return [];
  }

  const names: string[] = [];
  const message = obj.message;
  if (isRecord(message)) {
    const content = Array.isArray(message.content) ? message.content : [];
    for (const item of content) {
      if (isRecord(item) && item.type === "tool_use" && item.name === "Skill") {
        const inputField = item.input;
        if (isRecord(inputField) && typeof inputField.skill === "string") {
          names.push(inputField.skill);
        }
      }
    }
  }

  const denials = Array.isArray(obj.permission_denials)
    ? obj.permission_denials
    : [];
  for (const denial of denials) {
    if (isRecord(denial) && denial.tool_name === "Skill") {
      const toolInput = denial.tool_input;
      if (isRecord(toolInput) && typeof toolInput.skill === "string") {
        names.push(toolInput.skill);
      }
    }
  }

  return names;
};

const skillAvailableLine = (
  line: string,
  target: string,
  short: string,
): boolean => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return false;
  }
  if (!isRecord(obj) || obj.type !== "system" || obj.subtype !== "init") {
    return false;
  }
  const skills = Array.isArray(obj.skills) ? obj.skills : [];
  return skills.some((skill) => matchSkill(skill, target, short));
};

const textDeltaLength = (line: string): number => {
  try {
    return streamTextDeltaLength(JSON.parse(line));
  } catch {
    return 0;
  }
};

// What the harness did with a Skill call. An invocation naming something the harness
// cannot resolve launches nothing, so it says nothing about which skill the request
// belongs to; only a launch settles that. The two are indistinguishable from the
// tool_use line alone, which carries whatever name the model wrote.
export type ToolResultOutcome = "launched" | "rejected";

export const toolResultOutcome = (line: string): ToolResultOutcome | null => {
  let obj: unknown;
  try {
    obj = JSON.parse(line);
  } catch {
    return null;
  }
  if (!isRecord(obj)) {
    return null;
  }
  const message = obj.message;
  if (!isRecord(message)) {
    return null;
  }
  const content = Array.isArray(message.content) ? message.content : [];
  for (const item of content) {
    if (!isRecord(item) || item.type !== "tool_result") {
      continue;
    }
    if (item.is_error === true) {
      return "rejected";
    }
    const text =
      typeof item.content === "string"
        ? item.content
        : JSON.stringify(item.content);
    return text.includes("tool_use_error") ? "rejected" : "launched";
  }
  return null;
};

export const checkTriggered = (
  query: string,
  claudeArgs: readonly string[],
  target: string,
  short: string,
  claudeExecutable: string,
): Promise<TriggerOutcome> =>
  new Promise((resolvePromise) => {
    const proc = spawn(claudeExecutable, [...claudeArgs, query], {
      stdio: ["ignore", "pipe", "pipe"],
    });

    let matched = false;
    let targetAvailable = false;
    let timedOut = false;
    let pendingSkill: string | null = null;
    let competingSkill: string | null = null;
    let outputLimitExceeded = false;
    let outputCharacters = 0;
    let stderr = "";
    let spawnError: string | null = null;
    let settled = false;
    let killTimer: NodeJS.Timeout | undefined;

    const rl = createInterface({input: proc.stdout, crlfDelay: Infinity});

    const finish = (outcome: TriggerOutcome): void => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timeoutTimer);
      if (killTimer) {
        clearTimeout(killTimer);
      }
      resolvePromise(outcome);
    };

    const timeoutTimer = setTimeout(() => {
      timedOut = true;
      killProc();
    }, TRIGGER_TIMEOUT_MS);

    const killProc = (): void => {
      if (proc.exitCode === null && proc.signalCode === null) {
        proc.kill("SIGKILL");
        killTimer = setTimeout(() => proc.kill("SIGKILL"), 5000);
      }
    };

    rl.on("line", (raw) => {
      if (matched) {
        return;
      }
      const line = raw.trim();
      if (!line) {
        return;
      }
      targetAvailable ||= skillAvailableLine(line, target, short);
      const invoked = invokedSkillNames(line);
      if (invoked.some((skill) => matchSkill(skill, target, short))) {
        matched = true;
        rl.close();
        killProc();
        return;
      }
      // Some other name was invoked. Whether it settles the run depends on what the
      // harness did with it, which arrives on a later line, so hold it and read on.
      if (invoked.length > 0) {
        pendingSkill = invoked[0] ?? "";
        return;
      }
      if (pendingSkill !== null) {
        const outcome = toolResultOutcome(line);
        if (outcome === "launched") {
          // Another skill answered the request, so the target lost it. Ending here
          // spends no further turns on a settled result.
          competingSkill = pendingSkill;
          rl.close();
          killProc();
          return;
        }
        if (outcome === "rejected") {
          // The harness resolved nothing, so no skill answered and the run is not
          // settled. The model still has a turn to name a skill that exists.
          pendingSkill = null;
        }
        return;
      }
      outputCharacters += textDeltaLength(line);
      if (outputCharacters > TRIGGER_OUTPUT_LIMIT) {
        outputLimitExceeded = true;
        rl.close();
        killProc();
      }
    });

    proc.stderr.setEncoding("utf8");
    proc.stderr.on("data", (chunk: string) => {
      stderr += chunk;
    });

    proc.on("error", (error) => {
      spawnError = error.message;
    });

    proc.on("close", (code) => {
      if (matched) {
        finish({triggered: true, error: null, selected: "target"});
        return;
      }
      if (timedOut) {
        finish({triggered: false, error: "timeout"});
        return;
      }
      if (competingSkill !== null || outputLimitExceeded) {
        const selected =
          competingSkill === null
            ? "output-limit"
            : `competitor:${competingSkill}`;
        finish(
          targetAvailable
            ? {triggered: false, error: null, selected}
            : {triggered: false, error: "target skill unavailable"},
        );
        return;
      }
      if (spawnError !== null) {
        finish({triggered: false, error: spawnError});
        return;
      }
      if (code !== 0) {
        const detail = stderr.trim();
        const detailSuffix = detail.length > 0 ? `: ${detail}` : "";
        finish({
          triggered: false,
          error: `claude exited ${String(code)}${detailSuffix}`,
        });
        return;
      }
      if (!targetAvailable) {
        finish({triggered: false, error: "target skill unavailable"});
        return;
      }
      finish({triggered: false, error: null, selected: "none"});
    });
  });

const codexAgentMessage = (line: string): string => {
  try {
    const event: unknown = JSON.parse(line);
    if (!isRecord(event) || event.type !== "item.completed") return "";
    const item = event.item;
    return isRecord(item) &&
      item.type === "agent_message" &&
      typeof item.text === "string"
      ? item.text
      : "";
  } catch {
    return "";
  }
};

export const checkCodexTriggered = (
  options: CodexTriggerOptions,
): Promise<TriggerOutcome> => {
  const workspace = mkdtempSync(join(tmpdir(), "xonovex-codex-trigger-"));
  const skillsDirectory = join(workspace, ".agents", "skills");
  const isolatedHome = join(workspace, "home");
  const isolatedCodexHome = join(workspace, ".codex");
  mkdirSync(skillsDirectory, {recursive: true});
  mkdirSync(isolatedHome, {recursive: true});
  mkdirSync(isolatedCodexHome, {recursive: true});
  const candidateGuides = options.candidateGuides ?? [
    {guideDirectory: options.guideDirectory, shortName: options.shortName},
  ];
  for (const candidate of candidateGuides) {
    cpSync(
      candidate.guideDirectory,
      join(skillsDirectory, candidate.shortName),
      {recursive: true},
    );
  }
  const skillDirectory = join(skillsDirectory, options.shortName);
  appendFileSync(
    join(skillDirectory, "SKILL.md"),
    `\n## Evaluation signal\n\nWhen this skill applies, reply with exactly ${CODEX_TRIGGER_SIGNAL}.\n`,
  );

  const prompt =
    `${options.query}\n\nDecide only whether an available skill applies. ` +
    "Do not perform the task. If no skill applies, reply exactly NOT_APPLICABLE.";

  return new Promise((resolvePromise) => {
    const proc = spawn(options.executable, [...options.args, prompt], {
      cwd: workspace,
      env: {
        ...process.env,
        CODEX_HOME: isolatedCodexHome,
        HOME: isolatedHome,
      },
      stdio: ["ignore", "pipe", "pipe"],
    });
    let matched = false;
    let timedOut = false;
    let outputLimitExceeded = false;
    let outputCharacters = 0;
    let stderr = "";
    let spawnError: string | null = null;
    let settled = false;

    const finish = (outcome: TriggerOutcome): void => {
      if (settled) return;
      settled = true;
      clearTimeout(timeoutTimer);
      rmSync(workspace, {recursive: true, force: true});
      resolvePromise(outcome);
    };
    const killProc = (): void => {
      if (proc.exitCode === null && proc.signalCode === null) {
        proc.kill("SIGKILL");
      }
    };
    const timeoutTimer = setTimeout(() => {
      timedOut = true;
      killProc();
    }, TRIGGER_TIMEOUT_MS);
    const rl = createInterface({input: proc.stdout, crlfDelay: Infinity});
    rl.on("line", (raw) => {
      const message = codexAgentMessage(raw.trim());
      outputCharacters += message.length;
      if (message.includes(CODEX_TRIGGER_SIGNAL)) {
        matched = true;
        rl.close();
        killProc();
      } else if (outputCharacters > TRIGGER_OUTPUT_LIMIT) {
        outputLimitExceeded = true;
        rl.close();
        killProc();
      }
    });
    proc.stderr.setEncoding("utf8");
    proc.stderr.on("data", (chunk: string) => {
      stderr += chunk;
    });
    proc.on("error", (error) => {
      spawnError = error.message;
    });
    proc.on("close", (code) => {
      if (matched) {
        finish({triggered: true, error: null});
        return;
      }
      if (timedOut) {
        finish({triggered: false, error: "timeout"});
        return;
      }
      if (outputLimitExceeded) {
        finish({triggered: false, error: null});
        return;
      }
      if (spawnError !== null) {
        finish({triggered: false, error: spawnError});
        return;
      }
      if (code !== 0) {
        const detail = stderr.trim();
        const suffix = detail.length > 0 ? `: ${detail}` : "";
        finish({
          triggered: false,
          error: `codex exited ${String(code)}${suffix}`,
        });
        return;
      }
      finish({triggered: false, error: null});
    });
  });
};
