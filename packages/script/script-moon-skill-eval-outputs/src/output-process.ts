import {spawn} from "node:child_process";
import {mkdirSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  claudeFailureDetail,
  extractJson,
  isRecord,
  summarize,
  sumTokens,
  type AssertionResult,
  type Graded,
  type JobRecord,
} from "./output-results.js";
import {
  buildGenerationPrompt,
  buildJudgeClaudeArgs,
  parseJudgeResults,
  streamTextDeltaLength,
  type EvaluationArm,
  type NormalizedEval,
} from "./validation.js";

interface GenerationResult {
  readonly text: string;
  readonly totalTokens: number;
  readonly durationMs: number;
  readonly skillTriggered: boolean;
  readonly error: string | null;
}

interface ProcessOutput {
  readonly stdout: string;
  readonly timedOut: boolean;
  readonly outputLimitExceeded: boolean;
  readonly error: string | null;
}

export interface RunContext {
  readonly withArgs: readonly string[];
  readonly withoutArgs: readonly string[];
  readonly cwd: string | undefined;
  readonly timeout: number;
  readonly target: string;
  readonly shortName: string;
  readonly judgeModel: string;
  readonly judgeBudget: number;
  readonly iterationDirectory: string;
  readonly runs: number;
  readonly buildPrompt: (evaluation: NormalizedEval) => string;
}

const matchSkill = (
  skillField: unknown,
  target: string,
  shortName: string,
): boolean =>
  typeof skillField === "string" &&
  (skillField === target ||
    skillField === shortName ||
    skillField.endsWith(`:${shortName}`));

const skillInvoked = (
  value: Record<string, unknown>,
  target: string,
  shortName: string,
): boolean => {
  const message = value.message;
  if (!isRecord(message)) return false;
  const content = Array.isArray(message.content) ? message.content : [];
  return content.some((item) => {
    if (!isRecord(item) || item.type !== "tool_use" || item.name !== "Skill") {
      return false;
    }
    return (
      isRecord(item.input) && matchSkill(item.input.skill, target, shortName)
    );
  });
};

const skillAvailable = (
  value: Record<string, unknown>,
  target: string,
  shortName: string,
): boolean =>
  value.type === "system" &&
  value.subtype === "init" &&
  Array.isArray(value.skills) &&
  value.skills.some((skill) => matchSkill(skill, target, shortName));

const runClaude = (
  args: readonly string[],
  finalArgument: string,
  cwd: string | undefined,
  timeoutMs: number,
  maxOutputCharacters: number | null = null,
): Promise<ProcessOutput> =>
  new Promise((resolvePromise) => {
    const child = spawn(resolveExecutable("claude"), [...args, finalArgument], {
      stdio: ["ignore", "pipe", "pipe"],
      ...(cwd ? {cwd} : {}),
    });
    let stdout = "";
    let stderr = "";
    let timedOut = false;
    let outputLimitExceeded = false;
    let partialLine = "";
    let outputCharacters = 0;
    let settled = false;
    const finish = (error: string | null): void => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      resolvePromise({stdout, timedOut, outputLimitExceeded, error});
    };
    const timer = setTimeout(() => {
      timedOut = true;
      child.kill("SIGKILL");
    }, timeoutMs);
    child.stdout.setEncoding("utf8");
    child.stdout.on("data", (chunk: string) => {
      stdout += chunk;
      if (maxOutputCharacters === null || outputLimitExceeded) return;
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
        outputCharacters += streamTextDeltaLength(event);
        if (outputCharacters > maxOutputCharacters) {
          outputLimitExceeded = true;
          child.kill("SIGKILL");
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
      const detail = stderr.trim() || claudeFailureDetail(stdout);
      const suffix = detail.length > 0 ? `: ${detail}` : "";
      finish(`claude exited ${String(code)}${suffix}`);
    });
  });

const generationFailure = (
  error: string,
  durationMs = 0,
): GenerationResult => ({
  text: "",
  totalTokens: 0,
  durationMs,
  skillTriggered: false,
  error,
});

const generate = async (
  prompt: string,
  claudeArgs: readonly string[],
  cwd: string | undefined,
  timeout: number,
  target: string,
  shortName: string,
  expectSkill: boolean,
): Promise<GenerationResult> => {
  const proc = await runClaude(claudeArgs, prompt, cwd, timeout * 1000, 10_000);
  if (proc.outputLimitExceeded) return generationFailure("output-limit");
  if (proc.timedOut) return generationFailure("timeout", timeout * 1000);
  if (proc.error) return generationFailure(proc.error);

  let text = "";
  let usage: unknown = {};
  let durationMs = 0;
  let invoked = false;
  let available = false;
  for (const rawLine of proc.stdout.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line) continue;
    let value: unknown;
    try {
      value = JSON.parse(line);
    } catch {
      continue;
    }
    if (!isRecord(value)) continue;
    available ||= skillAvailable(value, target, shortName);
    invoked ||= skillInvoked(value, target, shortName);
    if (value.type === "result") {
      text = typeof value.result === "string" ? value.result : "";
      usage = value.usage ?? {};
      durationMs =
        typeof value.duration_ms === "number" ? value.duration_ms : 0;
    }
  }

  return {
    text,
    totalTokens: sumTokens(usage),
    durationMs,
    skillTriggered: expectSkill && (invoked || available),
    error: text ? null : "no-result",
  };
};

const JUDGE_RUBRIC = `\
You are a strict output evaluator. Grade the ASSISTANT RESPONSE against each \
assertion independently.

Rules:
- Binary verdict per assertion: passed = true or false. No partial credit.
- Cite concrete evidence: quote the response or name the specific gap.
- No benefit of the doubt — vagueness, omission, or a hedge is FAIL.
- Judge ONLY against the assertion. Ignore response length, tone, and style.
- If the response lacks the information to decide, mark FAIL, evidence "insufficient".
- Use the EXPECTED OUTPUT only as a reference for what success looks like; the \
response need not match it word for word.

TASK PROMPT:
{prompt}

EXPECTED OUTPUT (reference):
{expected}

ASSERTIONS (grade each, in order):
{assertions}

ASSISTANT RESPONSE:
{response}

Return ONLY minified JSON, no markdown fences, one object per assertion in order:
{"assertion_results":[{"text":"<assertion>","passed":true,"evidence":"<quote or reason>"}]}
`;

const grade = async (
  prompt: string,
  expected: string,
  assertions: readonly string[],
  response: string,
  model: string,
  budget: number,
): Promise<Graded> => {
  const allFail = (reason: string): Graded =>
    summarize(
      assertions.map((assertion) => ({
        text: assertion,
        passed: false,
        evidence: reason,
      })),
      reason,
    );

  if (!response.trim()) return allFail("empty response");
  const numbered = assertions
    .map((assertion, index) => `${String(index + 1)}. ${assertion}`)
    .join("\n");
  const rubric = JUDGE_RUBRIC.replace("{prompt}", prompt)
    .replace("{expected}", expected || "(none provided)")
    .replace("{assertions}", numbered)
    .replace("{response}", response);
  const args = buildJudgeClaudeArgs({
    model,
    budget,
    assertionCount: assertions.length,
  });
  const proc = await runClaude(args, rubric, undefined, 300_000);
  if (proc.error) return allFail(`judge process error: ${proc.error}`);
  if (proc.timedOut) return allFail("judge timeout");

  let verdict: unknown = null;
  try {
    const outer: unknown = JSON.parse(proc.stdout);
    const structured = isRecord(outer) ? outer.structured_output : null;
    const resultText =
      isRecord(outer) && typeof outer.result === "string" ? outer.result : "";
    verdict = isRecord(structured) ? structured : extractJson(resultText);
  } catch {
    verdict = null;
  }
  const verdictResults = parseJudgeResults(verdict);
  if (verdictResults === undefined) return allFail("unparseable judge output");
  const results: AssertionResult[] = assertions.map((assertion, index) => ({
    text: assertion,
    passed: verdictResults[index]?.passed ?? false,
    evidence: verdictResults[index]?.evidence ?? "no evidence",
  }));
  return summarize(results);
};

export const runJob = async (
  evaluation: NormalizedEval,
  arm: EvaluationArm,
  runIndex: number,
  context: RunContext,
): Promise<JobRecord> => {
  const prompt = buildGenerationPrompt(
    context.buildPrompt(evaluation),
    arm,
    context.target,
  );
  const args = arm === "with_skill" ? context.withArgs : context.withoutArgs;
  const generation = await generate(
    prompt,
    args,
    context.cwd,
    context.timeout,
    context.target,
    context.shortName,
    arm === "with_skill",
  );
  const healthy =
    generation.error === null &&
    (arm !== "with_skill" || generation.skillTriggered);
  const graded = healthy
    ? await grade(
        evaluation.prompt,
        evaluation.expected_output,
        evaluation.assertions,
        generation.text,
        context.judgeModel,
        context.judgeBudget,
      )
    : summarize(
        evaluation.assertions.map((assertion) => ({
          text: assertion,
          passed: false,
          evidence: "not graded because generation evidence is invalid",
        })),
      );

  let armDirectory = join(
    context.iterationDirectory,
    `eval-${String(evaluation.id)}`,
    arm,
  );
  if (context.runs > 1) {
    armDirectory = join(armDirectory, `run-${String(runIndex + 1)}`);
  }
  mkdirSync(join(armDirectory, "outputs"), {recursive: true});
  writeFileSync(join(armDirectory, "outputs", "response.md"), generation.text, {
    encoding: "utf8",
  });
  writeFileSync(
    join(armDirectory, "timing.json"),
    JSON.stringify(
      {
        total_tokens: generation.totalTokens,
        duration_ms: generation.durationMs,
        skill_triggered: generation.skillTriggered,
        error: generation.error,
      },
      null,
      2,
    ),
    {encoding: "utf8"},
  );
  writeFileSync(
    join(armDirectory, "grading.json"),
    JSON.stringify(graded, null, 2),
    {encoding: "utf8"},
  );

  const runTag = context.runs === 1 ? "" : `/run-${String(runIndex + 1)}`;
  const skillTag =
    arm === "with_skill" && generation.skillTriggered ? " (skill fired)" : "";
  const errorTag = generation.error ? ` [${generation.error}]` : "";
  process.stderr.write(
    `  [${String(evaluation.id)}/${arm}${runTag}] ` +
      `pass_rate=${String(graded.summary.pass_rate)} tokens=${String(generation.totalTokens)}` +
      `${skillTag}${errorTag}\n`,
  );

  return {
    id: evaluation.id,
    arm,
    pass_rate: graded.summary.pass_rate,
    tokens: generation.totalTokens,
    duration_ms: generation.durationMs,
    skill_triggered: generation.skillTriggered,
    error: generation.error ?? graded.error,
  };
};
