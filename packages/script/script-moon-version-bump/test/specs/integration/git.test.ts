import {mkdirSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {describe, expect, it} from "vitest";
import {detectDepUpdates} from "../../../src/dep-updates.js";
import {getGitVersion} from "../../../src/git.js";
import {
  commitAll,
  gitRepositories,
  temporaryDirectories,
} from "../../util/git-repository.js";

const repository = gitRepositories();
const temporaryDirectory = temporaryDirectories();

const createRepository = (): {
  readonly root: string;
  readonly pkgPath: string;
} => {
  const root = repository("version-bump-git-");
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
  commitAll(root, "test: initial package");
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
    expect(detectDepUpdates(root, pkgPath, readPkg(pkgPath))).toEqual([
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
    expect(
      detectDepUpdates(root, newPackagePath, readPkg(newPackagePath)),
    ).toEqual([]);

    const nonRepository = temporaryDirectory("version-bump-no-git-");
    expect(() => getGitVersion(nonRepository, newPackagePath)).toThrow();
  });
});
