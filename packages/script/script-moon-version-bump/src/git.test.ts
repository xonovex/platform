import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {describe, expect, it} from "vitest";
import {detectDepUpdates} from "./dep-updates.js";
import {getGitVersion} from "./git.js";

const runGit = (cwd: string, args: readonly string[]): void => {
  execFileSync(resolveExecutable("git"), [...args], {cwd, stdio: "ignore"});
};

const createRepository = (): {
  readonly root: string;
  readonly pkgPath: string;
} => {
  const root = mkdtempSync(join(tmpdir(), "version-bump-git-"));
  const packageDir = join(root, "packages", "example");
  const pkgPath = join(packageDir, "package.json");
  mkdirSync(packageDir, {recursive: true});
  writeFileSync(
    pkgPath,
    JSON.stringify({
      name: "@xonovex/example",
      version: "1.0.0",
      dependencies: {"@xonovex/shared": "workspace:1.0.0"},
    }),
  );
  runGit(root, ["init", "--quiet"]);
  runGit(root, ["add", "packages/example/package.json"]);
  runGit(root, [
    "-c",
    "user.name=Xonovex Test",
    "-c",
    "user.email=test@xonovex.invalid",
    "commit",
    "--quiet",
    "-m",
    "test: initial package",
  ]);
  return {root, pkgPath};
};

describe("Git package history", () => {
  it("reads committed versions and dependency updates without shell parsing", () => {
    const {root, pkgPath} = createRepository();
    writeFileSync(
      pkgPath,
      JSON.stringify({
        name: "@xonovex/example",
        version: "2.0.0",
        dependencies: {"@xonovex/shared": "workspace:2.0.0"},
      }),
    );

    expect(getGitVersion(root, pkgPath)).toBe("1.0.0");
    expect(detectDepUpdates(root, pkgPath)).toEqual([
      {name: "@xonovex/shared", version: "workspace:2.0.0"},
    ]);
  });

  it("distinguishes a new package from a Git failure", () => {
    const {root} = createRepository();
    const newPackageDir = join(root, "packages", "new-package");
    const newPackagePath = join(newPackageDir, "package.json");
    mkdirSync(newPackageDir, {recursive: true});
    writeFileSync(
      newPackagePath,
      JSON.stringify({name: "@xonovex/new-package", version: "1.0.0"}),
    );

    expect(getGitVersion(root, newPackagePath)).toBeUndefined();
    expect(detectDepUpdates(root, newPackagePath)).toEqual([]);

    const nonRepository = mkdtempSync(join(tmpdir(), "version-bump-no-git-"));
    expect(() => getGitVersion(nonRepository, newPackagePath)).toThrow();
  });
});
