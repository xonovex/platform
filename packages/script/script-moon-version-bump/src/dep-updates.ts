import {execFileSync} from "node:child_process";
import {readFileSync} from "node:fs";
import {relative} from "node:path";
import type {PackageJson} from "@xonovex/script-moon-common";
import type {DepUpdate} from "./changelog.js";

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

const detectDepUpdates = (
  rootDir: string,
  pkgPath: string,
): readonly DepUpdate[] => {
  const currentPkg = JSON.parse(readFileSync(pkgPath, "utf8")) as PackageJson;
  const currentDeps = getWorkspaceDeps(currentPkg);

  const relativePath = relative(rootDir, pkgPath);
  const listed = execFileSync(
    // eslint-disable-next-line sonarjs/no-os-command-from-path
    "git",
    ["ls-tree", "-z", "--name-only", "HEAD", "--", relativePath],
    {cwd: rootDir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"]},
  );
  if (listed.length === 0) return [];

  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const oldPkgJson = execFileSync("git", ["show", `HEAD:${relativePath}`], {
    cwd: rootDir,
    encoding: "utf8",
  });
  const oldDeps = getWorkspaceDeps(JSON.parse(oldPkgJson) as PackageJson);

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
