import type {z} from "zod";
import type {FileRead} from "./release-snapshot.js";
import type {Checker} from "./validate.js";

const issueDetail = (issue: z.core.$ZodIssue): string => {
  const path =
    issue.path.length > 0 ? issue.path.map(String).join(".") : "root";
  return `${path}: ${issue.message}`;
};

const parseJson = (
  read: FileRead,
): {readonly value: unknown} | {readonly error: string} => {
  if (read.kind === "unreadable") return {error: read.reason};
  try {
    return {value: JSON.parse(read.text)};
  } catch (error: unknown) {
    return {error: error instanceof Error ? error.message : String(error)};
  }
};

/**
 * The file's JSON value when it parses and matches the schema. Otherwise each
 * problem is checked as a failure under label and the result is undefined.
 */
export const jsonInput = <Output>(
  read: FileRead,
  label: string,
  schema: z.ZodType<Output>,
  check: Checker["check"],
): Output | undefined => {
  const parsed = parseJson(read);
  if ("error" in parsed) {
    check(false, `${label} could not be read or parsed: ${parsed.error}`);
    return undefined;
  }
  const result = schema.safeParse(parsed.value);
  if (!result.success) {
    for (const issue of result.error.issues) {
      check(false, `${label} is invalid: ${issueDetail(issue)}`);
    }
    return undefined;
  }
  return result.data;
};

/** The file's text, or undefined after a failure is checked under label. */
export const textInput = (
  read: FileRead,
  label: string,
  check: Checker["check"],
): string | undefined => {
  if (read.kind === "text") return read.text;
  check(false, `${label} could not be read: ${read.reason}`);
  return undefined;
};
