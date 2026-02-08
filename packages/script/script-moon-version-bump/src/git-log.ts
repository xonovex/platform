import {execSync} from "node:child_process";

interface Commit {
  readonly hash: string;
  readonly author: string;
  readonly messages: readonly string[];
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

const isIncludedType = (
  type: string,
  includedTypes?: ReadonlySet<string>,
): boolean => (includedTypes ?? INCLUDED_TYPES).has(type);

const getLastVersionRef = (
  rootDir: string,
  pkgDir: string,
  currentVersion: string,
): string | undefined => {
  // Walk back through commits that touched package.json to find where version differs
  const hashes = execSync(`git log --format=%H -- ${pkgDir}/package.json`, {
    cwd: rootDir,
    encoding: "utf8",
  })
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

  // Fall back to the parent of the earliest commit that introduced this package
  const earliest = hashes.at(-1);
  return earliest ? `${earliest}~1` : undefined;
};

const CONVENTIONAL_COMMIT_RE = /^\w+(?:\([^)]*\))?:\s*.+$/;

const getCommitsSince = (
  rootDir: string,
  pkgDir: string,
  sinceRef: string,
): readonly Commit[] => {
  const raw = execSync(
    `git log --format="%x00%H|%aN%n%B" ${sinceRef}..HEAD -- ${pkgDir}`,
    {cwd: rootDir, encoding: "utf8"},
  );
  return raw
    .split("\0")
    .filter(Boolean)
    .map((record) => {
      const lines = record.trim().split("\n");
      const header = lines[0] ?? "";
      const separatorIndex = header.indexOf("|");
      const hash = header.slice(0, separatorIndex);
      const author = header.slice(separatorIndex + 1);
      const bodyLines = lines.slice(1);
      const messages = bodyLines.filter((line) =>
        CONVENTIONAL_COMMIT_RE.test(line.trim()),
      );
      return {
        hash,
        author,
        messages:
          messages.length > 0
            ? messages.map((l) => l.trim())
            : bodyLines.slice(0, 1).map((l) => l.trim()),
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
