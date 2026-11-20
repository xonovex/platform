import {basename, join, resolve, sep} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {resolveSkillDirectory} from "@xonovex/script-moon-common/skill-directory";

// Every path here stays as the caller wrote it, relative or absolute, because
// the report prints it; the filesystem reads it resolved against cwd.

// The shared skill-directory resolver builds its paths and its error message from
// the path it is given, so this view lets it work on the path as written.
const anchoredAt = (cwd: string, fs: FileSystem) => ({
  readDirectory: (path: string) => fs.readDirectory(resolve(cwd, path)),
  isFile: (path: string) => fs.isFile(resolve(cwd, path)),
  isDirectory: (path: string) => fs.isDirectory(resolve(cwd, path)),
});

/** The SOURCES.md a target names: the file itself, or the one in its skill directory. */
export const resolveSourcesFile = (
  target: string,
  cwd: string,
  fs: FileSystem,
): string | undefined => {
  const view = anchoredAt(cwd, fs);
  if (view.isFile(target) && basename(target) === "SOURCES.md") return target;
  if (!view.isDirectory(target)) return undefined;
  if (view.isFile(join(target, "SOURCES.md"))) {
    return join(target, "SOURCES.md");
  }
  const sourcesFile = join(resolveSkillDirectory(target, view), "SOURCES.md");
  return view.isFile(sourcesFile) ? sourcesFile : undefined;
};

// comparePaths orders path components before comparing path depth.
export const comparePaths = (a: string, b: string): number => {
  const partsA = a.split(sep);
  const partsB = b.split(sep);
  const len = Math.min(partsA.length, partsB.length);
  for (let i = 0; i < len; i++) {
    const ca = partsA[i] ?? "";
    const cb = partsB[i] ?? "";
    if (ca < cb) return -1;
    if (ca > cb) return 1;
  }
  return partsA.length - partsB.length;
};

/** Every SOURCES.md beside a SKILL.md below root, outside node_modules, in path order. */
export const findSourcesFiles = (
  root: string,
  cwd: string,
  fs: FileSystem,
): string[] => {
  const found: string[] = [];
  const walk = (directory: string): void => {
    const absolute = resolve(cwd, directory);
    let entries: string[];
    try {
      entries = [...fs.readDirectory(absolute)];
    } catch {
      return;
    }
    for (const entry of entries) {
      const full = join(directory, entry);
      if (fs.isDirectory(join(absolute, entry))) {
        if (entry === "node_modules") continue;
        walk(full);
      } else if (
        entry === "SOURCES.md" &&
        fs.isFile(join(absolute, "SKILL.md"))
      ) {
        found.push(full);
      }
    }
  };
  walk(root);
  return found.toSorted(comparePaths);
};

/** The Markdown file names directly in the references directory of a skill. */
export const listExistingRefs = (
  skillDirectory: string,
  fs: FileSystem,
): ReadonlySet<string> => {
  const refsDir = join(skillDirectory, "references");
  if (!fs.isDirectory(refsDir)) return new Set<string>();
  return new Set(
    fs
      .readDirectory(refsDir)
      .filter(
        (entry) => entry.endsWith(".md") && fs.isFile(join(refsDir, entry)),
      ),
  );
};

// A checkout path is relative to the nearest Moon workspace; outside one it is
// relative to the start directory.
export const findWorkspaceRoot = (start: string, fs: FileSystem): string => {
  let dir = start;
  for (;;) {
    if (fs.isDirectory(join(dir, ".moon"))) return dir;
    const parent = resolve(dir, "..");
    if (parent === dir) return start;
    dir = parent;
  }
};
