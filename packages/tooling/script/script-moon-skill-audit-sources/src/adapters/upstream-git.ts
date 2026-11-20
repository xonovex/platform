import {execFileSync} from "node:child_process";
import {resolveExecutable} from "@xonovex/script-moon-common/adapters/executable-node";
import type {UpstreamRepository} from "../upstream.js";

// A long history on a watched path exceeds the default buffer.
const MAX_OUTPUT_BYTES = 10 * 1024 * 1024;

const pathArguments = (paths: readonly string[]): readonly string[] =>
  paths.length > 0 ? ["--", ...paths] : [];

/** Reads the checkout at directory by running git there; a failed command reads as empty. */
export const gitUpstreamRepository = (
  directory: string,
): UpstreamRepository => {
  const git = (args: readonly string[]): string | undefined => {
    try {
      return execFileSync(resolveExecutable("git"), [...args], {
        cwd: directory,
        encoding: "utf8",
        stdio: ["ignore", "pipe", "ignore"],
        maxBuffer: MAX_OUTPUT_BYTES,
      });
    } catch {
      return undefined;
    }
  };
  return {
    fetchTags: () => git(["fetch", "--tags", "--quiet"]) !== undefined,
    tags: () => git(["tag", "--list"])?.split("\n") ?? [],
    hasCommit: (ref) =>
      git(["rev-parse", "--verify", "--quiet", `${ref}^{commit}`]) !==
      undefined,
    commits: (range, paths) =>
      (git(["log", "--oneline", range, ...pathArguments(paths)]) ?? "")
        .split("\n")
        .filter((line) => line.trim()),
    changedFiles: (range, paths) =>
      (git(["diff", "--name-only", range, ...pathArguments(paths)]) ?? "")
        .split("\n")
        .map((file) => file.trim())
        .filter(Boolean),
  };
};
