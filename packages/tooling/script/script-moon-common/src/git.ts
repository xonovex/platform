/** One commit as the log reports it. */
export interface CommitRecord {
  readonly hash: string;
  readonly author: string;
  /** The full commit message, subject and body. */
  readonly message: string;
}

/**
 * The reads the scripts make from one git repository. Every path is relative to
 * the repository root. A consumer depends on the part it reads through Pick, so a
 * unit spec supplies only those methods.
 */
export interface GitRepository {
  /** Resolves ref to the commit it names, throwing when it names none. */
  readonly resolveCommit: (ref: string) => string;
  /** The text of path at commit, or undefined when commit does not hold path. */
  readonly readFile: (commit: string, path: string) => string | undefined;
  /** The hashes of the commits that touched path, newest first. */
  readonly pathHistory: (path: string) => readonly string[];
  readonly hasParent: (commit: string) => boolean;
  /**
   * The commits reachable from HEAD that touched path, newest first. A defined
   * since excludes since and every commit it reaches.
   */
  readonly commitsSince: (
    since: string | undefined,
    path: string,
  ) => readonly CommitRecord[];
  /** The committed files of the working tree. */
  readonly trackedFiles: () => readonly string[];
  /** The paths git ignores, a fully ignored directory collapsed to one entry. */
  readonly ignoredPaths: () => readonly string[];
}

/** The log format parseCommitLog reads: a NUL, then hash|author, then the message. */
export const COMMIT_LOG_FORMAT = "%x00%H|%aN%n%B";

export const parseCommitLog = (output: string): readonly CommitRecord[] =>
  output
    .split("\0")
    .filter((record) => record.trim().length > 0)
    .map((record) => {
      const [header = "", ...messageLines] = record.trim().split("\n");
      const separatorIndex = header.indexOf("|");
      return {
        hash: header.slice(0, separatorIndex),
        author: header.slice(separatorIndex + 1),
        message: messageLines.join("\n"),
      };
    });

/** The non-empty lines of a git listing. */
export const parseGitLines = (output: string): readonly string[] =>
  output.split("\n").filter((line) => line !== "");
