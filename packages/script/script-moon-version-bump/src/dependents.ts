import type {PackageJson} from "@xonovex/script-moon-common";
import {bumpVersion, updateDependencyVersions} from "./bump.js";

interface DependentUpdate {
  readonly path: string;
  readonly pkg: PackageJson;
  readonly depsChanged: boolean;
  readonly versionBumped: boolean;
  readonly oldVersion: string | undefined;
  readonly newVersion: string | undefined;
}

export const updateDependent = (
  depPkg: PackageJson,
  depPkgPath: string,
  packageName: string,
  newVersion: string,
  getGitVersion: () => string | undefined,
): DependentUpdate => {
  let depsChanged = false;
  if (updateDependencyVersions(depPkg.dependencies, packageName, newVersion))
    depsChanged = true;
  if (updateDependencyVersions(depPkg.devDependencies, packageName, newVersion))
    depsChanged = true;
  if (
    updateDependencyVersions(depPkg.peerDependencies, packageName, newVersion)
  )
    depsChanged = true;

  if (!depsChanged) {
    return {
      path: depPkgPath,
      pkg: depPkg,
      depsChanged: false,
      versionBumped: false,
      oldVersion: undefined,
      newVersion: undefined,
    };
  }

  // Patch-bump the dependent's own version if not already bumped
  if (depPkg.version && !depPkg.private) {
    const gitVersion = getGitVersion();
    if (gitVersion === depPkg.version) {
      const oldVersion = depPkg.version;
      depPkg.version = bumpVersion(oldVersion, "patch");
      return {
        path: depPkgPath,
        pkg: depPkg,
        depsChanged: true,
        versionBumped: true,
        oldVersion,
        newVersion: depPkg.version,
      };
    }
  }

  return {
    path: depPkgPath,
    pkg: depPkg,
    depsChanged: true,
    versionBumped: false,
    oldVersion: undefined,
    newVersion: undefined,
  };
};

export type {DependentUpdate};
