import {basename} from "node:path";
import {parse as parseYaml} from "yaml";
import {z} from "zod";

const TaskDepSchema = z.union([z.string(), z.object({target: z.string()})]);

const TaskSchema = z.object({
  deps: z.array(TaskDepSchema).optional(),
  options: z.object({mergeDeps: z.string().optional()}).optional(),
});

const TagTaskFileSchema = z.object({
  extends: z.string().optional(),
  tasks: z.record(z.string(), TaskSchema).optional(),
});

type TagTaskDefinition = z.infer<typeof TagTaskFileSchema>;

export interface TagTaskFile {
  // name is the file's base name, which is how `extends` refers to a sibling.
  readonly name: string;
  readonly text: string;
}

const depTargets = (
  deps: z.infer<typeof TaskSchema>["deps"],
): readonly string[] =>
  (deps ?? []).map((dep) => (typeof dep === "string" ? dep : dep.target));

// inheritedTaskDeps returns the deps each task carries once `extends` is applied.
// A task the child redefines with its own deps supersedes the parent's entry.
const inheritedTaskDeps = (
  fileName: string,
  parsed: ReadonlyMap<string, TagTaskDefinition>,
  visiting: ReadonlySet<string>,
): Map<string, readonly string[]> => {
  const file = parsed.get(fileName);
  if (file === undefined || visiting.has(fileName)) {
    return new Map<string, readonly string[]>();
  }
  const parentName =
    file.extends === undefined ? undefined : basename(file.extends);
  const deps =
    parentName === undefined
      ? new Map<string, readonly string[]>()
      : inheritedTaskDeps(parentName, parsed, new Set([...visiting, fileName]));
  for (const [task, definition] of Object.entries(file.tasks ?? {})) {
    if (definition.deps !== undefined)
      deps.set(task, depTargets(definition.deps));
  }
  return deps;
};

// taskInheritanceFailures reports a tag file that redefines an inherited task
// with fewer deps than its parent declares. Moon replaces the parent's deps
// rather than appending to them across `extends`, and mergeDeps does not change
// that, so a child adding one dep silently drops the rest of the gate.
export const taskInheritanceFailures = (
  files: readonly TagTaskFile[],
): readonly string[] => {
  const failures: string[] = [];
  const parsed = new Map<string, TagTaskDefinition>();
  for (const file of files) {
    let input: unknown;
    try {
      input = parseYaml(file.text);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : String(error);
      failures.push(`${file.name} is not valid YAML: ${message}`);
      continue;
    }
    const result = TagTaskFileSchema.safeParse(input);
    if (result.success) {
      parsed.set(file.name, result.data);
    } else {
      failures.push(`${file.name} has an unreadable task definition`);
    }
  }

  for (const [fileName, file] of parsed) {
    if (file.extends === undefined) continue;
    const parentName = basename(file.extends);
    if (!parsed.has(parentName)) {
      failures.push(`${fileName} extends missing ${file.extends}`);
      continue;
    }
    const parentDeps = inheritedTaskDeps(
      parentName,
      parsed,
      new Set([fileName]),
    );
    for (const [task, definition] of Object.entries(file.tasks ?? {})) {
      if (definition.deps === undefined) continue;
      if (definition.options?.mergeDeps === "replace") continue;
      const inherited = parentDeps.get(task);
      if (inherited === undefined) continue;
      const declared = new Set(depTargets(definition.deps));
      const dropped = inherited.filter((dep) => !declared.has(dep));
      if (dropped.length > 0) {
        failures.push(
          `${fileName} task '${task}' drops inherited ${dropped.join(", ")}: a task redefined across extends replaces the parent's deps, so restate them or declare mergeDeps: replace`,
        );
      }
    }
  }
  return failures;
};
