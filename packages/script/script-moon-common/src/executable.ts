import {accessSync, constants, statSync} from "node:fs";
import {delimiter, isAbsolute, resolve, sep} from "node:path";

const executableExtensions = (): readonly string[] =>
  process.platform === "win32"
    ? (process.env.PATHEXT ?? ".COM;.EXE;.BAT;.CMD").split(";").filter(Boolean)
    : [""];

export const resolveExecutable = (
  name: string,
  searchPath = process.env.PATH,
): string => {
  if (
    name.length === 0 ||
    name.includes(sep) ||
    (process.platform === "win32" && name.includes("/"))
  ) {
    throw new Error(
      `executable name must be a bare command: ${JSON.stringify(name)}`,
    );
  }
  if (searchPath === undefined || searchPath.length === 0) {
    throw new Error(
      `cannot resolve executable ${JSON.stringify(name)}: PATH is empty`,
    );
  }

  for (const directory of searchPath
    .split(delimiter)
    .filter((entry) => isAbsolute(entry))) {
    for (const extension of executableExtensions()) {
      const candidate = resolve(directory, `${name}${extension}`);
      try {
        accessSync(candidate, constants.X_OK);
        if (statSync(candidate).isFile()) return candidate;
      } catch {
        // Continue searching the remaining PATH entries.
      }
    }
  }

  throw new Error(`executable ${JSON.stringify(name)} was not found on PATH`);
};
