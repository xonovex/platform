import {delimiter, isAbsolute, resolve, sep} from "node:path";
import type {FileSystem} from "./file-system.js";

/** The process settings that decide where a bare command name resolves. */
export interface ExecutableEnvironment {
  /** The PATH value, absent when the process has none. */
  readonly searchPath: string | undefined;
  /** The PATHEXT value, which only Windows reads. */
  readonly extensions: string | undefined;
  readonly platform: NodeJS.Platform;
}

const executableExtensions = (
  environment: ExecutableEnvironment,
): readonly string[] =>
  environment.platform === "win32"
    ? (environment.extensions ?? ".COM;.EXE;.BAT;.CMD")
        .split(";")
        .filter(Boolean)
    : [""];

/** Returns the absolute path of a bare command name, or throws when none exists. */
export type ExecutableResolver = (name: string) => string;

export const findExecutable = (
  name: string,
  environment: ExecutableEnvironment,
  fs: FileSystem,
): string => {
  if (
    name.length === 0 ||
    name.includes(sep) ||
    (environment.platform === "win32" && name.includes("/"))
  ) {
    throw new Error(
      `executable name must be a bare command: ${JSON.stringify(name)}`,
    );
  }
  const {searchPath} = environment;
  if (searchPath === undefined || searchPath.length === 0) {
    throw new Error(
      `cannot resolve executable ${JSON.stringify(name)}: PATH is empty`,
    );
  }

  for (const directory of searchPath
    .split(delimiter)
    .filter((entry) => isAbsolute(entry))) {
    for (const extension of executableExtensions(environment)) {
      const candidate = resolve(directory, `${name}${extension}`);
      if (fs.isExecutableFile(candidate)) return candidate;
    }
  }

  throw new Error(`executable ${JSON.stringify(name)} was not found on PATH`);
};
