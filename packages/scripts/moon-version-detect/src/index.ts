#!/usr/bin/env node
import {execSync} from "node:child_process";
import {existsSync, readFileSync} from "node:fs";
import {dirname, join} from "node:path";
import {fileURLToPath} from "node:url";

const findWorkspaceRoot = (start: string): string => {
  let dir = start;
  while (dir !== dirname(dir)) {
    if (existsSync(join(dir, ".moon"))) return dir;
    dir = dirname(dir);
  }
  throw new Error("Could not find workspace root");
};
const ROOT_DIR = findWorkspaceRoot(dirname(fileURLToPath(import.meta.url)));

interface PackageJson {
  name?: string;
  version?: string;
  private?: boolean;
}

interface MoonProject {
  id: string;
  source: string;
}

const ref = process.argv[2] ?? "HEAD~1";

const moonOutput = JSON.parse(
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  execSync("npx moon query projects", {
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
    cwd: ROOT_DIR,
  }),
) as {projects: readonly MoonProject[]};

const projects: string[] = [];

for (const p of moonOutput.projects) {
  try {
    const pkgPath = join(ROOT_DIR, p.source, "package.json");
    if (!existsSync(pkgPath)) continue;

    const pkg = JSON.parse(readFileSync(pkgPath, "utf8")) as PackageJson;
    if (!pkg.name || pkg.private) continue;

    const oldPkg = JSON.parse(
       
      execSync(`git show ${ref}:${p.source}/package.json`, {
        encoding: "utf8",
        cwd: ROOT_DIR,
      }),
    ) as PackageJson;

    if (pkg.version !== oldPkg.version) {
      projects.push(p.id);
    }
  } catch {
    // Package didn't exist at ref — skip
  }
}

console.log(JSON.stringify(projects));
