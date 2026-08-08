import {spawn} from "node:child_process";
import {appendFileSync, cpSync, mkdirSync, mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {createInterface} from "node:readline";
import {terminateProcess} from "@xonovex/script-moon-common/child-process";
import {
  CODEX_TRIGGER_SIGNAL,
  initialCodexTriggerScan,
  initialTriggerScan,
  isCodexTriggerScanSettled,
  isTriggerScanSettled,
  resolveCodexTriggerOutcome,
  resolveTriggerOutcome,
  scanCodexTriggerLine,
  scanTriggerLine,
  type CodexTriggerScan,
  type TriggerOutcome,
  type TriggerScan,
} from "./trigger-scan.js";

const TRIGGER_TIMEOUT_MS = 60_000;

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

    let scan: TriggerScan = initialTriggerScan;
    let timedOut = false;
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
      clearTimeout(killTimer);
      resolvePromise(outcome);
    };

    const timeoutTimer = setTimeout(() => {
      timedOut = true;
      killProc();
    }, TRIGGER_TIMEOUT_MS);

    const killProc = (): void => {
      clearTimeout(killTimer);
      killTimer = terminateProcess(proc);
    };

    rl.on("line", (raw) => {
      scan = scanTriggerLine(scan, raw, target, short);
      if (isTriggerScanSettled(scan)) {
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
      finish(resolveTriggerOutcome(scan, {code, stderr, timedOut, spawnError}));
    });
  });

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
    let scan: CodexTriggerScan = initialCodexTriggerScan;
    let timedOut = false;
    let stderr = "";
    let spawnError: string | null = null;
    let settled = false;
    let killTimer: NodeJS.Timeout | undefined;

    const finish = (outcome: TriggerOutcome): void => {
      if (settled) return;
      settled = true;
      clearTimeout(timeoutTimer);
      clearTimeout(killTimer);
      rmSync(workspace, {recursive: true, force: true});
      resolvePromise(outcome);
    };
    const killProc = (): void => {
      clearTimeout(killTimer);
      killTimer = terminateProcess(proc);
    };
    const timeoutTimer = setTimeout(() => {
      timedOut = true;
      killProc();
    }, TRIGGER_TIMEOUT_MS);
    const rl = createInterface({input: proc.stdout, crlfDelay: Infinity});
    rl.on("line", (raw) => {
      scan = scanCodexTriggerLine(scan, raw);
      if (isCodexTriggerScanSettled(scan)) {
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
      finish(
        resolveCodexTriggerOutcome(scan, {code, stderr, timedOut, spawnError}),
      );
    });
  });
};
