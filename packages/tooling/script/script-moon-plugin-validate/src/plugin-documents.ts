import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";

const markdownFilesBelow = (
  directory: string,
  fs: FileSystem,
): readonly string[] => {
  if (!fs.isDirectory(directory)) return [];
  return fs.readDirectory(directory).flatMap((name) => {
    const path = join(directory, name);
    if (fs.isDirectory(path)) return markdownFilesBelow(path, fs);
    return fs.isFile(path) && name.endsWith(".md") ? [path] : [];
  });
};

// The README and every Markdown file below docs/ are the documents a plugin
// package writes for readers; skills carry their own contract and are not part
// of this set.
export const pluginDocuments = (
  packageDirectory: string,
  fs: FileSystem,
): readonly string[] => {
  const readme = join(packageDirectory, "README.md");
  return [
    ...(fs.isFile(readme) ? [readme] : []),
    ...markdownFilesBelow(join(packageDirectory, "docs"), fs),
  ];
};
