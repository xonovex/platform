import {spawnSync} from "node:child_process";
import type {ExecutableResolver} from "@xonovex/script-moon-common/executable";
import type {NpmView} from "../publish.js";

/** Asks the npm registry, through the npm binary, whether a version exists. */
export const runNpmView =
  (resolveExecutable: ExecutableResolver): NpmView =>
  ({name, version}) => {
    const result = spawnSync(
      resolveExecutable("npm"),
      ["view", `${name}@${version}`, "version"],
      {
        encoding: "utf8",
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    return {
      ...(result.error ? {error: result.error} : {}),
      status: result.status,
      stderr: result.stderr,
    };
  };
