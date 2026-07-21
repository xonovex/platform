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
  const pkg: PackageJson = {
    ...depPkg,
    ...(depPkg.dependencies === undefined
      ? {}
      : {dependencies: {...depPkg.dependencies}}),
    ...(depPkg.devDependencies === undefined
      ? {}
      : {devDependencies: {...depPkg.devDependencies}}),
    ...(depPkg.peerDependencies === undefined
      ? {}
      : {peerDependencies: {...depPkg.peerDependencies}}),
    ...(depPkg.optionalDependencies === undefined
      ? {}
      : {optionalDependencies: {...depPkg.optionalDependencies}}),
  };
  let depsChanged = false;
  if (updateDependencyVersions(pkg.dependencies, packageName, newVersion))
    depsChanged = true;
  if (updateDependencyVersions(pkg.devDependencies, packageName, newVersion))
    depsChanged = true;
  if (updateDependencyVersions(pkg.peerDependencies, packageName, newVersion))
    depsChanged = true;
  if (
    updateDependencyVersions(pkg.optionalDependencies, packageName, newVersion)
  )
    depsChanged = true;

  if (!depsChanged) {
    return {
      path: depPkgPath,
      pkg,
      depsChanged: false,
      versionBumped: false,
      oldVersion: undefined,
      newVersion: undefined,
    };
  }

  if (pkg.version && !pkg.private) {
    const gitVersion = getGitVersion();
    if (gitVersion === pkg.version) {
      const oldVersion = pkg.version;
      pkg.version = bumpVersion(oldVersion, "patch");
      return {
        path: depPkgPath,
        pkg,
        depsChanged: true,
        versionBumped: true,
        oldVersion,
        newVersion: pkg.version,
      };
    }
  }

  return {
    path: depPkgPath,
    pkg,
    depsChanged: true,
    versionBumped: false,
    oldVersion: undefined,
    newVersion: undefined,
  };
};

export type {DependentUpdate};
