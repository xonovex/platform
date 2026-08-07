import {execFileSync} from "node:child_process";
import {relative} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {parsePackageJson} from "@xonovex/script-moon-common/package-json";

// Every git read in this package goes through GitRunner, so a caller can supply
// recorded output in place of a repository. runGit is the only implementation
// that spawns anything.
export type GitRunner = (args: readonly string[], rootDir: string) => string;

export const runGit: GitRunner = (args, rootDir) =>
  execFileSync(resolveExecutable("git"), [...args], {
    cwd: rootDir,
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });

export const getGitVersion = (
  rootDir: string,
  pkgPath: string,
  git: GitRunner = runGit,
): string | undefined => {
  const rel = relative(rootDir, pkgPath);
  const listed = git(
    ["ls-tree", "-z", "--name-only", "HEAD", "--", rel],
    rootDir,
  );
  if (listed.length === 0) return undefined;

  const content = git(["show", `HEAD:${rel}`], rootDir);
  return parsePackageJson(content, `HEAD:${rel}`).version;
};
