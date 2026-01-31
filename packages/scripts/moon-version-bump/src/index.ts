#!/usr/bin/env node
import {existsSync, readFileSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {
  findAllPackageJsonPaths,
  findWorkspaceRoot,
  logError,
  logInfo,
  logSuccess,
  readPkg,
  writePkg,
} from "@xonovex/moon-scripts-common";
import {bumpVersion, updateDependencyVersions} from "./bump.js";
import type {BumpType} from "./bump.js";

const main = (): void => {
  const cwd = process.cwd();
  const pkgPath = join(cwd, "package.json");

  if (!existsSync(pkgPath)) {
    logError("No package.json found in current directory.");
    process.exit(1);
  }

  const pkg = readPkg(pkgPath);
  if (!pkg.name || !pkg.version) {
    logError("package.json missing name or version.");
    process.exit(1);
  }

  const args = process.argv.slice(2);
  const bumpType = (args[0] ?? "patch") as BumpType;

  if (!["patch", "minor", "major"].includes(bumpType)) {
    logError(
      `Invalid bump type: ${bumpType}. Use patch, minor, or major.`,
    );
    process.exit(1);
  }

  const oldVersion = pkg.version;
  const newVersion = bumpVersion(oldVersion, bumpType);
  pkg.version = newVersion;
  writePkg(pkgPath, pkg);
  logInfo(`${pkg.name}: ${oldVersion} -> ${newVersion}`);

  // Update dependents across workspace
  const rootDir = findWorkspaceRoot(cwd);
  const allPaths = findAllPackageJsonPaths(rootDir);
  let depsUpdated = 0;

  for (const depPkgPath of allPaths) {
    if (depPkgPath === pkgPath) continue;
    const depPkg = readPkg(depPkgPath);
    let changed = false;
    if (updateDependencyVersions(depPkg.dependencies, pkg.name, newVersion))
      changed = true;
    if (updateDependencyVersions(depPkg.devDependencies, pkg.name, newVersion))
      changed = true;
    if (updateDependencyVersions(depPkg.peerDependencies, pkg.name, newVersion))
      changed = true;
    if (changed) {
      writePkg(depPkgPath, depPkg);
      depsUpdated++;
    }
  }

  // Update plugin.json if this is @xonovex/skills
  if (pkg.name === "@xonovex/skills") {
    const pluginJsonPath = join(
      rootDir,
      "packages/plugins/skills/.claude-plugin/plugin.json",
    );
    if (existsSync(pluginJsonPath)) {
      const pluginJson = JSON.parse(
        readFileSync(pluginJsonPath, "utf8"),
      ) as Record<string, unknown>;
      pluginJson.version = newVersion;
      writeFileSync(pluginJsonPath, JSON.stringify(pluginJson, null, 2) + "\n");
      logInfo(`plugin.json -> ${newVersion}`);
    }
  }

  logSuccess(
    `Bumped ${pkg.name} to ${newVersion}, updated deps in ${String(depsUpdated)} file(s).`,
  );
};

main();
