import {execFileSync} from "node:child_process";
import {join} from "node:path";
import {z} from "zod";
import {resolveExecutable} from "./executable.js";
import {nodeFileSystem, type FileSystem} from "./file-system.js";

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

export const queryMoonProjects = (rootDir: string): readonly MoonProject[] => {
  const output = execFileSync(
    resolveExecutable("npx"),
    ["moon", "query", "projects"],
    {
      encoding: "utf8",
      maxBuffer: 10 * 1024 * 1024,
      cwd: rootDir,
    },
  );
  return parseMoonProjects(output);
};

export const findAllPackageJsonPaths = (
  rootDir: string,
  fs: FileSystem = nodeFileSystem,
): readonly string[] =>
  queryMoonProjects(rootDir)
    .map((p) => join(rootDir, p.source, "package.json"))
    .filter((p) => fs.isFile(p));
