import {dirname, sep} from "node:path";
import type {FileSystem} from "./file-system.js";

/** The tree a memory filesystem starts from: absolute path to file text. */
export interface MemoryTree {
  readonly files?: Readonly<Record<string, string>>;
  /** Directories that hold no file, which a path map cannot otherwise express. */
  readonly directories?: readonly string[];
  /** Files the process may execute, named so PATH resolution can be driven. */
  readonly executables?: readonly string[];
}

// Trailing separators would otherwise make the same directory compare unequal to
// itself, and the root is the one path allowed to end in one.
const normalize = (path: string): string =>
  path.length > 1 && path.endsWith(sep) ? path.slice(0, -1) : path;

const parents = (path: string): readonly string[] => {
  const result: string[] = [];
  let current = dirname(normalize(path));
  while (current !== dirname(current)) {
    result.push(current);
    current = dirname(current);
  }
  result.push(current);
  return result;
};

/**
 * A FileSystem held entirely in memory. Writes land in its own tree and are
 * readable afterwards, so a case can exercise code that reads back what it wrote
 * without leaving anything on disk.
 */
export const memoryFileSystem = (tree: MemoryTree = {}): FileSystem => {
  const files = new Map(Object.entries(tree.files ?? {}));
  const directories = new Set<string>();
  const executables = new Set<string>(
    (tree.executables ?? []).map((path) => normalize(path)),
  );
  // A seeded path implies every directory above it, so a tree given as a flat map
  // still answers isDirectory and readDirectory for its interior.
  for (const path of files.keys()) {
    for (const parent of parents(path)) directories.add(parent);
  }
  for (const path of tree.directories ?? []) {
    directories.add(normalize(path));
    for (const parent of parents(path)) directories.add(parent);
  }

  const isFile = (path: string): boolean => files.has(normalize(path));

  return {
    isFile,
    isDirectory: (path) => directories.has(normalize(path)),
    isExecutableFile: (path) =>
      executables.has(normalize(path)) && isFile(path),
    readText: (path) => {
      const text = files.get(normalize(path));
      if (text === undefined) {
        throw new Error(`ENOENT: no such file or directory, open '${path}'`);
      }
      return text;
    },
    readDirectory: (path) => {
      const directory = normalize(path);
      if (!directories.has(directory)) {
        throw new Error(`ENOENT: no such file or directory, scandir '${path}'`);
      }
      const prefix = directory === sep ? sep : `${directory}${sep}`;
      const names = new Set<string>();
      for (const candidate of [...files.keys(), ...directories]) {
        if (candidate === directory || !candidate.startsWith(prefix)) continue;
        const name = candidate.slice(prefix.length).split(sep)[0];
        if (name !== undefined && name.length > 0) names.add(name);
      }
      return [...names].toSorted();
    },
    writeFile: (path, text) => {
      const file = normalize(path);
      files.set(file, text);
      for (const parent of parents(file)) directories.add(parent);
    },
    makeDirectory: (path) => {
      const directory = normalize(path);
      directories.add(directory);
      for (const parent of parents(directory)) directories.add(parent);
    },
  };
};
