import {spawn} from "node:child_process";
import {appendFileSync, cpSync, mkdirSync, mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {terminateProcess} from "@xonovex/script-moon-common/child-process";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {CODEX_SKILL_SIGNAL, streamedAnswerLength} from "./output-parse.js";
import {claudeFailureDetail} from "./output-results.js";

/** What one harness invocation produced, as the generation and judge steps read it. */
export interface ProcessOutput {
  readonly stdout: string;
  readonly timedOut: boolean;
  readonly outputLimitExceeded: boolean;
  readonly error: string | null;
}

export interface HarnessRequest {
  readonly harness: "claude" | "codex";
  readonly args: readonly string[];
  readonly finalArgument: string;
  readonly cwd: string | undefined;
  readonly timeoutMs: number;
  // A ceiling on streamed answer characters, or null to let the run finish on its
  // own caps. Reaching it ends the run and reports outputLimitExceeded.
  readonly maxOutputCharacters: number | null;
  // A guide staged into the isolated Codex workspace, so the run can use the skill
  // under evaluation. Claude receives its skills through args instead.
  readonly skillGuide: string | undefined;
}

/**
 * Runs one harness invocation. The evaluator takes this as a port so its scoring
 * can be driven from recorded output without spawning anything.
 */
export type HarnessRunner = (request: HarnessRequest) => Promise<ProcessOutput>;

// Creates the throwaway HOME and CODEX_HOME a Codex run needs, plus the staged
// skill it is asked to use, and returns how the child should be started.
const isolateCodexWorkspace = (
  workspace: string,
  skillGuide: string | undefined,
): {readonly cwd: string; readonly env: NodeJS.ProcessEnv} => {
  const isolatedHome = join(workspace, "home");
  const isolatedCodexHome = join(workspace, ".codex");
  mkdirSync(isolatedHome, {recursive: true});
  mkdirSync(isolatedCodexHome, {recursive: true});
  if (skillGuide !== undefined) {
    const skillsRoot = join(isolatedHome, ".agents", "skills");
    const skillDirectory = join(skillsRoot, "target-skill");
    mkdirSync(skillsRoot, {recursive: true});
    cpSync(skillGuide, skillDirectory, {recursive: true});
    appendFileSync(
      join(skillDirectory, "SKILL.md"),
      `\n## Evaluation signal\n\nBegin the final response with ${CODEX_SKILL_SIGNAL} on its own line, then fulfill the request.\n`,
    );
  }
  return {
    cwd: workspace,
    env: {
      ...process.env,
      CODEX_HOME: isolatedCodexHome,
      HOME: isolatedHome,
    },
  };
};

/** The HarnessRunner that drives a real harness binary. */
export const spawnHarness: HarnessRunner = (request) =>
  new Promise((resolvePromise) => {
    const isolatedWorkspace =
      request.harness === "codex"
        ? mkdtempSync(join(tmpdir(), "xonovex-codex-output-"))
        : undefined;
    const isolation =
      isolatedWorkspace === undefined
        ? undefined
        : isolateCodexWorkspace(isolatedWorkspace, request.skillGuide);
    const childCwd = isolation?.cwd ?? request.cwd;
    const child = spawn(
      resolveExecutable(request.harness),
      [...request.args, request.finalArgument],
      {
        stdio: ["ignore", "pipe", "pipe"],
        ...(childCwd ? {cwd: childCwd} : {}),
        env: isolation?.env ?? process.env,
      },
    );
    let stdout = "";
    let stderr = "";
    let timedOut = false;
    let outputLimitExceeded = false;
    let partialLine = "";
    let outputCharacters = 0;
    let settled = false;
    let killTimer: NodeJS.Timeout | undefined;
    const finish = (error: string | null): void => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      clearTimeout(killTimer);
      if (isolatedWorkspace !== undefined) {
        rmSync(isolatedWorkspace, {recursive: true, force: true});
      }
      resolvePromise({stdout, timedOut, outputLimitExceeded, error});
    };
    const killChild = (): void => {
      clearTimeout(killTimer);
      killTimer = terminateProcess(child);
    };
    const timer = setTimeout(() => {
      timedOut = true;
      killChild();
    }, request.timeoutMs);
    child.stdout.setEncoding("utf8");
    child.stdout.on("data", (chunk: string) => {
      stdout += chunk;
      if (outputLimitExceeded || request.maxOutputCharacters === null) return;
      partialLine += chunk;
      const lines = partialLine.split(/\r?\n/);
      partialLine = lines.pop() ?? "";
      for (const line of lines) {
        let event: unknown;
        try {
          event = JSON.parse(line);
        } catch {
          continue;
        }
        outputCharacters += streamedAnswerLength(request.harness, event);
        if (outputCharacters > request.maxOutputCharacters) {
          outputLimitExceeded = true;
          killChild();
          break;
        }
      }
    });
    child.stderr.setEncoding("utf8");
    child.stderr.on("data", (chunk: string) => {
      stderr += chunk;
    });
    child.on("error", (error) => {
      finish(error.message);
    });
    child.on("close", (code) => {
      if (timedOut || code === 0) {
        finish(null);
        return;
      }
      const detail =
        stderr.trim() ||
        (request.harness === "claude" ? claudeFailureDetail(stdout) : "");
      const suffix = detail.length > 0 ? `: ${detail}` : "";
      finish(`${request.harness} exited ${String(code)}${suffix}`);
    });
  });
