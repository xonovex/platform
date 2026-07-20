import {resolve} from "node:path";
import {fileURLToPath} from "node:url";

export const isDirectExecution = (
  moduleUrl: string,
  entrypoint: string | undefined,
): boolean =>
  entrypoint !== undefined &&
  resolve(fileURLToPath(moduleUrl)) === resolve(entrypoint);
