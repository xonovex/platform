import type {GitRepository} from "@xonovex/script-moon-common/git";
import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import type {DepUpdate} from "./changelog.js";
import {readHeadPackage} from "./git.js";

const getWorkspaceDeps = (pkg: PackageJson): ReadonlyMap<string, string> => {
  const deps = new Map<string, string>();
  const allDeps = {
    ...pkg.dependencies,
    ...pkg.devDependencies,
    ...pkg.optionalDependencies,
  };
  for (const [name, version] of Object.entries(allDeps)) {
    if (version && name.startsWith("@xonovex/")) {
      deps.set(name, version);
    }
  }
  return deps;
};

// currentPkg is passed rather than read from pkgPath so a planned package that
// is not on disk yet is diffed against the same HEAD baseline.
const detectDepUpdates = (
  rootDir: string,
  pkgPath: string,
  currentPkg: PackageJson,
  git: Pick<GitRepository, "readFile">,
): readonly DepUpdate[] => {
  const headPkg = readHeadPackage(rootDir, pkgPath, git);
  if (headPkg === undefined) return [];
  const oldDeps = getWorkspaceDeps(headPkg);

  const updates: DepUpdate[] = [];
  for (const [name, version] of getWorkspaceDeps(currentPkg)) {
    const oldVersion = oldDeps.get(name);
    if (oldVersion && oldVersion !== version) {
      updates.push({name, version});
    }
  }
  return updates;
};

export {detectDepUpdates, getWorkspaceDeps};
