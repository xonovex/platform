import {join} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Scratch} from "../scratch.js";

/** A Scratch that also lists every path it was asked to remove. */
export interface MemoryScratch extends Scratch {
  readonly removed: () => readonly string[];
}

/**
 * A Scratch over a memory FileSystem. It numbers each directory it creates under
 * root, so a case knows where a harness run is staged. The memory FileSystem has
 * no removal, so a removed path stays readable and the case reads removed().
 */
export const memoryScratch = (
  fs: FileSystem,
  root = "/scratch",
): MemoryScratch => {
  const removed: string[] = [];
  let created = 0;
  return {
    makeTemporaryDirectory: (prefix) => {
      created += 1;
      const directory = join(root, `${prefix}${String(created)}`);
      fs.makeDirectory(directory);
      return directory;
    },
    remove: (path) => {
      removed.push(path);
    },
    removed: () => [...removed],
  };
};
