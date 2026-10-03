import {join} from "node:path";
import {nodeFileSystem, type FileSystem} from "./file-system.js";

export interface PackageLayout {
  readonly groups: readonly string[];
  readonly packages: readonly string[];
}

export const packageLayout = (
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): PackageLayout => {
  const visit = (directory: string): PackageLayout => {
    if (
      fs.isFile(join(directory, "package.json")) ||
      fs.isFile(join(directory, "moon.yml"))
    ) {
      return {groups: [], packages: [directory]};
    }
    const children = fs
      .readDirectory(directory)
      .toSorted()
      .map((name) => join(directory, name))
      .filter((path) => fs.isDirectory(path))
      .map(visit);
    return {
      groups: [directory, ...children.flatMap((child) => child.groups)],
      packages: children.flatMap((child) => child.packages),
    };
  };
  return visit(join(repositoryRoot, "packages"));
};

export const pluginSkillDirectories = (
  pluginDirectory: string,
  fs: FileSystem = nodeFileSystem,
): readonly string[] => {
  const root = join(pluginDirectory, "skills");
  if (!fs.isDirectory(root)) return [];
  return fs
    .readDirectory(root)
    .toSorted()
    .map((name) => join(root, name))
    .filter((path) => fs.isFile(join(path, "SKILL.md")));
};
