import {execFileSync} from "node:child_process";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {parsePackageJson} from "@xonovex/script-moon-common/package-json";

interface Commit {
  readonly hash: string;
  readonly author: string;
  readonly messages: readonly string[];
}

interface ParsedCommit {
  readonly type: string;
  readonly description: string;
}

interface ChangelogRange {
  // undefined covers the whole history, which is what a package introduced by
  // the root commit needs.
  readonly since: string | undefined;
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

const hasParent = (rootDir: string, hash: string): boolean => {
  const parents = execFileSync(
    resolveExecutable("git"),
    ["rev-list", "--parents", "-n", "1", hash],
    {cwd: rootDir, encoding: "utf8"},
  )
    .trim()
    .split(" ");
  return parents.length > 1;
};

const findChangelogRange = (
  rootDir: string,
  pkgDir: string,
  currentVersion: string,
): ChangelogRange | undefined => {
  // The first historical package version that differs bounds the current release.
  const hashes = execFileSync(
    resolveExecutable("git"),
    ["log", "--format=%H", "--", `${pkgDir}/package.json`],
    {
      cwd: rootDir,
      encoding: "utf8",
    },
  )
    .trim()
    .split("\n")
    .filter(Boolean);

  for (const hash of hashes) {
    const path = `${pkgDir}/package.json`;
    const listed = execFileSync(
      resolveExecutable("git"),
      ["ls-tree", "-z", "--name-only", hash, "--", path],
      {cwd: rootDir, encoding: "utf8"},
    );
    if (listed.length === 0) continue;

    const oldPkgJson = execFileSync(
      resolveExecutable("git"),
      ["show", `${hash}:${path}`],
      {
        cwd: rootDir,
        encoding: "utf8",
      },
    );
    const oldVersion = parsePackageJson(oldPkgJson, `${hash}:${path}`).version;
    if (oldVersion !== currentVersion) {
      return {since: hash};
    }
  }

  // A never-versioned package starts before the commit that introduced it, and
  // a root commit has no parent to name, so its range is the whole history.
  const earliest = hashes.at(-1);
  if (earliest === undefined) return undefined;
  return {since: hasParent(rootDir, earliest) ? `${earliest}~1` : undefined};
};

const CONVENTIONAL_COMMIT_RE = /^\w+(?:\([^)]*\))?:\s*.+$/;

const getCommitsSince = (
  rootDir: string,
  pkgDir: string,
  sinceRef: string | undefined,
): readonly Commit[] => {
  const raw = execFileSync(
    resolveExecutable("git"),
    [
      "log",
      "--format=%x00%H|%aN%n%B",
      ...(sinceRef === undefined ? ["HEAD"] : [`${sinceRef}..HEAD`]),
      "--",
      pkgDir,
    ],
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
  findChangelogRange,
  getCommitsSince,
  parseConventionalCommit,
  isIncludedType,
  REPO_URL,
};
export type {ChangelogRange, Commit};
