import {execFileSync} from "node:child_process";
import {
  COMMIT_LOG_FORMAT,
  parseCommitLog,
  parseGitLines,
  type GitRepository,
} from "../git.js";
import {resolveExecutable} from "./executable-node.js";

// A listing of every ignored path in a large checkout exceeds the default buffer.
const MAX_OUTPUT_BYTES = 64 * 1024 * 1024;

/** Reads the repository at rootDir by running git there. */
export const nodeGitRepository = (rootDir: string): GitRepository => {
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
    trackedFiles: () => parseGitLines(git(["ls-files"])),
    ignoredPaths: () =>
      parseGitLines(
        git([
          "ls-files",
          "--others",
          "--ignored",
          "--exclude-standard",
          "--directory",
        ]),
      ),
  };
};
