import {z} from "zod";
import {
  FiniteNumberTextSchema,
  parseWithSchema,
  PositiveIntegerTextSchema,
  type ValidationResult,
} from "./option-schemas.js";

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

const TriggerOptionsSchema = z.object({
  runs: PositiveIntegerTextSchema.pipe(z.number().max(3)),
  threshold: FiniteNumberTextSchema.pipe(z.number().min(0).max(1)),
  budget: FiniteNumberTextSchema.pipe(z.number().positive().max(0.05)),
  batchSize: PositiveIntegerTextSchema.optional(),
});

export const MAX_TRIGGER_MODEL_RUNS = 24;

export const triggerModelRunCount = (
  queryCount: number,
  runs: number,
): number => queryCount * runs;

export const parseQueries = (
  input: unknown,
): ValidationResult<readonly z.infer<typeof QuerySchema>[]> =>
  parseWithSchema(QueryListSchema, input);

export const parseQuerySplit = (
  input: unknown,
): ValidationResult<z.infer<typeof QuerySplitSchema>> =>
  parseWithSchema(QuerySplitSchema, input);

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
): ValidationResult<z.output<typeof TriggerOptionsSchema>> =>
  parseWithSchema(TriggerOptionsSchema, input);

export type Query = z.infer<typeof QuerySchema>;
