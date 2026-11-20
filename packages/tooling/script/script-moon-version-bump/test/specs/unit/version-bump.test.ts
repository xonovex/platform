import {basename, dirname, join, relative} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import type {FileChange} from "../../../src/file-transaction.js";
import type {VersionGit} from "../../../src/git.js";
import {main, type VersionBumpDependencies} from "../../../src/version-bump.js";

const ROOT = "/repo";

const packagePath = (name: string): string =>
  join(ROOT, "packages", name, "package.json");

// A workspace is just a tree holding .moon plus package manifests: main finds the
// root by walking up for .moon, and every port is supplied, so no repository, no
// moon process, and no disk are involved.
const workspace = (packages: Readonly<Record<string, unknown>>): FileSystem =>
  memoryFileSystem({
    directories: [join(ROOT, ".moon")],
    files: Object.fromEntries(
      Object.entries(packages).map(([name, manifest]) => [
        packagePath(name),
        `${JSON.stringify(manifest, null, 2)}\n`,
      ]),
    ),
  });

// Answers the HEAD reads getGitVersion makes. A path without a version is
// absent at HEAD, which is how a never-committed package looks. Each committed
// package carries a release tag, so a dependent counts as released. The history
// reads throw, so a case that plans a changelog must supply its own.
const gitAtHead = (versions: Readonly<Record<string, string>>): VersionGit => ({
  readFile: (commit, path) => {
    const version = versions[path];
    return commit === "HEAD" && version !== undefined
      ? `${JSON.stringify({name: path, version})}\n`
      : undefined;
  },
  pathHistory: () => {
    throw new Error("unexpected history read");
  },
  hasParent: () => {
    throw new Error("unexpected history read");
  },
  commitsSince: () => {
    throw new Error("unexpected history read");
  },
  tags: () =>
    Object.entries(versions).map(
      ([path, version]) => `${basename(dirname(path))}-v${version}`,
    ),
});

// applyChanges installs into the same tree the reads come from, so a case can
// assert on what a run wrote without a disk.
const dependencies = (
  git: VersionGit,
  paths: readonly string[],
  fs: FileSystem,
  output = memoryOutput(),
): VersionBumpDependencies => ({
  openRepository: () => git,
  fs,
  output,
  listProjects: () =>
    paths.map((path) => ({
      id: basename(dirname(path)),
      source: relative(ROOT, dirname(path)),
    })),
  applyChanges: (changes: readonly FileChange[]) => {
    for (const change of changes) fs.writeFile(change.path, change.content);
  },
});

describe("main", () => {
  it("prints the help text and plans nothing on --help", () => {
    const fs = workspace({core: {name: "@xonovex/core", version: "1.2.3"}});
    const output = memoryOutput();

    const code = main(
      ["--help"],
      join(ROOT, "packages", "core"),
      dependencies(gitAtHead({}), [packagePath("core")], fs, output),
    );

    expect(code).toBe(0);
    expect(output.stdoutLines()[0]).toContain("moon-version-bump");
    expect(JSON.parse(fs.readText(packagePath("core")))).toMatchObject({
      version: "1.2.3",
    });
  });

  it("bumps a package that is unchanged at HEAD and writes the manifest", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");

    const code = main(
      ["--type", "minor", "--no-changelog"],
      join(ROOT, "packages", "core"),
      dependencies(
        gitAtHead({"packages/core/package.json": "1.2.3"}),
        [corePath],
        fs,
      ),
    );

    expect(code).toBe(0);
    expect(JSON.parse(fs.readText(corePath))).toMatchObject({
      version: "1.3.0",
    });
  });

  it("leaves an already-bumped package alone", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "2.0.0"},
    });
    const corePath = packagePath("core");

    const code = main(
      ["--no-changelog"],
      join(ROOT, "packages", "core"),
      dependencies(
        // HEAD still holds 1.0.0, so the working tree is ahead already.
        gitAtHead({"packages/core/package.json": "1.0.0"}),
        [corePath],
        fs,
      ),
    );

    expect(code).toBe(0);
    expect(JSON.parse(fs.readText(corePath))).toMatchObject({
      version: "2.0.0",
    });
  });

  it("writes nothing under --dry-run", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");
    const before = fs.readText(corePath);

    const code = main(
      ["--dry-run", "--no-changelog"],
      join(ROOT, "packages", "core"),
      dependencies(
        gitAtHead({"packages/core/package.json": "1.2.3"}),
        [corePath],
        fs,
      ),
    );

    expect(code).toBe(0);
    expect(fs.readText(corePath)).toBe(before);
  });

  it("applies an exact version instead of a bump type", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");

    main(
      ["--exact", "4.5.6", "--no-changelog"],
      join(ROOT, "packages", "core"),
      dependencies(
        gitAtHead({"packages/core/package.json": "1.2.3"}),
        [corePath],
        fs,
      ),
    );

    expect(JSON.parse(fs.readText(corePath))).toMatchObject({
      version: "4.5.6",
    });
  });

  it("updates a dependent that pins the bumped package exactly", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
      consumer: {
        name: "@xonovex/consumer",
        version: "0.1.0",
        dependencies: {"@xonovex/core": "1.2.3"},
      },
    });
    const corePath = packagePath("core");
    const consumerPath = packagePath("consumer");

    main(
      ["--type", "minor", "--no-changelog"],
      join(ROOT, "packages", "core"),
      dependencies(
        gitAtHead({
          "packages/core/package.json": "1.2.3",
          "packages/consumer/package.json": "0.1.0",
        }),
        [corePath, consumerPath],
        fs,
      ),
    );

    expect(JSON.parse(fs.readText(consumerPath))).toMatchObject({
      dependencies: {"@xonovex/core": "1.3.0"},
    });
  });

  it("skips dependent updates when asked", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
      consumer: {
        name: "@xonovex/consumer",
        version: "0.1.0",
        dependencies: {"@xonovex/core": "1.2.3"},
      },
    });
    const corePath = packagePath("core");
    const consumerPath = packagePath("consumer");

    main(
      ["--type", "minor", "--no-changelog", "--no-dependents"],
      join(ROOT, "packages", "core"),
      dependencies(
        gitAtHead({"packages/core/package.json": "1.2.3"}),
        [corePath, consumerPath],
        fs,
      ),
    );

    expect(JSON.parse(fs.readText(consumerPath))).toMatchObject({
      dependencies: {"@xonovex/core": "1.2.3"},
    });
  });

  it("rejects a malformed exact version", () => {
    const fs = workspace({core: {name: "@xonovex/core", version: "1.2.3"}});

    expect(() =>
      main(
        ["--exact", "not-a-version"],
        join(ROOT, "packages", "core"),
        dependencies(gitAtHead({}), [], fs),
      ),
    ).toThrow("invalid exact version");
  });

  it("rejects an unknown bump type", () => {
    const fs = workspace({core: {name: "@xonovex/core", version: "1.2.3"}});

    expect(() =>
      main(
        ["--type", "sideways"],
        join(ROOT, "packages", "core"),
        dependencies(gitAtHead({}), [], fs),
      ),
    ).toThrow("invalid bump type");
  });

  it("rejects a directory with no package.json", () => {
    const fs = workspace({});

    expect(() => main([], ROOT, dependencies(gitAtHead({}), [], fs))).toThrow(
      "no package.json found",
    );
  });

  it("rejects a manifest missing a name or version", () => {
    const fs = workspace({core: {description: "no identity"}});

    expect(() =>
      main(
        [],
        join(ROOT, "packages", "core"),
        dependencies(gitAtHead({}), [packagePath("core")], fs),
      ),
    ).toThrow("missing a name or version");
  });

  it("writes a changelog of the chosen commit types since the given base", () => {
    const fs = workspace({core: {name: "@xonovex/core", version: "1.2.3"}});
    const corePath = packagePath("core");
    const since: (string | undefined)[] = [];

    main(
      ["--type", "minor", "--git-base", "v1.2.3", "--include-types", "chore"],
      join(ROOT, "packages", "core"),
      dependencies(
        {
          ...gitAtHead({"packages/core/package.json": "1.2.3"}),
          commitsSince: (ref) => {
            since.push(ref);
            return [
              {hash: "aaa1111", author: "Ada", message: "feat: add a thing"},
              {hash: "bbb2222", author: "Ada", message: "chore: tidy a thing"},
            ];
          },
        },
        [corePath],
        fs,
      ),
    );

    expect(since).toEqual(["v1.2.3"]);
    const changelog = fs.readText(
      join(ROOT, "packages", "core", "CHANGELOG.md"),
    );
    expect(changelog).toContain("## 1.3.0");
    expect(changelog).toContain("tidy a thing");
    expect(changelog).not.toContain("add a thing");
  });

  it("moves the named packages to one version under --lockstep", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
      extra: {name: "@xonovex/extra", version: "1.0.0"},
    });

    const code = main(
      ["--lockstep", "@xonovex/core,@xonovex/extra", "--no-changelog"],
      ROOT,
      dependencies(
        gitAtHead({
          "packages/core/package.json": "1.2.3",
          "packages/extra/package.json": "1.0.0",
        }),
        [packagePath("core"), packagePath("extra")],
        fs,
      ),
    );

    expect(code).toBe(0);
    expect(JSON.parse(fs.readText(packagePath("core")))).toMatchObject({
      version: "1.2.4",
    });
    expect(JSON.parse(fs.readText(packagePath("extra")))).toMatchObject({
      version: "1.2.4",
    });
  });
});
