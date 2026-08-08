import {parse as parseYaml} from "yaml";
import {z} from "zod";

// A key present but empty parses as null, which moon accepts, so neither an
// empty dependsOn list nor an empty fileGroups map makes a project unreadable.
const ProjectFileSchema = z.looseObject({
  tags: z.array(z.string()).nullish(),
  dependsOn: z
    .array(z.union([z.string(), z.looseObject({id: z.string()})]))
    .nullish(),
  fileGroups: z.record(z.string(), z.array(z.string()).nullish()).nullish(),
});

export interface ProjectFile {
  /** Repository-relative path to the project's moon.yml, minus the file name. */
  readonly source: string;
  readonly id: string;
  readonly text: string;
}

// A project carrying any of these inherits the typescript task templates, whose
// tasks read the source of the packages it depends on.
const TYPESCRIPT_TAGS: ReadonlySet<string> = new Set([
  "typescript",
  "typescript-acceptance",
  "typescript-config",
  "typescript-integration",
  "typescript-script",
]);

// Whose source a TypeScript task actually reads: another TypeScript package
// through its exports map, and a ts-config package through the tsconfig it is
// extended from. A Go dependency is read by the Go tasks, never by these.
const READABLE_TAGS: ReadonlySet<string> = new Set([
  ...TYPESCRIPT_TAGS,
  "tsconfig",
]);

const DEPENDENCY_SOURCES = "dependencySources";

const dependencyIds = (
  parsed: z.infer<typeof ProjectFileSchema>,
): readonly string[] =>
  (parsed.dependsOn ?? []).map((entry) =>
    typeof entry === "string" ? entry : entry.id,
  );

/**
 * Every internal dependency publishes its types from src, so a project's tasks
 * read the source of the packages it depends on. Moon does not fold a dependency
 * task's hash into its dependent, so a project that leaves one out serves a
 * cached pass over code its dependency has already broken. This reports a
 * dependencySources file group that has drifted from the transitive dependsOn
 * closure in either direction.
 */
export const dependencySourceFailures = (
  files: readonly ProjectFile[],
): readonly string[] => {
  const failures: string[] = [];
  const parsed = new Map<string, z.infer<typeof ProjectFileSchema>>();
  const sourceById = new Map<string, string>();

  for (const file of files) {
    let input: unknown;
    try {
      input = parseYaml(file.text);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      failures.push(`${file.id} is not valid YAML: ${message}`);
      continue;
    }
    const result = ProjectFileSchema.safeParse(input);
    if (!result.success) {
      failures.push(`${file.id} has an unreadable project definition`);
      continue;
    }
    parsed.set(file.id, result.data);
    sourceById.set(file.id, file.source);
  }

  const closure = (id: string, seen: Set<string>): Set<string> => {
    for (const dependency of dependencyIds(parsed.get(id) ?? {})) {
      if (seen.has(dependency) || !parsed.has(dependency)) continue;
      seen.add(dependency);
      closure(dependency, seen);
    }
    return seen;
  };

  for (const [id, project] of parsed) {
    const tags = project.tags ?? [];
    if (tags.every((tag) => !TYPESCRIPT_TAGS.has(tag))) continue;

    const expected = new Set(
      [...closure(id, new Set<string>())]
        .filter((dependency) =>
          (parsed.get(dependency)?.tags ?? []).some((tag) =>
            READABLE_TAGS.has(tag),
          ),
        )
        .map((dependency) => sourceById.get(dependency))
        .filter((source): source is string => source !== undefined)
        .map((source) => `/${source}/src/**/*`),
    );
    const group = project.fileGroups?.[DEPENDENCY_SOURCES];
    const declared =
      group === undefined || group === null
        ? new Set<string>()
        : new Set(group);
    // The template's default is the project's own sources, which every task
    // already reads, so a project with no internal dependencies leaves it alone.
    if (expected.size === 0) continue;

    const missing = [...expected.difference(declared)];
    if (missing.length > 0) {
      failures.push(
        `${id} file group '${DEPENDENCY_SOURCES}' is missing ${missing.toSorted().join(", ")}: a dependency whose source is not an input leaves this project's hash unchanged when that dependency changes, and moon serves a cached pass`,
      );
    }
    const extra = [...declared.difference(expected)].filter(
      (glob) => glob !== "src/**/*",
    );
    if (extra.length > 0) {
      failures.push(
        `${id} file group '${DEPENDENCY_SOURCES}' declares ${extra.toSorted().join(", ")}, which dependsOn does not reach: drop it or add the dependency`,
      );
    }
  }

  return failures.toSorted();
};
