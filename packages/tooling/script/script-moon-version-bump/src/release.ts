import {basename, dirname, relative} from "node:path";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import {findPreviousVersionCommit} from "./git-log.js";

/** The git reads that tell whether a package was ever released. */
type ReleaseGit = Pick<GitRepository, "pathHistory" | "readFile" | "tags">;

/** Whether the package whose manifest is at path was released at version. */
type ReleasePredicate = (path: string, version: string) => boolean;

const RELEASE_TAG_VERSION = /^v\d+\.\d+\.\d+/;

const hasReleaseTag = (project: string, tags: readonly string[]): boolean =>
  tags.some(
    (tag) =>
      tag.startsWith(`${project}-`) &&
      RELEASE_TAG_VERSION.test(tag.slice(project.length + 1)),
  );

/**
 * Answers from the two records a release leaves in git. A committed version
 * moves only through a version bump, which ships as a release, so an earlier
 * committed version that differs records one. A GitHub release tags its line
 * `<project directory>-v<version>`, which also records a first release that
 * kept its initial version. A package with neither was never released.
 */
const releaseHistory = (rootDir: string, git: ReleaseGit): ReleasePredicate => {
  const tags = git.tags();
  return (path, version) =>
    hasReleaseTag(basename(dirname(path)), tags) ||
    findPreviousVersionCommit(
      relative(rootDir, dirname(path)),
      version,
      git,
    ) !== undefined;
};

export {releaseHistory};
export type {ReleaseGit, ReleasePredicate};
