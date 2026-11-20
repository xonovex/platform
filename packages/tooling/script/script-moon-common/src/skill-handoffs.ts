import {join} from "node:path";
import {nodeFileSystem} from "./adapters/file-system-node.js";
import type {FileSystem} from "./file-system.js";
import {pluginSkillDirectories} from "./package-layout.js";
import {boldSkillNameMentions} from "./skill-name.js";

const markdownFiles = (directory: string, fs: FileSystem): readonly string[] =>
  fs.isDirectory(directory)
    ? fs
        .readDirectory(directory)
        .toSorted()
        .filter((entry) => entry.endsWith(".md"))
        .map((entry) => join(directory, entry))
        .filter((path) => fs.isFile(path))
    : [];

// pluginHandoffNames returns each skill name that a SKILL.md or a reference of
// the plugin sets in bold, the one written form of a handoff to another skill.
export const pluginHandoffNames = (
  pluginDirectory: string,
  fs: FileSystem = nodeFileSystem,
): ReadonlySet<string> =>
  new Set(
    pluginSkillDirectories(pluginDirectory, fs)
      .flatMap((directory) => [
        join(directory, "SKILL.md"),
        ...markdownFiles(join(directory, "references"), fs),
      ])
      .flatMap((path) => boldSkillNameMentions(fs.readText(path))),
  );
