import {isRecord} from "@xonovex/script-moon-common/records";
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

interface IsolatedCodexOptions {
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

// One turn to choose a skill, and a second only for the run that needs it: a model
// that writes a skill name the harness cannot resolve launches nothing, and the turn
// it spent says nothing about routing. The second turn lets it answer the harness's
// error with a name that exists. A run that picks a skill on its first turn ends
// there, so the extra turn costs nothing on the runs that get it right.
export const TRIGGER_MAX_TURNS = 2;

// buildIsolatedClaudeArgs carries the flags that keep an eval run hermetic: no
// settings from the developer's machine, no MCP servers, no persisted session, no
// browser. A run that inherits any of them scores the local environment instead of
// the skill, so every Claude eval invocation starts from this set.
export const buildIsolatedClaudeArgs = (
  outputFormat: "json" | "stream-json",
): readonly string[] => [
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
];

// buildIsolatedCodexArgs is the Codex counterpart of buildIsolatedClaudeArgs:
// an ephemeral read-only exec that reads neither user config nor rules.
export const buildIsolatedCodexArgs = (
  options: IsolatedCodexOptions,
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

export const buildTriggerClaudeArgs = (
  options: TriggerClaudeOptions,
): readonly string[] => {
  const args = [
    ...buildIsolatedClaudeArgs("stream-json"),
    "--verbose",
    "--include-partial-messages",
    "--model",
    options.model,
    "--max-budget-usd",
    String(options.budget),
    "--max-turns",
    String(TRIGGER_MAX_TURNS),
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

export const streamTextDeltaLength = (input: unknown): number => {
  if (!isRecord(input) || input.type !== "stream_event") return 0;
  const event = input.event;
  if (!isRecord(event) || event.type !== "content_block_delta") return 0;
  const delta = event.delta;
  if (!isRecord(delta) || delta.type !== "text_delta") return 0;
  return typeof delta.text === "string" ? delta.text.length : 0;
};

export type Query = z.infer<typeof QuerySchema>;
