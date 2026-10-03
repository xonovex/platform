import {execFileSync} from "node:child_process";
import {relative} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  parsePackageJson,
  type PackageJson,
} from "@xonovex/script-moon-common/package-json";
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

// currentPkg is passed rather than read from pkgPath so a planned package that
// is not on disk yet is diffed against the same HEAD baseline.
const detectDepUpdates = (
  rootDir: string,
  pkgPath: string,
  currentPkg: PackageJson,
): readonly DepUpdate[] => {
  const currentDeps = getWorkspaceDeps(currentPkg);

  const relativePath = relative(rootDir, pkgPath);
  const listed = execFileSync(
    resolveExecutable("git"),
    ["ls-tree", "-z", "--name-only", "HEAD", "--", relativePath],
    {cwd: rootDir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"]},
  );
  if (listed.length === 0) return [];

  const oldPkgJson = execFileSync(
    resolveExecutable("git"),
    ["show", `HEAD:${relativePath}`],
    {
      cwd: rootDir,
      encoding: "utf8",
    },
  );
  const oldDeps = getWorkspaceDeps(
    parsePackageJson(oldPkgJson, `HEAD:${relativePath}`),
  );

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
