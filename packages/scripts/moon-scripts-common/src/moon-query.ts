import {execSync} from "node:child_process";
import {existsSync} from "node:fs";
import {join} from "node:path";

export interface MoonProject {
  id: string;
  source: string;
}

export const queryMoonProjects = (rootDir: string): readonly MoonProject[] => {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const output = execSync("npx moon query projects", {
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
    cwd: rootDir,
  });
  return (JSON.parse(output) as {projects: readonly MoonProject[]}).projects;
};

export const findAllPackageJsonPaths = (
  rootDir: string,
): readonly string[] =>
  queryMoonProjects(rootDir)
    .map((p) => join(rootDir, p.source, "package.json"))
    .filter((p) => existsSync(p));
