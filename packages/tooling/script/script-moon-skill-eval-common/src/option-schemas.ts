import {z} from "zod";

export type ValidationResult<T> =
  | {readonly success: true; readonly data: T}
  | {readonly success: false; readonly error: string};

/** An option given as text that must name a whole number above zero. */
export const PositiveIntegerTextSchema = z
  .string()
  .regex(/^\d+$/, "must be a positive integer")
  .transform(Number)
  .pipe(z.int().positive());

/** An option given as text that must name a finite number. */
export const FiniteNumberTextSchema = z
  .string()
  .refine(
    (value) => value.trim().length > 0 && Number.isFinite(Number(value)),
    "must be a finite number",
  )
  .transform(Number);

const errorText = (error: z.ZodError): string =>
  error.issues
    .map((issue) => {
      const path = issue.path.length > 0 ? `${issue.path.join(".")}: ` : "";
      return `${path}${issue.message}`;
    })
    .join("; ");

/** Parses input with a schema and states every issue in one line. */
export const parseWithSchema = <S extends z.ZodType>(
  schema: S,
  input: unknown,
): ValidationResult<z.output<S>> => {
  const result = schema.safeParse(input);
  return result.success
    ? {success: true, data: result.data}
    : {success: false, error: errorText(result.error)};
};
