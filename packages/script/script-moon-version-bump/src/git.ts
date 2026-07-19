import {execFileSync} from "node:child_process";
import {relative} from "node:path";

export const getGitVersion = (
  rootDir: string,
  pkgPath: string,
): string | undefined => {
  const rel = relative(rootDir, pkgPath);
  const listed = execFileSync(
    // eslint-disable-next-line sonarjs/no-os-command-from-path
    "git",
    ["ls-tree", "-z", "--name-only", "HEAD", "--", rel],
    {cwd: rootDir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"]},
  );
  if (listed.length === 0) return undefined;

  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const content = execFileSync("git", ["show", `HEAD:${rel}`], {
    cwd: rootDir,
    encoding: "utf8",
    stdio: ["pipe", "pipe", "pipe"],
  });
  return (JSON.parse(content) as {version?: string}).version;
};
