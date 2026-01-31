import {execSync} from "node:child_process";
import {readFileSync} from "node:fs";
import type {PackageJson} from "@xonovex/moon-scripts-common";
import type {DepUpdate} from "./changelog.js";

const getWorkspaceDeps = (
  pkg: PackageJson,
): ReadonlyMap<string, string> => {
  const deps = new Map<string, string>();
  const allDeps = {...pkg.dependencies, ...pkg.devDependencies};
  for (const [name, version] of Object.entries(allDeps)) {
    if (version && name.startsWith("@xonovex/")) {
      deps.set(name, version);
    }
  }
  return deps;
};

const detectDepUpdates = (rootDir: string, pkgPath: string): readonly DepUpdate[] => {
  const currentPkg = JSON.parse(readFileSync(pkgPath, "utf8")) as PackageJson;
  const currentDeps = getWorkspaceDeps(currentPkg);

  let oldDeps: ReadonlyMap<string, string>;
  try {
    const relativePath = pkgPath.replace(rootDir + "/", "");
    const oldPkgJson = execSync(`git show HEAD:${relativePath}`, {
      cwd: rootDir,
      encoding: "utf8",
    });
    oldDeps = getWorkspaceDeps(JSON.parse(oldPkgJson) as PackageJson);
  } catch {
    // File didn't exist in HEAD, all deps are new
    return [];
  }

  const updates: DepUpdate[] = [];
  for (const [name, version] of currentDeps) {
    const oldVersion = oldDeps.get(name);
    if (oldVersion && oldVersion !== version) {
      updates.push({name, version});
    }
  }
  return updates;
};

export {detectDepUpdates, getWorkspaceDeps};
