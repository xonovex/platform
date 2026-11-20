import type {GitRepository} from "@xonovex/script-moon-common/git";
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

const findChangelogRange = (
  pkgDir: string,
  currentVersion: string,
  git: Pick<GitRepository, "pathHistory" | "readFile" | "hasParent">,
): ChangelogRange | undefined => {
  // The first historical package version that differs bounds the current release.
  const path = `${pkgDir}/package.json`;
  const hashes = git.pathHistory(path);

  for (const hash of hashes) {
    const oldPkgJson = git.readFile(hash, path);
    if (oldPkgJson === undefined) continue;
    const oldVersion = parsePackageJson(oldPkgJson, `${hash}:${path}`).version;
    if (oldVersion !== currentVersion) {
      return {since: hash};
    }
  }

  // A never-versioned package starts before the commit that introduced it, and
  // a root commit has no parent to name, so its range is the whole history.
  const earliest = hashes.at(-1);
  if (earliest === undefined) return undefined;
  return {
    since: git.hasParent(earliest) ? `${earliest}~1` : undefined,
  };
};

const CONVENTIONAL_COMMIT_RE = /^\w+(?:\([^)]*\))?:\s*.+$/;

const getCommitsSince = (
  pkgDir: string,
  sinceRef: string | undefined,
  git: Pick<GitRepository, "commitsSince">,
): readonly Commit[] =>
  git.commitsSince(sinceRef, pkgDir).map((record) => {
    const bodyLines = record.message === "" ? [] : record.message.split("\n");
    const messages = bodyLines.filter((line) =>
      CONVENTIONAL_COMMIT_RE.test(line.trim()),
    );
    return {
      hash: record.hash,
      author: record.author,
      messages:
        messages.length > 0
          ? messages.map((line) => line.trim())
          : bodyLines.slice(0, 1).map((line) => line.trim()),
    };
  });

export {
  findChangelogRange,
  getCommitsSince,
  parseConventionalCommit,
  isIncludedType,
  REPO_URL,
};
export type {ChangelogRange, Commit};
