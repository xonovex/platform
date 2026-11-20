import {dirname, join} from "node:path";
import type {FileSystem} from "./file-system.js";

export const findWorkspaceRoot = (
  start: string,
  fs: FileSystem,
  root?: string,
): string => {
  let dir = start;
  while (dir !== dirname(dir)) {
    const marker = join(dir, ".moon");
    if (fs.isDirectory(marker) || fs.isFile(marker)) return dir;
    if (root !== undefined && dir === root) break;
    dir = dirname(dir);
  }
  throw new Error("Could not find workspace root");
};
