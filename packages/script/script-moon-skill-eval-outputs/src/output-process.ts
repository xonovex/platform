import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {buildIsolatedCodexArgs} from "@xonovex/script-moon-skill-eval-common/validation";
import type {HarnessRunner} from "./output-harness.js";
import {
  GENERATION_OUTPUT_LIMIT,
  parseClaudeGeneration,
  parseCodexGeneration,
  parseJudgeVerdict,
} from "./output-parse.js";
import {
  summarize,
  sumTokens,
  type AssertionResult,
  type Graded,
  type JobRecord,
} from "./output-results.js";
import {
  buildCodexGenerationPrompt,
  buildGenerationPrompt,
  buildJudgeClaudeArgs,
  parseJudgeResults,
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

export interface RunContext {
  readonly withArgs: readonly string[];
  readonly withoutArgs: readonly string[];
  readonly cwd: string | undefined;
  readonly guideDirectory: string;
  readonly harness: "claude" | "codex";
  readonly timeout: number;
  readonly target: string;
  readonly shortName: string;
  readonly judgeModel: string;
  readonly judgeBudget: number;
  readonly iterationDirectory: string;
  readonly runs: number;
  readonly buildPrompt: (evaluation: NormalizedEval) => string;
  // How a harness invocation is carried out. The composition root supplies the
  // spawning implementation; a test supplies recorded output instead.
  readonly runHarness: HarnessRunner;
  // Where the run's evidence is written. A test supplies an in-memory tree.
  readonly fs: FileSystem;
  // When the clock is read, so a recorded run reports a fixed duration.
  readonly now: () => number;
}

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
  harnessArgs: readonly string[],
  expectSkill: boolean,
  context: RunContext,
): Promise<GenerationResult> => {
  const startedAt = context.now();
  const proc = await context.runHarness({
    harness: context.harness,
    args: harnessArgs,
    finalArgument: prompt,
    cwd: context.cwd,
    timeoutMs: context.timeout * 1000,
    // Codex bounds a generation by wall clock alone, so a character ceiling is what
    // stops a runaway answer there. Claude already carries a per-run spend cap and a
    // turn cap, which bound it sooner; applying the ceiling as well only discards an
    // answer for being long, and length is what a thorough skill produces.
    maxOutputCharacters:
      context.harness === "codex" ? GENERATION_OUTPUT_LIMIT : null,
    skillGuide:
      expectSkill && context.harness === "codex"
        ? context.guideDirectory
        : undefined,
  });
  if (proc.outputLimitExceeded) return generationFailure("output-limit");
  if (proc.timedOut)
    return generationFailure("timeout", context.timeout * 1000);
  if (proc.error) return generationFailure(proc.error);

  const parsed =
    context.harness === "claude"
      ? parseClaudeGeneration(proc.stdout, context.target, context.shortName)
      : parseCodexGeneration(proc.stdout, context.now() - startedAt);

  return {
    text: parsed.text,
    totalTokens: sumTokens(parsed.usage),
    durationMs: parsed.durationMs,
    skillTriggered: expectSkill && (parsed.invoked || parsed.available),
    error: parsed.text ? null : "no-result",
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
  evaluation: NormalizedEval,
  response: string,
  context: RunContext,
): Promise<Graded> => {
  const assertions = evaluation.assertions;
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
  const rubric = JUDGE_RUBRIC.replace("{prompt}", evaluation.prompt)
    .replace("{expected}", evaluation.expected_output || "(none provided)")
    .replace("{assertions}", numbered)
    .replace("{response}", response);
  const args =
    context.harness === "claude"
      ? buildJudgeClaudeArgs({
          model: context.judgeModel,
          budget: context.judgeBudget,
          assertionCount: assertions.length,
        })
      : buildIsolatedCodexArgs({model: context.judgeModel});
  const proc = await context.runHarness({
    harness: context.harness,
    args,
    finalArgument: rubric,
    cwd: undefined,
    timeoutMs: 300_000,
    maxOutputCharacters: null,
    skillGuide: undefined,
  });
  if (proc.error) return allFail(`judge process error: ${proc.error}`);
  if (proc.timedOut) return allFail("judge timeout");

  const verdictResults = parseJudgeResults(
    parseJudgeVerdict(context.harness, proc.stdout),
  );
  if (verdictResults === undefined) return allFail("unparseable judge output");
  const results: AssertionResult[] = assertions.map((assertion, index) => ({
    text: assertion,
    passed: verdictResults[index]?.passed ?? false,
    evidence: verdictResults[index]?.evidence ?? "no evidence",
  }));
  return summarize(results);
};

// Where one arm's evidence is written. A multi-run sweep keeps each run apart so a
// later triage reads them individually rather than the last one overwriting the rest.
const armDirectoryPath = (
  evaluation: NormalizedEval,
  arm: EvaluationArm,
  runIndex: number,
  context: RunContext,
): string => {
  const base = join(
    context.iterationDirectory,
    `eval-${String(evaluation.id)}`,
    arm,
  );
  return context.runs > 1 ? join(base, `run-${String(runIndex + 1)}`) : base;
};

export const runJob = async (
  evaluation: NormalizedEval,
  arm: EvaluationArm,
  runIndex: number,
  context: RunContext,
): Promise<JobRecord> => {
  const buildArmPrompt =
    context.harness === "claude"
      ? buildGenerationPrompt
      : buildCodexGenerationPrompt;
  const prompt = buildArmPrompt(
    context.buildPrompt(evaluation),
    arm,
    context.shortName,
  );
  const args = arm === "with_skill" ? context.withArgs : context.withoutArgs;
  const generation = await generate(
    prompt,
    args,
    arm === "with_skill",
    context,
  );
  const healthy =
    generation.error === null &&
    (arm !== "with_skill" || generation.skillTriggered);
  const graded = healthy
    ? await grade(evaluation, generation.text, context)
    : summarize(
        evaluation.assertions.map((assertion) => ({
          text: assertion,
          passed: false,
          evidence: "not graded because generation evidence is invalid",
        })),
      );

  const armDirectory = armDirectoryPath(evaluation, arm, runIndex, context);
  context.fs.makeDirectory(join(armDirectory, "outputs"));
  context.fs.writeFile(
    join(armDirectory, "outputs", "response.md"),
    generation.text,
  );
  context.fs.writeFile(
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
  );
  context.fs.writeFile(
    join(armDirectory, "grading.json"),
    JSON.stringify(graded, null, 2),
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
