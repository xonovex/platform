import {z} from "zod";

const QuerySchema = z.object({
  query: z.string().min(1),
  should_trigger: z.boolean(),
  rationale: z.string().default(""),
  split: z.enum(["train", "validation"]).optional(),
});

const QueryListSchema = z
  .array(QuerySchema)
  .min(1, "must contain at least one query");
const QuerySplitSchema = z.enum(["train", "validation", "all"]);

const FiniteNumberTextSchema = z
  .string()
  .refine(
    (value) => value.trim().length > 0 && Number.isFinite(Number(value)),
    "must be a finite number",
  )
  .transform(Number);

const TriggerOptionsSchema = z.object({
  runs: z
    .string()
    .regex(/^\d+$/, "must be a positive integer")
    .transform(Number)
    .pipe(z.int().positive().max(3)),
  threshold: FiniteNumberTextSchema.pipe(z.number().min(0).max(1)),
  budget: FiniteNumberTextSchema.pipe(z.number().positive().max(0.05)),
  batchSize: z
    .string()
    .regex(/^\d+$/, "must be a positive integer")
    .transform(Number)
    .pipe(z.int().positive())
    .optional(),
});

interface TriggerClaudeOptions {
  readonly model: string;
  readonly budget: number;
  readonly pluginDirectories: readonly string[];
}

interface TriggerCodexOptions {
  readonly model: string;
}

type ValidationResult<T> =
  | {readonly success: true; readonly data: T}
  | {readonly success: false; readonly error: string};

export const MAX_TRIGGER_MODEL_RUNS = 24;

export const triggerModelRunCount = (
  queryCount: number,
  runs: number,
): number => queryCount * runs;

const errorText = (error: z.ZodError): string =>
  error.issues
    .map((issue) => {
      const path = issue.path.length > 0 ? `${issue.path.join(".")}: ` : "";
      return `${path}${issue.message}`;
    })
    .join("; ");

export const parseQueries = (
  input: unknown,
): ValidationResult<readonly z.infer<typeof QuerySchema>[]> => {
  const result = QueryListSchema.safeParse(input);
  return result.success
    ? {success: true, data: result.data}
    : {success: false, error: errorText(result.error)};
};

export const parseQuerySplit = (
  input: unknown,
): ValidationResult<z.infer<typeof QuerySplitSchema>> => {
  const result = QuerySplitSchema.safeParse(input);
  return result.success
    ? {success: true, data: result.data}
    : {success: false, error: errorText(result.error)};
};

export const selectQueries = (
  queries: readonly z.infer<typeof QuerySchema>[],
  split: z.infer<typeof QuerySplitSchema>,
): ValidationResult<readonly z.infer<typeof QuerySchema>[]> => {
  const selected =
    split === "all"
      ? queries
      : queries.filter((query) => query.split === split);
  return selected.length > 0
    ? {success: true, data: selected}
    : {success: false, error: `split '${split}' has no queries`};
};

export const parseTriggerOptions = (
  input: z.input<typeof TriggerOptionsSchema>,
): ValidationResult<z.output<typeof TriggerOptionsSchema>> => {
  const result = TriggerOptionsSchema.safeParse(input);
  return result.success
    ? {success: true, data: result.data}
    : {success: false, error: errorText(result.error)};
};

export const buildTriggerClaudeArgs = (
  options: TriggerClaudeOptions,
): readonly string[] => {
  const args = [
    "-p",
    "--output-format",
    "stream-json",
    "--verbose",
    "--include-partial-messages",
    "--setting-sources",
    "",
    "--strict-mcp-config",
    "--mcp-config",
    '{"mcpServers":{}}',
    "--no-session-persistence",
    "--no-chrome",
    "--model",
    options.model,
    "--max-budget-usd",
    String(options.budget),
    "--max-turns",
    "1",
    "--system-prompt",
    "Decide which available skill best matches the user request. " +
      "If one applies, invoke only that Skill immediately. Otherwise reply with one short sentence. " +
      "Do not perform the requested task.",
    "--tools",
    "Skill",
  ];
  for (const pluginDirectory of options.pluginDirectories) {
    args.push("--plugin-dir", pluginDirectory);
  }
  return args;
};

export const buildTriggerCodexArgs = (
  options: TriggerCodexOptions,
): readonly string[] => {
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

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

export const streamTextDeltaLength = (input: unknown): number => {
  if (!isRecord(input) || input.type !== "stream_event") return 0;
  const event = input.event;
  if (!isRecord(event) || event.type !== "content_block_delta") return 0;
  const delta = event.delta;
  if (!isRecord(delta) || delta.type !== "text_delta") return 0;
  return typeof delta.text === "string" ? delta.text.length : 0;
};

export type Query = z.infer<typeof QuerySchema>;
