import {execSync} from "node:child_process";

interface Commit {
  readonly hash: string;
  readonly author: string;
  readonly message: string;
}

interface ParsedCommit {
  readonly type: string;
  readonly description: string;
}

const REPO_URL = "https://github.com/xonovex/platform";

const INCLUDED_TYPES = new Set(["feat", "fix", "refactor", "perf", "docs"]);

const parseConventionalCommit = (message: string): ParsedCommit | undefined => {
  const match = /^(\w+)(?:\([^)]*\))?:\s*(.+)$/.exec(message);
  if (!match?.[1] || !match[2]) return undefined;
  return {type: match[1], description: match[2]};
};

const isIncludedType = (type: string): boolean => INCLUDED_TYPES.has(type);

const getLastVersionRef = (
  rootDir: string,
  pkgDir: string,
): string | undefined => {
  const currentPkgJson = execSync(`git show HEAD:${pkgDir}/package.json`, {
    cwd: rootDir,
    encoding: "utf8",
  });
  const currentVersion = (JSON.parse(currentPkgJson) as {version?: string})
    .version;

  // Walk back through commits that touched package.json to find where version differs
  const hashes = execSync(
    `git log --diff-filter=M --format=%H -- ${pkgDir}/package.json`,
    {cwd: rootDir, encoding: "utf8"},
  )
    .trim()
    .split("\n")
    .filter(Boolean);

  for (const hash of hashes) {
    try {
      const oldPkgJson = execSync(`git show ${hash}:${pkgDir}/package.json`, {
        cwd: rootDir,
        encoding: "utf8",
      });
      const oldVersion = (JSON.parse(oldPkgJson) as {version?: string}).version;
      if (oldVersion !== currentVersion) {
        return hash;
      }
    } catch {
      // Commit may not have this file, skip
    }
  }

  return undefined;
};

const getCommitsSince = (
  rootDir: string,
  pkgDir: string,
  sinceRef: string,
): readonly Commit[] => {
  const raw = execSync(
    `git log --format="%H|%aN|%s" ${sinceRef}..HEAD -- ${pkgDir}`,
    {cwd: rootDir, encoding: "utf8"},
  );
  return raw
    .trim()
    .split("\n")
    .filter(Boolean)
    .map((line) => {
      const parts = line.split("|");
      return {
        hash: parts[0] ?? "",
        author: parts[1] ?? "",
        message: parts.slice(2).join("|"),
      };
    });
};

export {
  getLastVersionRef,
  getCommitsSince,
  parseConventionalCommit,
  isIncludedType,
  REPO_URL,
  INCLUDED_TYPES,
};
export type {Commit, ParsedCommit};
