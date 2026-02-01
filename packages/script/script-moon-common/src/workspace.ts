import {existsSync} from "node:fs";
import {dirname, join} from "node:path";

export const findWorkspaceRoot = (start: string, root?: string): string => {
  let dir = start;
  while (dir !== dirname(dir)) {
    if (existsSync(join(dir, ".moon"))) return dir;
    if (root !== undefined && dir === root) break;
    dir = dirname(dir);
  }
  throw new Error("Could not find workspace root");
};
