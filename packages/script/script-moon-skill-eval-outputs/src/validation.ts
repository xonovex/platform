import {isAbsolute} from "node:path";
import {z} from "zod";

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

const PositiveNumberTextSchema = z
  .string()
  .refine(
    (value) => value.trim().length > 0 && Number.isFinite(Number(value)),
    "must be a finite number",
  )
  .transform(Number)
  .pipe(z.number().positive());

const OutputOptionsSchema = z.object({
  runs: PositiveIntegerTextSchema,
  concurrency: PositiveIntegerTextSchema,
  timeout: PositiveNumberTextSchema,
  budget: PositiveNumberTextSchema.optional(),
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

type ValidationResult<T> =
  | {readonly success: true; readonly data: T}
  | {readonly success: false; readonly error: string};

const errorText = (error: z.ZodError): string =>
  error.issues
    .map((issue) => {
      const path = issue.path.length > 0 ? `${issue.path.join(".")}: ` : "";
      return `${path}${issue.message}`;
    })
    .join("; ");

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

export type {NormalizedEval};
