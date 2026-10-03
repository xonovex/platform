import {delimiter, isAbsolute, resolve, sep} from "node:path";
import {nodeFileSystem, type FileSystem} from "./file-system.js";

const executableExtensions = (): readonly string[] =>
  process.platform === "win32"
    ? (process.env.PATHEXT ?? ".COM;.EXE;.BAT;.CMD").split(";").filter(Boolean)
    : [""];

export const resolveExecutable = (
  name: string,
  searchPath = process.env.PATH,
  fs: FileSystem = nodeFileSystem,
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
      if (fs.isExecutableFile(candidate)) return candidate;
    }
  }

  throw new Error(`executable ${JSON.stringify(name)} was not found on PATH`);
};
