import {execSync} from "node:child_process";
import {relative} from "node:path";

export const getGitVersion = (
  rootDir: string,
  pkgPath: string,
): string | undefined => {
  try {
    const rel = relative(rootDir, pkgPath);
    const content = execSync(`git show HEAD:${rel}`, {
      cwd: rootDir,
      encoding: "utf8",
      stdio: ["pipe", "pipe", "pipe"],
    });
    return (JSON.parse(content) as {version?: string}).version;
  } catch {
    return undefined;
  }
};
