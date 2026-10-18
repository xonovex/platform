import {isAbsolute, relative, resolve, sep} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import type {z} from "zod";

export const inside = (root: string, path: string): string => {
  const resolved = resolve(root, path);
  const local = relative(root, resolved);
  if (local === ".." || local.startsWith(".." + sep) || isAbsolute(local)) {
    throw new Error("Path escapes its source directory: " + path);
  }
  return resolved;
};

export const readText = (
  root: string,
  path: string,
  fs: FileSystem = nodeFileSystem,
): string => fs.readText(inside(root, path));

export const jsonFile = <T>(
  root: string,
  path: string,
  schema: z.ZodType<T>,
  fs: FileSystem = nodeFileSystem,
): T => {
  const raw: unknown = JSON.parse(readText(root, path, fs));
  return schema.parse(raw);
};

export const exists = (
  root: string,
  path: string,
  fs: FileSystem = nodeFileSystem,
): boolean => fs.isFile(inside(root, path));
