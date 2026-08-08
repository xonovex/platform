import {execFileSync} from "node:child_process";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import type {MoonProject} from "@xonovex/script-moon-common/moon-query";

interface PackageJson {
  readonly name?: unknown;
  readonly version?: unknown;
  readonly private?: unknown;
}

interface GitReader {
  readonly readFile: (commit: string, path: string) => string | undefined;
}

const parsePackage = (text: string, path: string): PackageJson => {
  const value: unknown = JSON.parse(text);
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error(`${path} must contain a JSON object`);
  }
  return value;
};

const publishableVersion = (
  pkg: PackageJson,
  path: string,
): string | undefined => {
  if (pkg.private === true) return undefined;
  if (typeof pkg.name !== "string" || pkg.name.length === 0) {
    throw new Error(`${path} must declare a non-empty package name`);
  }
  if (typeof pkg.version !== "string" || pkg.version.length === 0) {
    throw new Error(`${path} must declare a non-empty package version`);
  }
  return pkg.version;
};

export const resolveGitRef = (rootDir: string, ref: string): string =>
  execFileSync(
    resolveExecutable("git"),
    ["rev-parse", "--verify", "--end-of-options", `${ref}^{commit}`],
    {cwd: rootDir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"]},
  ).trim();

export const createGitReader = (rootDir: string): GitReader => ({
  readFile: (commit, path) => {
    const listed = execFileSync(
      resolveExecutable("git"),
      ["ls-tree", "-z", "--name-only", commit, "--", path],
      {cwd: rootDir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"]},
    );
    if (listed.length === 0) return;
    return execFileSync(
      resolveExecutable("git"),
      ["show", `${commit}:${path}`],
      {
        cwd: rootDir,
        encoding: "utf8",
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
  },
});

export const detectVersionChanges = (
  rootDir: string,
  commit: string,
  projects: readonly MoonProject[],
  git: GitReader,
  fs: FileSystem = nodeFileSystem,
): readonly string[] => {
  const changed: string[] = [];

  for (const project of projects) {
    const relativePath = `${project.source}/package.json`;
    const currentPath = join(rootDir, relativePath);
    if (!fs.isFile(currentPath)) continue;

    const current = parsePackage(fs.readText(currentPath), currentPath);
    const currentVersion = publishableVersion(current, currentPath);
    if (currentVersion === undefined) continue;

    const previousText = git.readFile(commit, relativePath);
    if (previousText === undefined) continue;
    const previous = parsePackage(previousText, `${commit}:${relativePath}`);
    const previousVersion = publishableVersion(
      previous,
      `${commit}:${relativePath}`,
    );
    if (previousVersion === undefined) {
      changed.push(project.id);
      continue;
    }

    if (currentVersion !== previousVersion) changed.push(project.id);
  }

  return changed;
};

export type {GitReader};
