import {dirname, join} from "node:path";
import {nodeFileSystem} from "./adapters/file-system-node.js";
import type {FileSystem} from "./file-system.js";

export const resolveSkillDirectory = (
  base: string,
  fs: FileSystem = nodeFileSystem,
): string => {
  if (fs.isFile(join(base, "SKILL.md"))) return base;
  const nested = fs
    .readDirectory(base)
    .map((entry) => join(base, entry))
    .filter((path) => fs.isFile(join(path, "SKILL.md")));
  if (nested.length > 1) {
    throw new Error(
      `multiple SKILL.md found under ${base}; pass one explicitly`,
    );
  }
  return nested[0] ?? base;
};

export const findPluginDirectory = (
  skillDirectory: string,
  fs: FileSystem = nodeFileSystem,
): string => {
  const parent = dirname(skillDirectory);
  return fs.isFile(join(dirname(parent), ".claude-plugin", "plugin.json"))
    ? dirname(parent)
    : parent;
};
