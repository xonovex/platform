import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {getGitVersion, type VersionGit} from "./git.js";
import type {WorkspacePackage} from "./lockstep.js";

/** Each manifest as the tree holds it, with its version at HEAD. */
export const readWorkspacePackages = (
  rootDir: string,
  git: VersionGit,
  packagePaths: readonly string[],
  fs: FileSystem,
): readonly WorkspacePackage[] =>
  packagePaths.map((path) => ({
    path,
    pkg: readPkg(path, fs),
    headVersion: getGitVersion(rootDir, path, git),
  }));
