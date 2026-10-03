import {mkdirSync, readFileSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/version-bump.js";
import {
  commitAll,
  gitRepositories,
  temporaryDirectories,
} from "../../util/git-repository.js";

const repository = gitRepositories();
const temporaryDirectory = temporaryDirectories();
const defaultPackage = {
  name: "@xonovex/example",
  version: "1.0.0",
} as const;

const createRepository = (
  pkg: Readonly<Record<string, unknown>> = defaultPackage,
): {
  readonly root: string;
  readonly packageDirectory: string;
  readonly packagePath: string;
} => {
  const root = repository("version-command-");
  mkdirSync(join(root, ".moon"));
  const packageDirectory = join(root, "packages", "example");
  const packagePath = join(packageDirectory, "package.json");
  mkdirSync(packageDirectory, {recursive: true});
  writeFileSync(packagePath, `${JSON.stringify(pkg, null, 2)}\n`);
  commitAll(root, "feat: initial");
  return {root, packageDirectory, packagePath};
};

describe("version bump command", () => {
  it("bumps a package through the atomic write path", () => {
    const {packageDirectory, packagePath} = createRepository();

    const exitCode = main(
      ["--no-dependents", "--no-changelog"],
      packageDirectory,
    );

    expect(exitCode).toBe(0);
    expect(readPkg(packagePath).version).toBe("1.0.1");
  });

  it("previews a requested version without changing the package", () => {
    const {packageDirectory, packagePath} = createRepository();

    const exitCode = main(
      [
        "--dry-run",
        "--no-dependents",
        "--no-changelog",
        "--exact",
        "2.0.0-beta.1",
      ],
      packageDirectory,
    );

    expect(exitCode).toBe(0);
    expect(readPkg(packagePath).version).toBe("1.0.0");
  });

  it("updates the package and changelog in one command", () => {
    const {root, packageDirectory, packagePath} = createRepository();
    writeFileSync(
      packagePath,
      `${JSON.stringify({name: "@xonovex/example", version: "1.1.0"}, null, 2)}\n`,
    );
    commitAll(root, "chore: release 1.1.0");

    const exitCode = main(["--no-dependents"], packageDirectory);

    expect(exitCode).toBe(0);
    expect(readPkg(packagePath).version).toBe("1.1.1");
    expect(
      readFileSync(join(packageDirectory, "CHANGELOG.md"), "utf8"),
    ).toContain("## 1.1.1");
  });

  it("does not bump an uncommitted version a second time", () => {
    const {packageDirectory, packagePath} = createRepository();
    main(["--no-dependents", "--no-changelog"], packageDirectory);

    const exitCode = main(
      ["--no-dependents", "--no-changelog"],
      packageDirectory,
    );

    expect(exitCode).toBe(0);
    expect(readPkg(packagePath).version).toBe("1.0.1");
  });

  it("rejects invalid requests and package metadata", () => {
    const {packageDirectory} = createRepository();
    expect(() =>
      main(
        ["--no-dependents", "--no-changelog", "--exact", "invalid"],
        packageDirectory,
      ),
    ).toThrow("invalid exact version");
    expect(() =>
      main(["--no-dependents", "--no-changelog", "invalid"], packageDirectory),
    ).toThrow("invalid bump type");

    const missingPackage = temporaryDirectory("version-command-empty-");
    expect(() => main([], missingPackage)).toThrow("no package.json found");

    const invalidPackage = createRepository({version: "1.0.0"});
    expect(() => main([], invalidPackage.packageDirectory)).toThrow(
      "missing a name or version",
    );
  });
});
