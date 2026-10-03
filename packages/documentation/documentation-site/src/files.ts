import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {isAbsolute, join, relative, resolve, sep} from "node:path";
import {parse as parseYaml} from "yaml";
import {z} from "zod";

const excludedDirectories = new Set([
  "node_modules",
  "dist",
  "target",
  "bin",
  "build",
  "coverage",
  "test",
  "tests",
  "testdata",
  "fixtures",
  "plans",
  "content",
  ".git",
  ".moon",
  ".vitepress",
  "__pycache__",
]);

export const repositoryPath = (root: string, path: string): string =>
  relative(root, path).split(sep).join("/");

export const inside = (root: string, path: string): string => {
  const resolved = resolve(root, path);
  const local = relative(root, resolved);
  if (local === ".." || local.startsWith(".." + sep) || isAbsolute(local)) {
    throw new Error("Path escapes its source directory: " + path);
  }
  return resolved;
};

export const readText = (root: string, path: string): string =>
  readFileSync(inside(root, path), "utf8");

export const jsonFile = <T>(
  root: string,
  path: string,
  schema: z.ZodType<T>,
): T => {
  const raw: unknown = JSON.parse(readText(root, path));
  return schema.parse(raw);
};

const frontmatterSchema = z.looseObject({
  name: z.string().optional(),
  description: z.string().optional(),
});

export const markdown = (
  text: string,
): {
  readonly body: string;
  readonly name: string | undefined;
  readonly description: string | undefined;
} => {
  const match = /^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/u.exec(text);
  const raw: unknown = match?.[1] === undefined ? {} : parseYaml(match[1]);
  const metadata = frontmatterSchema.parse(raw ?? {});
  return {
    body: match === null ? text : text.slice(match[0].length),
    name: metadata.name,
    description: metadata.description,
  };
};

export const filesBelow = (
  root: string,
  directory: string,
): readonly string[] => {
  const path = inside(root, directory);
  if (!existsSync(path)) return [];
  return readdirSync(path, {withFileTypes: true})
    .toSorted((a, b) => a.name.localeCompare(b.name))
    .flatMap((entry) => {
      if (entry.isSymbolicLink()) return [];
      const next = join(directory, entry.name);
      if (entry.isDirectory())
        return excludedDirectories.has(entry.name)
          ? []
          : filesBelow(root, next);
      return [next.split(sep).join("/")];
    });
};

export const exists = (root: string, path: string): boolean =>
  existsSync(inside(root, path));
export const isFile = (root: string, path: string): boolean =>
  exists(root, path) && statSync(inside(root, path)).isFile();

export const documentTitle = (body: string, fallback: string): string =>
  /^#\s+(.+)$/mu.exec(body)?.[1]?.trim() ?? fallback;
