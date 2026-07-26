import {isAbsolute} from "node:path";
import {z} from "zod";

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const SafePathSegmentSchema = z
  .string()
  .min(1)
  .regex(/^[\w.-]+$/, "must be a safe path segment");

const EvalIdSchema = z.union([z.int().nonnegative(), SafePathSegmentSchema]);

const RelativeFileSchema = z
  .string()
  .min(1)
  .refine(
    (path) =>
      !isAbsolute(path) && path.split(/[\\/]/).every((part) => part !== ".."),
    "must stay within the eval directory",
  );

const EvalInputSchema = z.object({
  id: EvalIdSchema.optional(),
  prompt: z.string().min(1),
  expected_output: z.string().default(""),
  assertions: z.array(z.string().min(1)).default([]),
  files: z.array(RelativeFileSchema).default([]),
});

const OutputTierSchema = z.enum(["aggressive", "moderate", "conservative"]);

const EvalFileSchema = z.object({
  skill_name: z.string().min(1),
  tier: OutputTierSchema,
  evals: z.array(z.unknown()),
});

const JudgeResultSchema = z.object({
  passed: z.boolean(),
  evidence: z.string().optional(),
});

const JudgeVerdictSchema = z.object({
  assertion_results: z.array(JudgeResultSchema),
});

const PositiveIntegerTextSchema = z
  .string()
  .regex(/^\d+$/, "must be a positive integer")
  .transform(Number)
  .pipe(z.int().positive());

const RunsTextSchema = PositiveIntegerTextSchema.pipe(z.number().max(3));
const ConcurrencyTextSchema = PositiveIntegerTextSchema.pipe(z.number().max(2));

const PositiveNumberTextSchema = z
  .string()
  .refine(
    (value) => value.trim().length > 0 && Number.isFinite(Number(value)),
    "must be a finite number",
  )
  .transform(Number)
  .pipe(z.number().positive());

const GenerationBudgetTextSchema = PositiveNumberTextSchema.pipe(
  z.number().max(0.1),
);

const JudgeBudgetTextSchema = PositiveNumberTextSchema.pipe(
  z.number().max(0.1),
);

// A with-skill run spends turns loading the skill and its references before it
// answers, so a guide with progressive disclosure needs more headroom than a bare
// answer. Too low a ceiling reports error_max_turns, which invalidates the batch
// rather than scoring the arm; the cap stays configurable up to twice the default
// for skills whose references need extra reads.
export const GENERATION_MAX_TURNS = 12;

const MaxTurnsTextSchema = PositiveIntegerTextSchema.pipe(
  z.number().max(GENERATION_MAX_TURNS * 2),
);

const OutputOptionsSchema = z.object({
  runs: RunsTextSchema,
  concurrency: ConcurrencyTextSchema,
  timeout: PositiveNumberTextSchema,
  budget: GenerationBudgetTextSchema.default(0.1),
  judgeBudget: JudgeBudgetTextSchema.default(0.1),
  maxTurns: MaxTurnsTextSchema.default(GENERATION_MAX_TURNS),
  batchSize: z
    .string()
    .regex(/^\d+$/, "must be a positive integer")
    .transform(Number)
    .pipe(z.int().positive())
    .optional(),
  iteration: z
    .string()
    .regex(/^iteration-[1-9]\d*$/, "must match iteration-N")
    .optional(),
});

interface NormalizedEval {
  readonly id: string | number;
  readonly prompt: string;
  readonly expected_output: string;
  readonly assertions: readonly string[];
  readonly files: readonly string[];
}

type OutputTier = z.infer<typeof OutputTierSchema>;

interface ParsedEvalFile {
  readonly skillName: string;
  readonly tier: OutputTier;
  readonly evals: readonly unknown[];
}

interface OutputGatePolicy {
  readonly minimumWithSkillPassRate: number;
  readonly minimumDeltaPassRate: number;
}

interface OutputGateResult {
  readonly passed: boolean;
  readonly policy: OutputGatePolicy;
  readonly checks: {
    readonly withSkillPassRate: boolean;
    readonly deltaPassRate: boolean;
    readonly skillTriggerRate: boolean;
  };
}

type EvaluationArm = "with_skill" | "without_skill";

interface GenerationClaudeOptions {
  readonly arm: EvaluationArm;
  readonly model: string;
  readonly budget: number;
  readonly maxTurns: number;
  readonly disallowedTools: string;
  readonly pluginDirectories?: readonly string[];
}

interface JudgeClaudeOptions {
  readonly model: string;
  readonly budget: number;
  readonly assertionCount: number;
}

interface CodexOptions {
  readonly model: string;
}

interface EvaluationHealthRecord {
  readonly id: string | number;
  readonly arm: EvaluationArm;
  readonly skill_triggered: boolean;
  readonly error: string | null;
}

interface EvaluationInfrastructureFailure {
  readonly id: string | number;
  readonly arm: EvaluationArm;
  readonly reason: string;
}

type ValidationResult<T> =
  | {readonly success: true; readonly data: T}
  | {readonly success: false; readonly error: string};

export const MAX_OUTPUT_MODEL_CALLS = 24;

const OUTPUT_GATE_POLICIES: Readonly<Record<OutputTier, OutputGatePolicy>> = {
  aggressive: {
    minimumWithSkillPassRate: 0.75,
    minimumDeltaPassRate: 0.05,
  },
  moderate: {
    minimumWithSkillPassRate: 0.8,
    minimumDeltaPassRate: 0.05,
  },
  conservative: {
    minimumWithSkillPassRate: 0.9,
    minimumDeltaPassRate: 0.1,
  },
};

export const outputModelCallCount = (evalCount: number, runs: number): number =>
  evalCount * runs * 4;

const errorText = (error: z.ZodError): string =>
  error.issues
    .map((issue) => {
      const path = issue.path.length > 0 ? `${issue.path.join(".")}: ` : "";
      return `${path}${issue.message}`;
    })
    .join("; ");

export const buildGenerationPrompt = (
  prompt: string,
  arm: EvaluationArm,
  skillName: string,
): string => (arm === "with_skill" ? `/${skillName} ${prompt}` : prompt);

export const buildCodexGenerationPrompt = (
  prompt: string,
  arm: EvaluationArm,
  skillName: string,
): string => (arm === "with_skill" ? `$${skillName}\n\n${prompt}` : prompt);

const isolatedClaudeArgs = (outputFormat: "json" | "stream-json") =>
  [
    "-p",
    "--output-format",
    outputFormat,
    "--setting-sources",
    "",
    "--strict-mcp-config",
    "--mcp-config",
    '{"mcpServers":{}}',
    "--no-session-persistence",
    "--no-chrome",
  ] as const;

const GENERATION_SYSTEM_PROMPT =
  "Answer the user request directly. Use the explicitly invoked skill as " +
  "authoritative guidance. Read only files that the skill itself identifies " +
  "as necessary. Keep the final response under 1,000 words.";

const JUDGE_SYSTEM_PROMPT =
  "Grade only the supplied response against the supplied assertions and " +
  "return exactly the requested JSON.";

export const buildGenerationClaudeArgs = (
  options: GenerationClaudeOptions,
): readonly string[] => {
  const args: string[] = [
    ...isolatedClaudeArgs("stream-json"),
    "--verbose",
    "--include-partial-messages",
    "--model",
    options.model,
    "--max-budget-usd",
    String(options.budget),
    "--max-turns",
    String(options.maxTurns),
    "--system-prompt",
    GENERATION_SYSTEM_PROMPT,
  ];
  if (options.arm === "with_skill") {
    args.push("--tools", "Skill,Read");
    if (options.disallowedTools) {
      args.push(`--disallowedTools=${options.disallowedTools}`);
    }
    for (const pluginDirectory of options.pluginDirectories ?? []) {
      args.push("--plugin-dir", pluginDirectory);
    }
    return args;
  }
  const blocked = ["Skill", options.disallowedTools].filter(Boolean).join(",");
  args.push("--tools", "Read", `--disallowedTools=${blocked}`);
  return args;
};

export const buildJudgeClaudeArgs = (
  options: JudgeClaudeOptions,
): readonly string[] => {
  const schema = JSON.stringify({
    type: "object",
    properties: {
      assertion_results: {
        type: "array",
        minItems: options.assertionCount,
        maxItems: options.assertionCount,
        items: {
          type: "object",
          properties: {
            text: {type: "string"},
            passed: {type: "boolean"},
            evidence: {type: "string"},
          },
          required: ["text", "passed", "evidence"],
          additionalProperties: false,
        },
      },
    },
    required: ["assertion_results"],
    additionalProperties: false,
  });
  const args: string[] = [
    ...isolatedClaudeArgs("json"),
    "--tools",
    "",
    "--max-budget-usd",
    String(options.budget),
    "--max-turns",
    "1",
    "--system-prompt",
    JUDGE_SYSTEM_PROMPT,
    "--json-schema",
    schema,
  ];
  if (options.model) args.push("--model", options.model);
  return args;
};

export const buildCodexArgs = (options: CodexOptions): readonly string[] => {
  const args = [
    "exec",
    "--json",
    "--ephemeral",
    "--sandbox",
    "read-only",
    "--ignore-user-config",
    "--ignore-rules",
    "--skip-git-repo-check",
  ];
  if (options.model.length > 0) args.push("--model", options.model);
  return args;
};

export const findEvaluationInfrastructureFailures = (
  records: readonly EvaluationHealthRecord[],
): readonly EvaluationInfrastructureFailure[] =>
  records.flatMap((record) => {
    if (record.error !== null) {
      return [{id: record.id, arm: record.arm, reason: record.error}];
    }
    if (record.arm === "with_skill" && !record.skill_triggered) {
      return [
        {
          id: record.id,
          arm: record.arm,
          reason: "target skill did not activate",
        },
      ];
    }
    return [];
  });

export const streamTextDeltaLength = (input: unknown): number => {
  if (!isRecord(input) || input.type !== "stream_event") return 0;
  const event = input.event;
  if (!isRecord(event) || event.type !== "content_block_delta") return 0;
  const delta = event.delta;
  if (!isRecord(delta) || delta.type !== "text_delta") return 0;
  return typeof delta.text === "string" ? delta.text.length : 0;
};

// A harness error is usually transient (a dropped stream, an output-limit trip, an
// exhausted turn cap), so retry the job before discarding the whole batch's
// evidence. A with-skill run whose target skill never activated carries no error
// and is not retried: that is a finding, not a flake. A job that fails every
// attempt still invalidates: partial evidence must never be reported as a pass.
export const OUTPUT_RUN_ATTEMPTS = 3;

export const runWithTransientRetry = async <
  R extends {readonly error: string | null},
>(
  run: () => Promise<R>,
  attempts: number,
  onRetry: (attempt: number, error: string) => void,
): Promise<R> => {
  let result = await run();
  for (
    let attempt = 1;
    attempt < attempts && result.error !== null;
    attempt += 1
  ) {
    onRetry(attempt, result.error);
    result = await run();
  }
  return result;
};

export const runFailFastPool = async <T, R>(
  items: readonly T[],
  limit: number,
  worker: (item: T) => Promise<R>,
  shouldStop: (result: R) => boolean,
): Promise<R[]> => {
  const results: R[] = [];
  let next = 0;
  let stopped = false;
  const runWorker = async (): Promise<void> => {
    while (!stopped && next < items.length) {
      const idx = next;
      next += 1;
      const item = items[idx];
      if (item === undefined) break;
      const result = await worker(item);
      results.push(result);
      if (shouldStop(result)) stopped = true;
    }
  };
  const pool = Array.from({length: Math.min(limit, items.length)}, () =>
    runWorker(),
  );
  await Promise.all(pool);
  return results;
};

export const evalEntries = (
  input: unknown,
): ValidationResult<ParsedEvalFile> => {
  const result = EvalFileSchema.safeParse(input);
  if (!result.success) return {success: false, error: errorText(result.error)};
  return {
    success: true,
    data: {
      skillName: result.data.skill_name,
      tier: result.data.tier,
      evals: result.data.evals,
    },
  };
};

export const evaluateOutputGate = (
  tier: OutputTier,
  withSkillPassRate: number,
  withoutSkillPassRate: number,
  skillTriggerRate: number,
): OutputGateResult => {
  const policy = OUTPUT_GATE_POLICIES[tier];
  const checks = {
    withSkillPassRate: withSkillPassRate >= policy.minimumWithSkillPassRate,
    deltaPassRate:
      withSkillPassRate - withoutSkillPassRate >= policy.minimumDeltaPassRate,
    skillTriggerRate: skillTriggerRate === 1,
  };
  return {
    passed: Object.values(checks).every(Boolean),
    policy,
    checks,
  };
};

export const normalizeEval = (
  input: unknown,
  fallbackId: number,
): ValidationResult<NormalizedEval> => {
  const result = EvalInputSchema.safeParse(input);
  if (!result.success) return {success: false, error: errorText(result.error)};
  let assertions: readonly string[] = result.data.assertions;
  if (assertions.length === 0 && result.data.expected_output.length > 0) {
    assertions = [result.data.expected_output];
  }
  if (assertions.length === 0) {
    return {success: false, error: "no assertions or expected_output"};
  }
  return {
    success: true,
    data: {
      id: result.data.id ?? fallbackId,
      prompt: result.data.prompt,
      expected_output: result.data.expected_output,
      assertions,
      files: result.data.files,
    },
  };
};

export const validateUniqueEvaluationIds = (
  evaluations: readonly NormalizedEval[],
): ValidationResult<readonly NormalizedEval[]> => {
  const seen = new Set<string>();
  for (const evaluation of evaluations) {
    const id = String(evaluation.id);
    if (seen.has(id)) {
      return {success: false, error: `duplicate eval id: ${id}`};
    }
    seen.add(id);
  }
  return {success: true, data: evaluations};
};

export const parseJudgeResults = (
  input: unknown,
): readonly z.infer<typeof JudgeResultSchema>[] | undefined => {
  const result = JudgeVerdictSchema.safeParse(input);
  return result.success ? result.data.assertion_results : undefined;
};

export const parseOutputOptions = (
  input: z.input<typeof OutputOptionsSchema>,
): ValidationResult<z.output<typeof OutputOptionsSchema>> => {
  const result = OutputOptionsSchema.safeParse(input);
  return result.success
    ? {success: true, data: result.data}
    : {success: false, error: errorText(result.error)};
};

export type {
  EvaluationArm,
  EvaluationHealthRecord,
  EvaluationInfrastructureFailure,
  NormalizedEval,
  OutputGateResult,
  OutputTier,
  ParsedEvalFile,
};
