#!/usr/bin/env node
import {spawn, spawnSync} from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common";

const TOKEN_KEYS = [
  "input_tokens",
  "output_tokens",
  "cache_creation_input_tokens",
  "cache_read_input_tokens",
] as const;

interface EvalInput {
  readonly id?: string | number;
  readonly prompt?: unknown;
  readonly expected_output?: unknown;
  readonly assertions?: unknown;
  readonly files?: unknown;
  readonly [key: string]: unknown;
}

interface NormEval {
  readonly id: string | number;
  readonly prompt: string;
  readonly expected_output: string;
  readonly assertions: readonly string[];
  readonly files: readonly string[];
}

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
}

interface JobRecord {
  readonly id: string | number;
  readonly arm: string;
  readonly pass_rate: number;
  readonly tokens: number;
  readonly duration_ms: number;
  readonly skill_triggered: boolean;
}

interface RunContext {
  readonly with_args: readonly string[];
  readonly without_args: readonly string[];
  readonly cwd: string | undefined;
  readonly timeout: number;
  readonly target: string;
  readonly short: string;
  readonly judge_model: string;
  readonly iter_dir: string;
  readonly runs: number;
  readonly build_prompt: (e: NormEval) => string;
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

const isFile = (path: string): boolean => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

const isDir = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

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
  const probe = process.platform === "win32" ? "where" : "which";
  const result = spawnSync(probe, [cmd], {stdio: "ignore"});
  return result.status === 0;
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
}

const runClaude = (
  args: readonly string[],
  finalArg: string,
  cwd: string | undefined,
  timeoutMs: number,
): Promise<ProcOutput> =>
  new Promise((resolvePromise) => {
    // eslint-disable-next-line sonarjs/no-os-command-from-path
    const child = spawn("claude", [...args, finalArg], {
      stdio: ["ignore", "pipe", "pipe"],
      ...(cwd ? {cwd} : {}),
    });
    let stdout = "";
    let timedOut = false;
    const timer = setTimeout(() => {
      timedOut = true;
      child.kill("SIGKILL");
    }, timeoutMs);
    child.stdout.setEncoding("utf8");
    child.stdout.on("data", (chunk: string) => {
      stdout += chunk;
    });
    child.stderr.on("data", () => {
      // discard; Python captures but does not use generation stderr
    });
    child.on("error", () => {
      clearTimeout(timer);
      resolvePromise({stdout, timedOut});
    });
    child.on("close", () => {
      clearTimeout(timer);
      resolvePromise({stdout, timedOut});
    });
  });

const generate = async (
  prompt: string,
  claudeArgs: readonly string[],
  cwd: string | undefined,
  timeout: number,
  target: string,
  short: string,
): Promise<GenResult> => {
  const proc = await runClaude(claudeArgs, prompt, cwd, timeout * 1000);
  if (proc.timedOut) {
    return {
      text: "",
      total_tokens: 0,
      duration_ms: timeout * 1000,
      skill_triggered: false,
      error: "timeout",
    };
  }

  let text = "";
  let usage: unknown = {};
  let duration = 0;
  let triggered = false;
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
    skill_triggered: triggered,
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

const summarize = (results: readonly AssertionResult[]): Graded => {
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
  };
};

const grade = async (
  prompt: string,
  expected: string,
  assertions: readonly string[],
  response: string,
  model: string,
): Promise<Graded> => {
  const allFail = (reason: string): Graded =>
    summarize(
      assertions.map((a) => ({text: a, passed: false, evidence: reason})),
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

  const args = [
    "-p",
    "--output-format",
    "json",
    "--disallowedTools=Bash,Edit,Write,Read,NotebookEdit,WebFetch," +
      "WebSearch,Glob,Grep,Task,Skill,TodoWrite",
  ];
  if (model) {
    args.push("--model", model);
  }

  let verdict: Record<string, unknown> | null = null;
  const proc = await runClaude(args, rubric, undefined, 300 * 1000);
  if (!proc.timedOut) {
    try {
      const outer: unknown = JSON.parse(proc.stdout);
      const inner =
        isRecord(outer) && typeof outer.result === "string" ? outer.result : "";
      verdict = extractJson(inner);
    } catch {
      verdict = null;
    }
  }

  if (!verdict || !Array.isArray(verdict.assertion_results)) {
    return allFail("unparseable judge output");
  }

  const verdictResults = verdict.assertion_results as unknown[];
  const results: AssertionResult[] = assertions.map((a, i) => {
    const item: unknown = i < verdictResults.length ? verdictResults[i] : {};
    const passed = isRecord(item) ? Boolean(item.passed) : false;
    const evidenceRaw = isRecord(item) ? item.evidence : "";
    const evidence =
      (typeof evidenceRaw === "string" ? evidenceRaw : "") || "no evidence";
    return {text: a, passed, evidence};
  });
  return summarize(results);
};

const runJob = async (
  evalObj: NormEval,
  arm: string,
  runIdx: number,
  ctx: RunContext,
): Promise<JobRecord> => {
  const prompt = ctx.build_prompt(evalObj);
  const args = arm === "with_skill" ? ctx.with_args : ctx.without_args;
  const gen = await generate(
    prompt,
    args,
    ctx.cwd,
    ctx.timeout,
    ctx.target,
    ctx.short,
  );
  const graded = await grade(
    evalObj.prompt,
    evalObj.expected_output,
    evalObj.assertions,
    gen.text,
    ctx.judge_model,
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

// Async concurrency pool over jobs (replaces ThreadPoolExecutor).
const runPool = async <T, R>(
  items: readonly T[],
  limit: number,
  worker: (item: T) => Promise<R>,
): Promise<R[]> => {
  const results: R[] = [];
  let next = 0;
  const runWorker = async (): Promise<void> => {
    while (next < items.length) {
      const idx = next;
      next += 1;
      const item = items[idx];
      if (item === undefined) break;
      results.push(await worker(item));
    }
  };
  const pool = Array.from({length: Math.min(limit, items.length)}, () =>
    runWorker(),
  );
  await Promise.all(pool);
  return results;
};

// Resolve the guide dir for a base dir: base itself if it holds SKILL.md, else
// the single immediate subdir that does (the skill-package layout, e.g.
// skill-c99/c99-guide/SKILL.md, where the simple moon task runs the bin from
// the package root). >1 match is ambiguous; 0 falls back to base so the
// existing evals-not-found error still fires.
const resolveGuideDir = (base: string): string => {
  if (isFile(join(base, "SKILL.md"))) {
    return base;
  }
  const nested = readdirSync(base)
    .map((entry) => join(base, entry))
    .filter((p) => isFile(join(p, "SKILL.md")));
  if (nested.length > 1) {
    process.stderr.write(
      `Error: multiple SKILL.md found under ${base}; pass one explicitly\n`,
    );
    process.exit(2);
  }
  return nested[0] ?? base;
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
          description:
            "runs per arm per eval (env RUNS, default 1; >1 measures variance)",
        },
        concurrency: {
          type: "string",
          description:
            "parallel claude invocations (env CONCURRENCY, default 4)",
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
        "max-budget-usd": {
          type: "string",
          description:
            "optional hard per-generation spend cap (env MAX_BUDGET_USD; unset = no cap)",
        },
      },
    },
    argv,
  );

  // Ergonomic positional defaults. With no explicit positionals the bin runs
  // from the skill package root, so descend into the single guide subdir that
  // holds SKILL.md before defaulting the evals path / reading skill_name.
  const guideDir = resolveGuideDir(resolve("."));
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

  const dataEvals = isRecord(data) ? data.evals : data;
  const evals: unknown[] = Array.isArray(dataEvals) ? dataEvals : [];
  if (evals.length === 0) {
    process.stderr.write(`Error: ${evalsArg} has no evals\n`);
    return 2;
  }

  // Normalize each eval: require id + prompt + (assertions | expected_output).
  const norm: NormEval[] = [];
  for (const [i, raw] of evals.entries()) {
    const e = raw as EvalInput;
    if (!isRecord(e) || !("prompt" in e)) {
      process.stderr.write(`Skipping eval #${String(i)}: missing prompt\n`);
      continue;
    }
    let assertions: string[] = Array.isArray(e.assertions)
      ? (e.assertions as unknown[]).map(String)
      : [];
    const expectedOutput =
      typeof e.expected_output === "string" ? e.expected_output : "";
    if (assertions.length === 0 && expectedOutput) {
      assertions = [expectedOutput];
    }
    if (assertions.length === 0) {
      const idLabel = e.id ?? i;
      process.stderr.write(
        `Skipping eval ${String(idLabel)}: no assertions or expected_output\n`,
      );
      continue;
    }
    const files = Array.isArray(e.files)
      ? (e.files as unknown[]).map(String)
      : [];
    norm.push({
      id: e.id ?? i + 1,
      prompt: String(e.prompt),
      expected_output: expectedOutput,
      assertions,
      files,
    });
  }
  if (norm.length === 0) {
    process.stderr.write("Error: no gradable evals\n");
    return 2;
  }

  const runsRaw =
    (values.runs as string | undefined) ?? process.env.RUNS ?? "1";
  const runs = Number(runsRaw);
  const concurrencyRaw =
    (values.concurrency as string | undefined) ??
    process.env.CONCURRENCY ??
    "4";
  const concurrency = Math.max(1, Number(concurrencyRaw));
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
  const timeout = Number(timeoutRaw);
  const workspaceArg =
    (values.workspace as string | undefined) ?? process.env.WORKSPACE ?? "";
  const cwdArg =
    (values["eval-cwd"] as string | undefined) ?? process.env.EVAL_CWD ?? "";
  const cwd = cwdArg || undefined;
  const budgetRaw =
    (values["max-budget-usd"] as string | undefined) ??
    process.env.MAX_BUDGET_USD ??
    "";
  const budget = budgetRaw ? Number(budgetRaw) : null;

  const short = skillName.split(":").pop() ?? skillName;

  const base = workspaceArg
    ? resolve(workspaceArg)
    : resolve(`${short}-workspace`);
  if (!iteration) {
    let maxExisting = 0;
    if (existsSync(base)) {
      for (const entry of readdirSync(base)) {
        const m = /^iteration-(\d+)$/.exec(entry);
        if (m && isDir(join(base, entry))) {
          maxExisting = Math.max(maxExisting, Number(m[1]));
        }
      }
    }
    iteration = `iteration-${String(maxExisting + 1)}`;
  }
  const iterDir = join(base, iteration);
  mkdirSync(iterDir, {recursive: true});

  const genBase = ["-p", "--output-format", "stream-json", "--verbose"];
  if (claudeModel) {
    genBase.push("--model", claudeModel);
  }
  if (budget && budget > 0) {
    genBase.push("--max-budget-usd", String(budget));
  }
  const withArgs = disallowed
    ? [...genBase, `--disallowedTools=${disallowed}`]
    : [...genBase];
  const withoutDisallowed = [disallowed, "Skill"].filter(Boolean).join(",");
  const withoutArgs = [...genBase, `--disallowedTools=${withoutDisallowed}`];

  const evalsDir = dirname(evalsFile);

  const buildPrompt = (e: NormEval): string => {
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
    iter_dir: iterDir,
    runs,
    build_prompt: buildPrompt,
  };

  process.stderr.write(
    `skill: ${skillName}  evals: ${String(norm.length)}  runs/arm: ${String(runs)}  ` +
      `concurrency: ${String(concurrency)}  workspace: ${iterDir}\n` +
      `gen model: ${claudeModel || "<default>"}  judge model: ${judgeModel || "<default>"}\n---\n`,
  );

  const arms = ["with_skill", "without_skill"] as const;
  interface Job {
    readonly e: NormEval;
    readonly arm: string;
    readonly r: number;
  }
  const jobs: Job[] = [];
  for (const e of norm) {
    for (const arm of arms) {
      for (let r = 0; r < runs; r += 1) {
        jobs.push({e, arm, r});
      }
    }
  }

  const records = await runPool(jobs, concurrency, (job) =>
    runJob(job.e, job.arm, job.r, ctx),
  );

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
  writeFileSync(
    join(iterDir, "benchmark.json"),
    JSON.stringify(benchmark, null, 2),
    {encoding: "utf8"},
  );

  const delta = benchmark.run_summary.delta;
  process.stderr.write(
    `---\nwith_skill pass_rate: ${String(withBlock.pass_rate.mean)}  ` +
      `(skill fired: ${String(withBlock.skill_trigger_rate?.mean)})  ` +
      `tokens: ${String(withBlock.tokens.mean)}\n` +
      `without_skill pass_rate: ${String(withoutBlock.pass_rate.mean)}  ` +
      `tokens: ${String(withoutBlock.tokens.mean)}\n` +
      `delta pass_rate: ${String(delta.pass_rate)}  tokens: ${String(delta.tokens)}\n` +
      `benchmark: ${join(iterDir, "benchmark.json")}\n`,
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
