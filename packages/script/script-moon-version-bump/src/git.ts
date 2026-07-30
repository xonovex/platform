import {execFileSync} from "node:child_process";
import {relative} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {parsePackageJson} from "@xonovex/script-moon-common/package-json";

export const getGitVersion = (
  rootDir: string,
  pkgPath: string,
): string | undefined => {
  const rel = relative(rootDir, pkgPath);
  const listed = execFileSync(
    resolveExecutable("git"),
    ["ls-tree", "-z", "--name-only", "HEAD", "--", rel],
    {cwd: rootDir, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"]},
  );
  if (listed.length === 0) return undefined;

  const content = execFileSync(
    resolveExecutable("git"),
    ["show", `HEAD:${rel}`],
    {
      cwd: rootDir,
      encoding: "utf8",
      stdio: ["pipe", "pipe", "pipe"],
    },
  );
  return parsePackageJson(content, `HEAD:${rel}`).version;
};
