import {existsSync} from "node:fs";
import {dirname, join} from "node:path";

export const findWorkspaceRoot = (start: string): string => {
  let dir = start;
  while (dir !== dirname(dir)) {
    if (existsSync(join(dir, ".moon"))) return dir;
    dir = dirname(dir);
  }
  throw new Error("Could not find workspace root");
};
