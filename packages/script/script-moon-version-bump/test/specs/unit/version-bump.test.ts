import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import type {GitRunner} from "../../../src/git.js";
import {main, type VersionBumpDependencies} from "../../../src/version-bump.js";

const directories: string[] = [];

afterEach(() => {
  vi.restoreAllMocks();
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
});

const writeJson = (path: string, value: unknown): void => {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`);
};

// A workspace is just a directory holding .moon plus package manifests: main
// finds the root by walking up for .moon, and both ports are supplied, so no
// repository and no moon process are involved.
const workspace = (
  packages: Readonly<Record<string, unknown>>,
): {root: string; packagePath: (name: string) => string} => {
  const root = mkdtempSync(join(tmpdir(), "version-bump-unit-"));
  directories.push(root);
  mkdirSync(join(root, ".moon"));
  for (const [name, manifest] of Object.entries(packages)) {
    mkdirSync(join(root, "packages", name), {recursive: true});
    writeJson(join(root, "packages", name, "package.json"), manifest);
  }
  return {
    root,
    packagePath: (name) => join(root, "packages", name, "package.json"),
  };
};

// Answers the two reads getGitVersion makes. An empty ls-tree means the path is
// absent at HEAD, which is how a never-committed package looks.
const gitAtHead = (versions: Readonly<Record<string, string>>): GitRunner => {
  return (args) => {
    const [command] = args;
    if (command === "ls-tree") {
      const path = args.at(-1) ?? "";
      return versions[path] === undefined ? "" : `${path}\0`;
    }
    if (command === "show") {
      const ref = args[1] ?? "";
      const path = ref.slice("HEAD:".length);
      return `${JSON.stringify({name: path, version: versions[path]})}\n`;
    }
    throw new Error(`unexpected git command ${String(command)}`);
  };
};

const dependencies = (
  git: GitRunner,
  paths: readonly string[],
): VersionBumpDependencies => ({git, listPackagePaths: () => paths});

describe("main", () => {
  it("bumps a package that is unchanged at HEAD and writes the manifest", () => {
    const {root, packagePath} = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");
    vi.spyOn(console, "error").mockImplementation(() => {});

    const code = main(
      ["--type", "minor", "--no-changelog"],
      join(root, "packages", "core"),
      dependencies(gitAtHead({"packages/core/package.json": "1.2.3"}), [
        corePath,
      ]),
    );

    expect(code).toBe(0);
    expect(JSON.parse(readFileSync(corePath, "utf8"))).toMatchObject({
      version: "1.3.0",
    });
  });

  it("leaves an already-bumped package alone", () => {
    const {root, packagePath} = workspace({
      core: {name: "@xonovex/core", version: "2.0.0"},
    });
    const corePath = packagePath("core");
    vi.spyOn(console, "error").mockImplementation(() => {});

    const code = main(
      ["--no-changelog"],
      join(root, "packages", "core"),
      dependencies(
        // HEAD still holds 1.0.0, so the working tree is ahead already.
        gitAtHead({"packages/core/package.json": "1.0.0"}),
        [corePath],
      ),
    );

    expect(code).toBe(0);
    expect(JSON.parse(readFileSync(corePath, "utf8"))).toMatchObject({
      version: "2.0.0",
    });
  });

  it("writes nothing under --dry-run", () => {
    const {root, packagePath} = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");
    const before = readFileSync(corePath, "utf8");
    vi.spyOn(console, "error").mockImplementation(() => {});

    const code = main(
      ["--dry-run", "--no-changelog"],
      join(root, "packages", "core"),
      dependencies(gitAtHead({"packages/core/package.json": "1.2.3"}), [
        corePath,
      ]),
    );

    expect(code).toBe(0);
    expect(readFileSync(corePath, "utf8")).toBe(before);
  });

  it("applies an exact version instead of a bump type", () => {
    const {root, packagePath} = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");
    vi.spyOn(console, "error").mockImplementation(() => {});

    main(
      ["--exact", "4.5.6", "--no-changelog"],
      join(root, "packages", "core"),
      dependencies(gitAtHead({"packages/core/package.json": "1.2.3"}), [
        corePath,
      ]),
    );

    expect(JSON.parse(readFileSync(corePath, "utf8"))).toMatchObject({
      version: "4.5.6",
    });
  });

  it("updates a dependent that pins the bumped package exactly", () => {
    const {root, packagePath} = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
      consumer: {
        name: "@xonovex/consumer",
        version: "0.1.0",
        dependencies: {"@xonovex/core": "1.2.3"},
      },
    });
    const corePath = packagePath("core");
    const consumerPath = packagePath("consumer");
    vi.spyOn(console, "error").mockImplementation(() => {});

    main(
      ["--type", "minor", "--no-changelog"],
      join(root, "packages", "core"),
      dependencies(
        gitAtHead({
          "packages/core/package.json": "1.2.3",
          "packages/consumer/package.json": "0.1.0",
        }),
        [corePath, consumerPath],
      ),
    );

    expect(JSON.parse(readFileSync(consumerPath, "utf8"))).toMatchObject({
      dependencies: {"@xonovex/core": "1.3.0"},
    });
  });

  it("skips dependent updates when asked", () => {
    const {root, packagePath} = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
      consumer: {
        name: "@xonovex/consumer",
        version: "0.1.0",
        dependencies: {"@xonovex/core": "1.2.3"},
      },
    });
    const corePath = packagePath("core");
    const consumerPath = packagePath("consumer");
    vi.spyOn(console, "error").mockImplementation(() => {});

    main(
      ["--type", "minor", "--no-changelog", "--no-dependents"],
      join(root, "packages", "core"),
      dependencies(gitAtHead({"packages/core/package.json": "1.2.3"}), [
        corePath,
        consumerPath,
      ]),
    );

    expect(JSON.parse(readFileSync(consumerPath, "utf8"))).toMatchObject({
      dependencies: {"@xonovex/core": "1.2.3"},
    });
  });

  it("rejects a malformed exact version", () => {
    const {root} = workspace({core: {name: "@xonovex/core", version: "1.2.3"}});
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() =>
      main(
        ["--exact", "not-a-version"],
        join(root, "packages", "core"),
        dependencies(gitAtHead({}), []),
      ),
    ).toThrow("invalid exact version");
  });

  it("rejects an unknown bump type", () => {
    const {root} = workspace({core: {name: "@xonovex/core", version: "1.2.3"}});
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() =>
      main(
        ["--type", "sideways"],
        join(root, "packages", "core"),
        dependencies(gitAtHead({}), []),
      ),
    ).toThrow("invalid bump type");
  });

  it("rejects a directory with no package.json", () => {
    const {root} = workspace({});
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => main([], root, dependencies(gitAtHead({}), []))).toThrow(
      "no package.json found",
    );
  });

  it("rejects a manifest missing a name or version", () => {
    const {root, packagePath} = workspace({core: {description: "no identity"}});
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() =>
      main(
        [],
        join(root, "packages", "core"),
        dependencies(gitAtHead({}), [packagePath("core")]),
      ),
    ).toThrow("missing a name or version");
  });
});
