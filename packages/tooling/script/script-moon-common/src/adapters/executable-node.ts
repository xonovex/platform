import {findExecutable} from "../executable.js";
import {nodeFileSystem} from "./file-system-node.js";

/** Resolves a bare command name on the PATH of the current process. */
export const resolveExecutable = (name: string): string =>
  findExecutable(
    name,
    {
      searchPath: process.env.PATH,
      extensions: process.env.PATHEXT,
      platform: process.platform,
    },
    nodeFileSystem,
  );
