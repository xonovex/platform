import {relative} from "node:path";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import {
  parsePackageJson,
  type PackageJson,
} from "@xonovex/script-moon-common/package-json";

/** The git reads a version bump makes. */
export type VersionGit = Pick<
  GitRepository,
  "readFile" | "pathHistory" | "hasParent" | "commitsSince" | "tags"
>;

/** The manifest at pkgPath as HEAD holds it, or undefined when HEAD lacks it. */
export const readHeadPackage = (
  rootDir: string,
  pkgPath: string,
  git: Pick<GitRepository, "readFile">,
): PackageJson | undefined => {
  const path = relative(rootDir, pkgPath);
  const text = git.readFile("HEAD", path);
  return text === undefined
    ? undefined
    : parsePackageJson(text, `HEAD:${path}`);
};

export const getGitVersion = (
  rootDir: string,
  pkgPath: string,
  git: Pick<GitRepository, "readFile">,
): string | undefined => readHeadPackage(rootDir, pkgPath, git)?.version;
