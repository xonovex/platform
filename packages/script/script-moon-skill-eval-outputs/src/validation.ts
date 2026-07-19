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

const EvalFileSchema = z.union([
  z.array(z.unknown()),
  z.object({evals: z.array(z.unknown())}),
]);

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

const OutputOptionsSchema = z.object({
  runs: RunsTextSchema,
  concurrency: ConcurrencyTextSchema,
  timeout: PositiveNumberTextSchema,
  budget: GenerationBudgetTextSchema.default(0.1),
  judgeBudget: JudgeBudgetTextSchema.default(0.1),
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

type EvaluationArm = "with_skill" | "without_skill";

interface GenerationClaudeOptions {
  readonly arm: EvaluationArm;
  readonly model: string;
  readonly budget: number;
  readonly disallowedTools: string;
  readonly pluginDirectories?: readonly string[];
}

interface JudgeClaudeOptions {
  readonly model: string;
  readonly budget: number;
  readonly assertionCount: number;
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
    "6",
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
): ValidationResult<readonly unknown[]> => {
  const result = EvalFileSchema.safeParse(input);
  if (!result.success) return {success: false, error: errorText(result.error)};
  return {
    success: true,
    data: Array.isArray(result.data) ? result.data : result.data.evals,
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
};
