import {parse as parseYaml} from "yaml";
import {z} from "zod";

const TaskSchema = z.looseObject({
  env: z.record(z.string(), z.unknown()).optional(),
});

// A key present but empty parses as null, which moon accepts, so neither an
// empty tags list nor an empty tasks map makes a project unreadable.
const ProjectFileSchema = z.looseObject({
  tags: z.array(z.string()).nullish(),
  tasks: z.record(z.string(), TaskSchema).nullish(),
});

export interface ProjectFile {
  // path identifies the project in a failure message, so it is repository
  // relative rather than absolute.
  readonly path: string;
  readonly text: string;
}

// A project carrying either tag inherits ts-coverage from the typescript tag
// templates, which supply a permissive default floor.
const TYPESCRIPT_TAGS: ReadonlySet<string> = new Set([
  "typescript",
  "typescript-script",
]);

const REQUIRED_FLOORS = [
  "TS_COVERAGE_MIN_LINES",
  "TS_COVERAGE_MIN_FUNCTIONS",
  "TS_COVERAGE_MIN_BRANCHES",
  "TS_COVERAGE_MIN_STATEMENTS",
] as const;

// coverageFloorFailures reports a TypeScript project that leaves any coverage
// floor to the tag template. The template's default is a permissive placeholder,
// so a project that does not override it coasts there instead of failing when
// its own coverage drops.
export const coverageFloorFailures = (
  files: readonly ProjectFile[],
): readonly string[] => {
  const failures: string[] = [];
  for (const file of files) {
    let input: unknown;
    try {
      input = parseYaml(file.text);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      failures.push(`${file.path} is not valid YAML: ${message}`);
      continue;
    }
    const result = ProjectFileSchema.safeParse(input);
    if (!result.success) {
      failures.push(`${file.path} has an unreadable project definition`);
      continue;
    }
    const tags = result.data.tags ?? [];
    if (tags.every((tag) => !TYPESCRIPT_TAGS.has(tag))) continue;

    const env = result.data.tasks?.["ts-coverage"]?.env ?? {};
    const missing = REQUIRED_FLOORS.filter(
      (floor) => env[floor] === undefined || env[floor] === "",
    );
    if (missing.length > 0) {
      failures.push(
        `${file.path} task 'ts-coverage' sets no ${missing.join(", ")}: a TypeScript project must declare its own coverage floor rather than inherit the tag template's permissive default`,
      );
    }
  }
  return failures;
};
