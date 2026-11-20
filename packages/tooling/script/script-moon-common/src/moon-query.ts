import {join} from "node:path";
import {z} from "zod";
import type {FileSystem} from "./file-system.js";

const MoonProjectSchema = z.looseObject({
  id: z.string().min(1),
  source: z.string().min(1),
});

const MoonQuerySchema = z.object({
  projects: z.array(MoonProjectSchema),
});

export type MoonProject = z.infer<typeof MoonProjectSchema>;

export const parseMoonProjects = (output: string): readonly MoonProject[] => {
  let value: unknown;
  try {
    value = JSON.parse(output);
  } catch (error) {
    throw new Error("invalid Moon project query output: malformed JSON", {
      cause: error,
    });
  }
  const result = MoonQuerySchema.safeParse(value);
  if (!result.success) {
    throw new Error(
      `invalid Moon project query output: ${z.prettifyError(result.error)}`,
    );
  }
  return result.data.projects;
};

/** The package.json of every project that has one, as absolute paths. */
export const packageJsonPaths = (
  rootDir: string,
  projects: readonly MoonProject[],
  fs: FileSystem,
): readonly string[] =>
  projects
    .map((project) => join(rootDir, project.source, "package.json"))
    .filter((path) => fs.isFile(path));
