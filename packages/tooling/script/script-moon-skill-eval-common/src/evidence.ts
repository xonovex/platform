import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Scratch} from "./scratch.js";

export const writeJsonEvidence = (
  fs: FileSystem,
  path: string,
  value: unknown,
): void => {
  fs.writeFile(path, `${JSON.stringify(value, null, 2)}\n`);
};

export const writeJsonLinesEvidence = (
  fs: FileSystem,
  path: string,
  records: readonly unknown[],
): void => {
  fs.writeFile(
    path,
    `${records.map((record) => JSON.stringify(record)).join("\n")}\n`,
  );
};

// Evidence of an earlier run in the same place must not survive a run that fails
// before it writes its own, or a reader would take the stale file for the new one.
export const clearEvidence = (
  scratch: Scratch,
  paths: readonly string[],
): void => {
  for (const path of paths) scratch.remove(path);
};
