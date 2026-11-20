import {findExecutable, type ExecutableResolver} from "../executable.js";
import type {FileSystem} from "../file-system.js";

/** Resolves a bare command name on the PATH of the current process. */
export const processExecutableResolver =
  (fs: FileSystem): ExecutableResolver =>
  (name) =>
    findExecutable(
      name,
      {
        searchPath: process.env.PATH,
        extensions: process.env.PATHEXT,
        platform: process.platform,
      },
      fs,
    );
