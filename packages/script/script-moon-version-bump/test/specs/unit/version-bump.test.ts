import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {afterEach, describe, expect, it, vi} from "vitest";
import type {FileChange} from "../../../src/file-transaction.js";
import type {GitRunner} from "../../../src/git.js";
import {main, type VersionBumpDependencies} from "../../../src/version-bump.js";

afterEach(() => {
  vi.restoreAllMocks();
});

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

// applyChanges installs into the same tree the reads come from, so a case can
// assert on what a run wrote without a disk.
const dependencies = (
  git: GitRunner,
  paths: readonly string[],
  fs: FileSystem,
): VersionBumpDependencies => ({
  git,
  fs,
  listPackagePaths: () => paths,
  applyChanges: (changes: readonly FileChange[]) => {
    for (const change of changes) fs.writeFile(change.path, change.content);
  },
});

describe("main", () => {
  it("bumps a package that is unchanged at HEAD and writes the manifest", () => {
    const fs = workspace({
      core: {name: "@xonovex/core", version: "1.2.3"},
    });
    const corePath = packagePath("core");
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

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
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => main([], ROOT, dependencies(gitAtHead({}), [], fs))).toThrow(
      "no package.json found",
    );
  });

  it("rejects a manifest missing a name or version", () => {
    const fs = workspace({core: {description: "no identity"}});
    vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() =>
      main(
        [],
        join(ROOT, "packages", "core"),
        dependencies(gitAtHead({}), [packagePath("core")], fs),
      ),
    ).toThrow("missing a name or version");
  });
});
