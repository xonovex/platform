import {spawnSync} from "node:child_process";
import {resolveExecutable} from "@xonovex/script-moon-common/adapters/executable-node";
import type {NpmView} from "../publish.js";

/** Asks the npm registry, through the npm binary, whether a version exists. */
export const runNpmView: NpmView = ({name, version}) => {
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
