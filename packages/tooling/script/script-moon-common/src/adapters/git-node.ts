import {execFileSync} from "node:child_process";
import type {ExecutableResolver} from "../executable.js";
import {
  COMMIT_LOG_FORMAT,
  parseCommitLog,
  parseGitLines,
  type GitRepository,
} from "../git.js";

// The log of a long history exceeds the default buffer.
const MAX_OUTPUT_BYTES = 64 * 1024 * 1024;

/** Reads the repository at rootDir by running the git that resolveExecutable finds. */
export const nodeGitRepository =
  (resolveExecutable: ExecutableResolver) =>
  (rootDir: string): GitRepository => {
    const git = (args: readonly string[]): string =>
      execFileSync(resolveExecutable("git"), [...args], {
        cwd: rootDir,
        encoding: "utf8",
        maxBuffer: MAX_OUTPUT_BYTES,
        stdio: ["ignore", "pipe", "pipe"],
      });
    return {
      resolveCommit: (ref) =>
        git([
          "rev-parse",
          "--verify",
          "--end-of-options",
          `${ref}^{commit}`,
        ]).trim(),
      readFile: (commit, path) =>
        git(["ls-tree", "-z", "--name-only", commit, "--", path]).length === 0
          ? undefined
          : git(["show", `${commit}:${path}`]),
      pathHistory: (path) =>
        parseGitLines(git(["log", "--format=%H", "--", path])),
      hasParent: (commit) =>
        git(["rev-list", "--parents", "-n", "1", commit]).trim().split(" ")
          .length > 1,
      tags: () => parseGitLines(git(["tag", "--list"])),
      commitsSince: (since, path) =>
        parseCommitLog(
          git([
            "log",
            `--format=${COMMIT_LOG_FORMAT}`,
            since === undefined ? "HEAD" : `${since}..HEAD`,
            "--",
            path,
          ]),
        ),
    };
  };
