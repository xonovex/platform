import {join} from "node:path";
import type {FileSystem} from "./file-system.js";

export const pluginSkillDirectories = (
  pluginDirectory: string,
  fs: FileSystem,
): readonly string[] => {
  const root = join(pluginDirectory, "skills");
  if (!fs.isDirectory(root)) return [];
  return fs
    .readDirectory(root)
    .toSorted()
    .map((name) => join(root, name))
    .filter((path) => fs.isFile(join(path, "SKILL.md")));
};
