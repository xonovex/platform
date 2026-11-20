import {basename} from "node:path";
import {stageClaudePluginArguments} from "@xonovex/script-moon-common/claude-plugin";
import type {SkillEvalHarness} from "@xonovex/script-moon-common/skill-eval-models";
import {stageCodexWorkspace} from "@xonovex/script-moon-skill-eval-common/codex/workspace";
import type {Environment} from "@xonovex/script-moon-skill-eval-common/environment";
import type {HarnessPorts} from "@xonovex/script-moon-skill-eval-common/harness-ports";
import type {
  LineFold,
  ProcessResult,
} from "@xonovex/script-moon-skill-eval-common/process-runner";
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
  readonly harness: SkillEvalHarness;
  readonly executable: string;
  readonly args: readonly string[];
  readonly finalArgument: string;
  readonly cwd: string | undefined;
  readonly timeoutMs: number;
  // A ceiling on streamed answer characters, or null to let the run finish on its
  // own caps. Reaching it ends the run and reports outputLimitExceeded.
  readonly maxOutputCharacters: number | null;
  // skillSource selects the skill copied into a harness's evaluation workspace.
  readonly skillSource: string | undefined;
}

interface AnswerLength {
  readonly characters: number;
  readonly exceeded: boolean;
}

const NO_ANSWER: AnswerLength = {characters: 0, exceeded: false};

const answerLength = (
  harness: SkillEvalHarness,
  maxOutputCharacters: number | null,
): LineFold<AnswerLength> => ({
  initial: NO_ANSWER,
  step: (state, line) => {
    if (maxOutputCharacters === null) return state;
    let event: unknown;
    try {
      event = JSON.parse(line);
    } catch {
      return state;
    }
    const characters = state.characters + streamedAnswerLength(harness, event);
    return {characters, exceeded: characters > maxOutputCharacters};
  },
  settled: (state) => state.exceeded,
});

const SIGNAL_APPENDIX = `\n## Evaluation signal\n\nBegin the final response with ${CODEX_SKILL_SIGNAL} on its own line, then fulfill the request.\n`;

// A Codex run always gets an isolated home; a Claude run gets a private directory
// only when it loads a local plugin, whose skills are staged there.
const workspacePrefix = (request: HarnessRequest): string | undefined => {
  if (request.harness === "codex") return "xonovex-codex-output-";
  return request.args.includes("--plugin-dir")
    ? "xonovex-claude-output-"
    : undefined;
};

interface StagedRun {
  readonly args: readonly string[];
  readonly cwd: string | undefined;
  readonly env: Environment;
}

const stageRun = (
  request: HarnessRequest,
  workspace: string | undefined,
  ports: HarnessPorts,
): StagedRun => {
  if (workspace === undefined) {
    return {args: request.args, cwd: request.cwd, env: ports.environment};
  }
  if (request.harness === "codex") {
    const staged = stageCodexWorkspace(
      workspace,
      "home",
      request.skillSource === undefined
        ? []
        : [
            {
              source: request.skillSource,
              name: "target-skill",
              appendix: SIGNAL_APPENDIX,
            },
          ],
      ports.environment,
      ports.fs,
    );
    return {args: request.args, cwd: staged.cwd, env: staged.env};
  }
  return {
    args: stageClaudePluginArguments(
      request.args,
      workspace,
      request.skillSource === undefined ? [] : [basename(request.skillSource)],
      ports.fs,
    ),
    cwd: request.cwd,
    env: ports.environment,
  };
};

const harnessError = (
  harness: SkillEvalHarness,
  result: ProcessResult<AnswerLength>,
): string | null => {
  if (result.spawnError !== null) return result.spawnError;
  if (result.timedOut || result.code === 0) return null;
  const detail =
    result.stderr.trim() ||
    (harness === "claude" ? claudeFailureDetail(result.stdout) : "");
  const suffix = detail.length > 0 ? `: ${detail}` : "";
  return `${harness} exited ${String(result.code)}${suffix}`;
};

/** Runs one generation or judge invocation in its staged workspace. */
export const runHarness = async (
  request: HarnessRequest,
  ports: HarnessPorts,
): Promise<ProcessOutput> => {
  const prefix = workspacePrefix(request);
  const workspace =
    prefix === undefined
      ? undefined
      : ports.scratch.makeTemporaryDirectory(prefix);
  try {
    const staged = stageRun(request, workspace, ports);
    const result = await ports.runProcess({
      executable: request.executable,
      args: [...staged.args, request.finalArgument],
      cwd: staged.cwd,
      env: staged.env,
      timeoutMs: request.timeoutMs,
      lines: answerLength(request.harness, request.maxOutputCharacters),
    });
    return {
      stdout: result.stdout,
      timedOut: result.timedOut,
      outputLimitExceeded: result.state.exceeded,
      error: harnessError(request.harness, result),
    };
  } finally {
    if (workspace !== undefined) ports.scratch.remove(workspace);
  }
};
