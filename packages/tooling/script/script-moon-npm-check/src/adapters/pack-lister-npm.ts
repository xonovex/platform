import {spawnSync} from "node:child_process";
import type {ExecutableResolver} from "@xonovex/script-moon-common/executable";
import {parsePackagedFilePaths, type PackLister} from "../packed-package.js";

// The file listing of a large package exceeds the default buffer.
const MAX_OUTPUT_BYTES = 10_485_760;

/** Asks npm which files it would pack, without writing a tarball. */
export const npmPackLister =
  (resolveExecutable: ExecutableResolver): PackLister =>
  (packageRoot) => {
    const result = spawnSync(
      resolveExecutable("npm"),
      ["pack", "--dry-run", "--json"],
      {cwd: packageRoot, encoding: "utf8", maxBuffer: MAX_OUTPUT_BYTES},
    );
    if (result.error) throw result.error;
    if (result.status !== 0) {
      const detail = result.stderr.trim() || `exit ${String(result.status)}`;
      throw new Error(`npm pack failed: ${detail}`);
    }
    return parsePackagedFilePaths(result.stdout);
  };
