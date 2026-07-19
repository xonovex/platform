#!/usr/bin/env node
import {spawn} from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {
  buildGenerationClaudeArgs,
  buildGenerationPrompt,
  buildJudgeClaudeArgs,
  evalEntries,
  findEvaluationInfrastructureFailures,
  MAX_OUTPUT_MODEL_CALLS,
  normalizeEval,
  outputModelCallCount,
  parseJudgeResults,
  parseOutputOptions,
  runFailFastPool,
  streamTextDeltaLength,
  type EvaluationArm,
  type NormalizedEval,
} from "./validation.js";

const TOKEN_KEYS = [
  "input_tokens",
  "output_tokens",
  "cache_creation_input_tokens",
  "cache_read_input_tokens",
] as const;

interface GenResult {
  readonly text: string;
  readonly total_tokens: number;
  readonly duration_ms: number;
  readonly skill_triggered: boolean;
  readonly error: string | null;
}

interface AssertionResult {
  readonly text: string;
  readonly passed: boolean;
  readonly evidence: string;
}

interface GradeSummary {
  readonly passed: number;
  readonly failed: number;
  readonly total: number;
  readonly pass_rate: number;
}

interface Graded {
  readonly assertion_results: readonly AssertionResult[];
  readonly summary: GradeSummary;
  readonly error: string | null;
}

interface JobRecord {
  readonly id: string | number;
  readonly arm: EvaluationArm;
  readonly pass_rate: number;
  readonly tokens: number;
  readonly duration_ms: number;
  readonly skill_triggered: boolean;
  readonly error: string | null;
}

interface RunContext {
  readonly with_args: readonly string[];
  readonly without_args: readonly string[];
  readonly cwd: string | undefined;
  readonly timeout: number;
  readonly target: string;
  readonly short: string;
  readonly judge_model: string;
  readonly judge_budget: number;
  readonly iter_dir: string;
  readonly runs: number;
  readonly build_prompt: (e: NormalizedEval) => string;
}

interface MeanBlock {
  mean: number;
  stddev?: number;
}

interface ArmBlock {
  readonly pass_rate: MeanBlock;
  readonly tokens: MeanBlock;
  readonly duration_ms: MeanBlock;
  skill_trigger_rate?: {readonly mean: number};
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const round = (value: number, digits = 0): number => {
  const factor = 10 ** digits;
  return Math.round(value * factor) / factor;
};

// statistics.fmean: sum / len.
const fmean = (values: readonly number[]): number =>
  values.reduce((acc, v) => acc + v, 0) / values.length;

// statistics.pstdev: population standard deviation, computed manually.
const pstdev = (values: readonly number[]): number => {
  const mean = fmean(values);
  const variance =
    values.reduce((acc, v) => acc + (v - mean) ** 2, 0) / values.length;
  return Math.sqrt(variance);
};

const which = (cmd: string): boolean => {
  try {
    resolveExecutable(cmd);
    return true;
  } catch {
    return false;
  }
};

const matchSkill = (
  skillField: unknown,
  target: string,
  short: string,
): boolean => {
  if (typeof skillField !== "string") return false;
  return (
    skillField === target ||
    skillField === short ||
    skillField.endsWith(":" + short)
  );
};

const skillInObj = (
  obj: Record<string, unknown>,
  target: string,
  short: string,
): boolean => {
  const message = obj.message;
  if (isRecord(message)) {
    const content = Array.isArray(message.content) ? message.content : [];
    for (const item of content) {
      if (isRecord(item) && item.type === "tool_use" && item.name === "Skill") {
        const inp = item.input;
        if (isRecord(inp) && matchSkill(inp.skill, target, short)) {
          return true;
        }
      }
    }
  }
  return false;
};

const skillAvailableInObj = (
  obj: Record<string, unknown>,
  target: string,
  short: string,
): boolean => {
  if (obj.type !== "system" || obj.subtype !== "init") return false;
  const skills = Array.isArray(obj.skills) ? obj.skills : [];
  return skills.some((skill) => matchSkill(skill, target, short));
};

const sumTokens = (usage: unknown): number => {
  if (!isRecord(usage)) return 0;
  let total = 0;
  for (const key of TOKEN_KEYS) {
    const raw = usage[key];
    const num = typeof raw === "number" ? raw : Number(raw);
    total += Number.isFinite(num) ? Math.trunc(num) : 0;
  }
  return total;
};

const extractJson = (text: string): Record<string, unknown> | null => {
  if (!text) return null;
  const fenced = /```(?:json)?\s*(\{[\s\S]*?\})\s*```/.exec(text);
  let candidate: string | null = fenced ? (fenced[1] ?? null) : null;
  if (candidate === null) {
    const start = text.indexOf("{");
    const end = text.lastIndexOf("}");
    candidate = start !== -1 && end > start ? text.slice(start, end + 1) : null;
  }
  if (candidate === null) return null;
  try {
    const obj: unknown = JSON.parse(candidate);
    return isRecord(obj) ? obj : null;
  } catch {
    return null;
  }
};

interface ProcOutput {
  readonly stdout: string;
  readonly timedOut: boolean;
  readonly outputLimitExceeded: boolean;
  readonly error: string | null;
}

const claudeFailureDetail = (stdout: string): string => {
  const lines = stdout.trim().split(/\r?\n/).toReversed();
  for (const line of lines) {
    let value: unknown;
    try {
      value = JSON.parse(line);
    } catch {
      continue;
    }
    if (!isRecord(value)) continue;
    const result = typeof value.result === "string" ? value.result.trim() : "";
    const subtype =
      typeof value.subtype === "string" ? value.subtype.trim() : "";
    if (result && (value.is_error === true || subtype.startsWith("error"))) {
      return result.slice(0, 500);
    }
    if (subtype.startsWith("error")) return subtype.slice(0, 500);
  }
  return "";
};

const runClaude = (
  args: readonly string[],
  finalArg: string,
  cwd: string | undefined,
  timeoutMs: number,
  maxOutputCharacters: number | null = null,
): Promise<ProcOutput> =>
  new Promise((resolvePromise) => {
    const child = spawn(resolveExecutable("claude"), [...args, finalArg], {
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
      const detailSuffix = detail.length > 0 ? `: ${detail}` : "";
      finish(`claude exited ${String(code)}${detailSuffix}`);
    });
  });

const generate = async (
  prompt: string,
  claudeArgs: readonly string[],
  cwd: string | undefined,
  timeout: number,
  target: string,
  short: string,
  expectSkill: boolean,
): Promise<GenResult> => {
  const proc = await runClaude(claudeArgs, prompt, cwd, timeout * 1000, 10_000);
  if (proc.outputLimitExceeded) {
    return {
      text: "",
      total_tokens: 0,
      duration_ms: 0,
      skill_triggered: false,
      error: "output-limit",
    };
  }
  if (proc.timedOut) {
    return {
      text: "",
      total_tokens: 0,
      duration_ms: timeout * 1000,
      skill_triggered: false,
      error: "timeout",
    };
  }
  if (proc.error) {
    return {
      text: "",
      total_tokens: 0,
      duration_ms: 0,
      skill_triggered: false,
      error: proc.error,
    };
  }

  let text = "";
  let usage: unknown = {};
  let duration = 0;
  let triggered = false;
  let available = false;
  for (const rawLine of proc.stdout.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line) continue;
    let obj: unknown;
    try {
      obj = JSON.parse(line);
    } catch {
      continue;
    }
    if (!isRecord(obj)) continue;
    if (skillAvailableInObj(obj, target, short)) {
      available = true;
    }
    if (skillInObj(obj, target, short)) {
      triggered = true;
    }
    if (obj.type === "result") {
      text = typeof obj.result === "string" ? obj.result : "";
      usage = obj.usage ?? {};
      const dur = obj.duration_ms;
      duration = typeof dur === "number" && dur ? dur : 0;
    }
  }

  return {
    text,
    total_tokens: sumTokens(usage),
    duration_ms: duration,
    skill_triggered: expectSkill && (triggered || available),
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

const summarize = (
  results: readonly AssertionResult[],
  error: string | null = null,
): Graded => {
  const passed = results.filter((r) => r.passed).length;
  const total = results.length;
  return {
    assertion_results: results,
    summary: {
      passed,
      failed: total - passed,
      total,
      pass_rate: total > 0 ? round(passed / total, 3) : 0,
    },
    error,
  };
};

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
      assertions.map((a) => ({text: a, passed: false, evidence: reason})),
      reason,
    );

  if (!response.trim()) {
    return allFail("empty response");
  }

  const numbered = assertions
    .map((a, i) => `${String(i + 1)}. ${a}`)
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
  const proc = await runClaude(args, rubric, undefined, 300 * 1000);
  if (proc.error) {
    return allFail(`judge process error: ${proc.error}`);
  }
  if (proc.timedOut) return allFail("judge timeout");
  let verdict: unknown = null;
  try {
    const outer: unknown = JSON.parse(proc.stdout);
    const structured = isRecord(outer) ? outer.structured_output : null;
    if (isRecord(structured)) {
      verdict = structured;
    } else {
      const inner =
        isRecord(outer) && typeof outer.result === "string" ? outer.result : "";
      verdict = extractJson(inner);
    }
  } catch {
    verdict = null;
  }
  const verdictResults = parseJudgeResults(verdict);
  if (verdictResults === undefined) {
    return allFail("unparseable judge output");
  }
  const results: AssertionResult[] = assertions.map((a, i) => {
    const item = verdictResults[i];
    return {
      text: a,
      passed: item?.passed ?? false,
      evidence: item?.evidence ?? "no evidence",
    };
  });
  return summarize(results);
};

const runJob = async (
  evalObj: NormalizedEval,
  arm: EvaluationArm,
  runIdx: number,
  ctx: RunContext,
): Promise<JobRecord> => {
  const prompt = buildGenerationPrompt(
    ctx.build_prompt(evalObj),
    arm,
    ctx.target,
  );
  const args = arm === "with_skill" ? ctx.with_args : ctx.without_args;
  const gen = await generate(
    prompt,
    args,
    ctx.cwd,
    ctx.timeout,
    ctx.target,
    ctx.short,
    arm === "with_skill",
  );
  const generationHealthy =
    gen.error === null && (arm !== "with_skill" || gen.skill_triggered);
  const graded = generationHealthy
    ? await grade(
        evalObj.prompt,
        evalObj.expected_output,
        evalObj.assertions,
        gen.text,
        ctx.judge_model,
        ctx.judge_budget,
      )
    : summarize(
        evalObj.assertions.map((assertion) => ({
          text: assertion,
          passed: false,
          evidence: "not graded because generation evidence is invalid",
        })),
      );

  let armDir = join(ctx.iter_dir, `eval-${String(evalObj.id)}`, arm);
  if (ctx.runs > 1) {
    armDir = join(armDir, `run-${String(runIdx + 1)}`);
  }
  mkdirSync(join(armDir, "outputs"), {recursive: true});
  writeFileSync(join(armDir, "outputs", "response.md"), gen.text, {
    encoding: "utf8",
  });
  writeFileSync(
    join(armDir, "timing.json"),
    JSON.stringify(
      {
        total_tokens: gen.total_tokens,
        duration_ms: gen.duration_ms,
        skill_triggered: gen.skill_triggered,
        error: gen.error,
      },
      null,
      2,
    ),
    {encoding: "utf8"},
  );
  writeFileSync(join(armDir, "grading.json"), JSON.stringify(graded, null, 2), {
    encoding: "utf8",
  });

  const runTag = ctx.runs === 1 ? "" : `/run-${String(runIdx + 1)}`;
  const skillTag =
    arm === "with_skill" && gen.skill_triggered ? " (skill fired)" : "";
  const errTag = gen.error ? ` [${gen.error}]` : "";
  process.stderr.write(
    `  [${String(evalObj.id)}/${arm}${runTag}] ` +
      `pass_rate=${String(graded.summary.pass_rate)} tokens=${String(gen.total_tokens)}` +
      `${skillTag}${errTag}\n`,
  );

  return {
    id: evalObj.id,
    arm,
    pass_rate: graded.summary.pass_rate,
    tokens: gen.total_tokens,
    duration_ms: gen.duration_ms,
    skill_triggered: gen.skill_triggered,
    error: gen.error ?? graded.error,
  };
};

const meanBlock = (values: readonly number[], runs: number): MeanBlock => {
  const block: MeanBlock = {
    mean: values.length > 0 ? round(fmean(values), 3) : 0,
  };
  if (runs > 1 && values.length > 1) {
    block.stddev = round(pstdev(values), 3);
  }
  return block;
};

const aggregateArm = (
  records: readonly JobRecord[],
  arm: string,
  runs: number,
): ArmBlock => {
  const rs = records.filter((r) => r.arm === arm);
  const byEval = new Map<string | number, JobRecord[]>();
  for (const r of rs) {
    const group = byEval.get(r.id) ?? [];
    group.push(r);
    byEval.set(r.id, group);
  }
  const groups = [...byEval.values()];
  const passRates = groups.map((g) => fmean(g.map((r) => r.pass_rate)));
  const tokens = groups.map((g) => fmean(g.map((r) => r.tokens)));
  const durations = groups.map((g) => fmean(g.map((r) => r.duration_ms)));
  const block: ArmBlock = {
    pass_rate: meanBlock(passRates, runs),
    tokens: meanBlock(tokens, runs),
    duration_ms: meanBlock(durations, runs),
  };
  if (arm === "with_skill") {
    const fired = rs.map((r) => (r.skill_triggered ? 1 : 0));
    block.skill_trigger_rate = {
      mean: fired.length > 0 ? round(fmean(fired), 3) : 0,
    };
  }
  return block;
};

// Pull the "name:" frontmatter value out of ./SKILL.md (ergonomic default).
const skillNameFromSkillMd = (path: string): string | undefined => {
  if (!isFile(path)) return undefined;
  const content = readFileSync(path, "utf8");
  const fmMatch = /^---\s*\n([\s\S]*?)\n---/.exec(content);
  const block = fmMatch?.[1] ?? content;
  const nameMatch = /^name:\s*(.+?)\s*$/m.exec(block);
  const name = nameMatch?.[1];
  if (!name) return undefined;
  return name.replaceAll(/^["']|["']$/g, "").trim() || undefined;
};

const main = async (argv: readonly string[]): Promise<number> => {
  const {values, positionals} = parseCliArgs(
    {
      name: "moon-skill-eval-outputs",
      description:
        "Run output-quality evals against a skill: with-skill vs without-skill.",
      options: {
        runs: {
          type: "string",
          description: "runs per arm per eval (env RUNS, default 1, maximum 3)",
        },
        concurrency: {
          type: "string",
          description:
            "parallel claude invocations (env CONCURRENCY, default/maximum 2)",
        },
        "batch-size": {
          type: "string",
          description: "evals per sequential batch",
        },
        model: {
          type: "string",
          description:
            "model for the generation runs (env CLAUDE_MODEL, default haiku)",
        },
        "judge-model": {
          type: "string",
          description: "model for grading (env JUDGE_MODEL)",
        },
        "disallowed-tools": {
          type: "string",
          description:
            "tools blocked in both arms (env DISALLOWED_TOOLS); without-skill also blocks Skill",
        },
        "gen-timeout": {
          type: "string",
          description:
            "per-generation timeout in seconds (env GEN_TIMEOUT, default 600)",
        },
        workspace: {
          type: "string",
          description:
            "workspace base dir (env WORKSPACE, default '<skill>-workspace')",
        },
        "eval-cwd": {
          type: "string",
          description:
            "working dir for generation runs (env EVAL_CWD, default current dir)",
        },
        "plugin-dir": {
          type: "string",
          description:
            "load the with-skill arm from a local plugin directory (env PLUGIN_DIR)",
        },
        "max-budget-usd": {
          type: "string",
          description:
            "hard per-generation spend cap up to 0.10 (env MAX_BUDGET_USD, default 0.10)",
        },
        "judge-max-budget-usd": {
          type: "string",
          description:
            "hard per-judge spend cap up to 0.10 (env JUDGE_MAX_BUDGET_USD, default 0.10)",
        },
      },
    },
    argv,
  );

  // Ergonomic positional defaults. With no explicit positionals the bin runs
  // from the skill package root, so descend into the single guide subdir that
  // holds SKILL.md before defaulting the evals path / reading skill_name.
  const guideDir = resolveGuideDirectory(resolve("."));
  const evalsArg = positionals[0] ?? join(guideDir, "evals.json");
  const evalsFile = resolve(evalsArg);

  if (!isFile(evalsFile)) {
    process.stderr.write(`Error: evals file not found: ${evalsArg}\n`);
    return 2;
  }
  if (!which("claude")) {
    process.stderr.write("Error: 'claude' CLI not found in PATH\n");
    return 2;
  }

  // skill_name defaults to the "name" frontmatter from the guide dir's SKILL.md.
  const skillArg = positionals[1];
  const skillMd = join(guideDir, "SKILL.md");
  const skillName =
    skillArg && skillArg.length > 0 ? skillArg : skillNameFromSkillMd(skillMd);
  if (!skillName) {
    process.stderr.write(
      `Error: no skill_name given and no 'name' frontmatter in ${skillMd}\n`,
    );
    return 2;
  }

  let iteration = positionals[2] ?? "";

  let data: unknown;
  try {
    data = JSON.parse(readFileSync(evalsFile, "utf8"));
  } catch (error) {
    const msg = error instanceof Error ? error.message : String(error);
    process.stderr.write(`Error: invalid JSON in ${evalsArg}: ${msg}\n`);
    return 2;
  }

  const entriesResult = evalEntries(data);
  if (!entriesResult.success) {
    process.stderr.write(
      `Error: invalid eval structure in ${evalsArg}: ${entriesResult.error}\n`,
    );
    return 2;
  }
  const evals = entriesResult.data;
  if (evals.length === 0) {
    process.stderr.write(`Error: ${evalsArg} has no evals\n`);
    return 2;
  }

  // Normalize each eval: require id + prompt + (assertions | expected_output).
  const norm: NormalizedEval[] = [];
  for (const [i, raw] of evals.entries()) {
    const result = normalizeEval(raw, i + 1);
    if (!result.success) {
      process.stderr.write(
        `Skipping eval #${String(i + 1)}: ${result.error}\n`,
      );
      continue;
    }
    norm.push(result.data);
  }
  if (norm.length === 0) {
    process.stderr.write("Error: no gradable evals\n");
    return 2;
  }

  const runsRaw =
    (values.runs as string | undefined) ?? process.env.RUNS ?? "1";
  const concurrencyRaw =
    (values.concurrency as string | undefined) ??
    process.env.CONCURRENCY ??
    "2";
  const claudeModel =
    (values.model as string | undefined) ?? process.env.CLAUDE_MODEL ?? "haiku";
  const judgeModel =
    (values["judge-model"] as string | undefined) ??
    process.env.JUDGE_MODEL ??
    "";
  const disallowed =
    (values["disallowed-tools"] as string | undefined) ??
    process.env.DISALLOWED_TOOLS ??
    "Bash,Edit,Write,NotebookEdit,WebFetch";
  const timeoutRaw =
    (values["gen-timeout"] as string | undefined) ??
    process.env.GEN_TIMEOUT ??
    "600";
  const workspaceArg =
    (values.workspace as string | undefined) ?? process.env.WORKSPACE ?? "";
  const cwdArg =
    (values["eval-cwd"] as string | undefined) ?? process.env.EVAL_CWD ?? "";
  const cwd = cwdArg || undefined;
  const pluginDirRaw =
    (values["plugin-dir"] as string | undefined) ??
    process.env.PLUGIN_DIR ??
    "";
  const pluginDir = pluginDirRaw ? resolve(pluginDirRaw) : undefined;
  if (pluginDir !== undefined && !isDirectory(pluginDir)) {
    process.stderr.write(
      `Error: local plugin directory not found: ${pluginDirRaw}\n`,
    );
    return 2;
  }
  const budgetRaw =
    (values["max-budget-usd"] as string | undefined) ??
    process.env.MAX_BUDGET_USD ??
    "0.10";
  const judgeBudgetRaw =
    (values["judge-max-budget-usd"] as string | undefined) ??
    process.env.JUDGE_MAX_BUDGET_USD ??
    "0.10";
  const batchSizeRaw = values["batch-size"] as string | undefined;
  const optionsResult = parseOutputOptions({
    runs: runsRaw,
    concurrency: concurrencyRaw,
    timeout: timeoutRaw,
    budget: budgetRaw,
    judgeBudget: judgeBudgetRaw,
    ...(batchSizeRaw === undefined ? {} : {batchSize: batchSizeRaw}),
    ...(iteration ? {iteration} : {}),
  });
  if (!optionsResult.success) {
    process.stderr.write(
      `Error: invalid evaluator options: ${optionsResult.error}\n`,
    );
    return 2;
  }
  const {runs, concurrency, timeout, batchSize} = optionsResult.data;
  const budget = optionsResult.data.budget;
  const judgeBudget = optionsResult.data.judgeBudget;
  iteration = optionsResult.data.iteration ?? "";

  const evalBatches = boundedBatches(
    norm,
    batchSize ?? Math.max(norm.length, 1),
  );
  const maxBatchModelCalls = Math.max(
    0,
    ...evalBatches.map((batch) => outputModelCallCount(batch.length, runs)),
  );
  if (maxBatchModelCalls > MAX_OUTPUT_MODEL_CALLS) {
    process.stderr.write(
      `Error: output eval batch would launch ${String(maxBatchModelCalls)} model calls; ` +
        `maximum is ${String(MAX_OUTPUT_MODEL_CALLS)}. Split the eval set into bounded batches.\n`,
    );
    return 2;
  }

  const short = skillName.split(":").pop() ?? skillName;

  const base = workspaceArg
    ? resolve(workspaceArg)
    : resolve(`${short}-workspace`);
  if (!iteration) {
    let maxExisting = 0;
    if (existsSync(base)) {
      for (const entry of readdirSync(base)) {
        const m = /^iteration-(\d+)$/.exec(entry);
        if (m && isDirectory(join(base, entry))) {
          maxExisting = Math.max(maxExisting, Number(m[1]));
        }
      }
    }
    iteration = `iteration-${String(maxExisting + 1)}`;
  }
  const iterDir = join(base, iteration);
  mkdirSync(iterDir, {recursive: true});
  const benchmarkPath = join(iterDir, "benchmark.json");
  const invalidRunPath = join(iterDir, "invalid-run.json");
  rmSync(benchmarkPath, {force: true});
  rmSync(invalidRunPath, {force: true});

  const withArgs = buildGenerationClaudeArgs({
    arm: "with_skill",
    model: claudeModel,
    budget,
    disallowedTools: disallowed,
    ...(pluginDir
      ? {pluginDirectories: resolveClaudePluginDirectories(pluginDir)}
      : {}),
  });
  const withoutArgs = buildGenerationClaudeArgs({
    arm: "without_skill",
    model: claudeModel,
    budget,
    disallowedTools: disallowed,
  });

  const evalsDir = dirname(evalsFile);

  const buildPrompt = (e: NormalizedEval): string => {
    let prompt = e.prompt;
    const files = e.files;
    if (files.length > 0) {
      const paths = files.map((f) => resolve(evalsDir, f));
      prompt +=
        "\n\nRelevant input files (read them as needed):\n" +
        paths.map((p) => `- ${p}`).join("\n");
    }
    return prompt;
  };

  const ctx: RunContext = {
    with_args: withArgs,
    without_args: withoutArgs,
    cwd,
    timeout,
    target: skillName,
    short,
    judge_model: judgeModel,
    judge_budget: judgeBudget,
    iter_dir: iterDir,
    runs,
    build_prompt: buildPrompt,
  };

  process.stderr.write(
    `skill: ${skillName}  evals: ${String(norm.length)}  runs/arm: ${String(runs)}  ` +
      `concurrency: ${String(concurrency)}  workspace: ${iterDir}\n` +
      `gen model: ${claudeModel || "<default>"}  judge model: ${judgeModel || "<default>"}\n` +
      `caps: generation=$${String(budget)}/6 turns  judge=$${String(judgeBudget)}/1 turn  ` +
      `model_calls=${String(outputModelCallCount(norm.length, runs))}  ` +
      `batches=${String(evalBatches.length)}  ` +
      `max_batch_calls=${String(maxBatchModelCalls)}/${String(MAX_OUTPUT_MODEL_CALLS)}  retries=0\n---\n`,
  );

  const arms = ["with_skill", "without_skill"] as const;
  interface Job {
    readonly e: NormalizedEval;
    readonly arm: EvaluationArm;
    readonly r: number;
  }
  const records: JobRecord[] = [];
  for (const [batchIndex, batch] of evalBatches.entries()) {
    process.stderr.write(
      `batch ${String(batchIndex + 1)}/${String(evalBatches.length)}: ` +
        `${String(batch.length)} evals\n`,
    );
    const jobs: Job[] = [];
    for (const e of batch) {
      for (const arm of arms) {
        for (let r = 0; r < runs; r += 1) {
          jobs.push({e, arm, r});
        }
      }
    }

    const batchRecords = await runFailFastPool(
      jobs,
      concurrency,
      (job) => runJob(job.e, job.arm, job.r, ctx),
      (record) => findEvaluationInfrastructureFailures([record]).length > 0,
    );
    records.push(...batchRecords);

    const infrastructureFailures =
      findEvaluationInfrastructureFailures(batchRecords);
    if (infrastructureFailures.length > 0) {
      writeFileSync(
        invalidRunPath,
        JSON.stringify(
          {
            skill: skillName,
            iteration,
            status: "invalid",
            batch: batchIndex + 1,
            failures: infrastructureFailures,
          },
          null,
          2,
        ),
        {encoding: "utf8"},
      );
      process.stderr.write(
        `invalid benchmark evidence: ${String(infrastructureFailures.length)} infrastructure failure(s)\n` +
          `diagnostic: ${invalidRunPath}\n`,
      );
      return 2;
    }
  }

  // Per-eval stdout lines.
  for (const e of norm) {
    const w = records.filter((r) => r.id === e.id && r.arm === "with_skill");
    const wo = records.filter(
      (r) => r.id === e.id && r.arm === "without_skill",
    );
    const wPr = w.length > 0 ? fmean(w.map((r) => r.pass_rate)) : 0;
    const woPr = wo.length > 0 ? fmean(wo.map((r) => r.pass_rate)) : 0;
    const wTok = w.length > 0 ? fmean(w.map((r) => r.tokens)) : 0;
    const woTok = wo.length > 0 ? fmean(wo.map((r) => r.tokens)) : 0;
    process.stdout.write(
      JSON.stringify({
        id: e.id,
        prompt: e.prompt,
        with_skill: {
          pass_rate: round(wPr, 3),
          tokens: round(wTok),
          skill_triggered: w.some((r) => r.skill_triggered),
        },
        without_skill: {pass_rate: round(woPr, 3), tokens: round(woTok)},
        delta_pass_rate: round(wPr - woPr, 3),
        delta_tokens: round(wTok - woTok),
      }) + "\n",
    );
  }

  const withBlock = aggregateArm(records, "with_skill", runs);
  const withoutBlock = aggregateArm(records, "without_skill", runs);
  const benchmark = {
    skill: skillName,
    iteration,
    runs_per_arm: runs,
    eval_count: norm.length,
    batch_count: evalBatches.length,
    run_summary: {
      with_skill: withBlock,
      without_skill: withoutBlock,
      delta: {
        pass_rate: round(
          withBlock.pass_rate.mean - withoutBlock.pass_rate.mean,
          3,
        ),
        tokens: round(withBlock.tokens.mean - withoutBlock.tokens.mean),
        duration_ms: round(
          withBlock.duration_ms.mean - withoutBlock.duration_ms.mean,
        ),
      },
    },
  };
  writeFileSync(benchmarkPath, JSON.stringify(benchmark, null, 2), {
    encoding: "utf8",
  });

  const delta = benchmark.run_summary.delta;
  process.stderr.write(
    `---\nwith_skill pass_rate: ${String(withBlock.pass_rate.mean)}  ` +
      `(skill fired: ${String(withBlock.skill_trigger_rate?.mean)})  ` +
      `tokens: ${String(withBlock.tokens.mean)}\n` +
      `without_skill pass_rate: ${String(withoutBlock.pass_rate.mean)}  ` +
      `tokens: ${String(withoutBlock.tokens.mean)}\n` +
      `delta pass_rate: ${String(delta.pass_rate)}  tokens: ${String(delta.tokens)}\n` +
      `benchmark: ${benchmarkPath}\n`,
  );
  return delta.pass_rate > 0 ? 0 : 1;
};

try {
  const code = await main(process.argv.slice(2));
  process.exit(code);
} catch (error) {
  const msg = error instanceof Error ? error.message : String(error);
  process.stderr.write(`Error: ${msg}\n`);
  process.exit(2);
}
