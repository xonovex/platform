import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {describe, expect, it} from "vitest";
import type {FileChange} from "../../../src/file-transaction.js";
import {
  runLockstep,
  type LockstepOptions,
} from "../../../src/lockstep-command.js";
import type {WorkspacePackage} from "../../../src/lockstep.js";
import {emptyHistory, manifestsAtHead} from "../../util/version-git.js";

const manifestPath = (directory: string): string =>
  `/repo/packages/${directory}/package.json`;

const workspacePackage = (
  directory: string,
  pkg: PackageJson,
): WorkspacePackage => ({
  path: manifestPath(directory),
  pkg,
  headVersion: pkg.version,
});

// The members depend on each other, and the consumer pins one member exactly
// and carries a release tag, so a lockstep bump moves the consumer reference and
// patch-bumps it.
const line = (): readonly WorkspacePackage[] => [
  workspacePackage("base", {
    name: "@xonovex/base",
    version: "0.1.0",
    devDependencies: {"@xonovex/prettier": "0.1.0"},
  }),
  workspacePackage("prettier", {
    name: "@xonovex/prettier",
    version: "0.1.0",
    dependencies: {"@xonovex/base": "0.1.0"},
  }),
  workspacePackage("consumer", {
    name: "@xonovex/consumer",
    version: "2.0.0",
    dependencies: {"@xonovex/base": "0.1.0"},
  }),
];

const options = (
  applied: FileChange[][],
  overrides: Partial<LockstepOptions> = {},
): LockstepOptions => ({
  rootDir: "/repo",
  packages: line(),
  names: ["@xonovex/base", "@xonovex/prettier"],
  bumpType: "minor",
  preid: undefined,
  exact: undefined,
  dryRun: false,
  noChangelog: false,
  noDependents: false,
  changelogPath: undefined,
  gitBase: "base",
  includedTypes: undefined,
  git: emptyHistory({
    readFile: manifestsAtHead(
      Object.fromEntries(
        line().map((entry) => [entry.path.slice("/repo/".length), entry.pkg]),
      ),
    ),
    commitsSince: () => [
      {hash: "abc1234", author: "Ada", message: "feat: add a capability"},
    ],
    tags: () => ["consumer-v2.0.0"],
  }),
  fs: memoryFileSystem(),
  output: memoryOutput(),
  applyChanges: (changes) => {
    applied.push([...changes]);
  },
  ...overrides,
});

const pathsOf = (changes: readonly FileChange[] | undefined): string[] =>
  (changes ?? []).map((change) => change.path).toSorted();

describe("runLockstep", () => {
  it("applies the manifests and changelogs of members and dependents at once", () => {
    const applied: FileChange[][] = [];

    expect(runLockstep(options(applied))).toBe(0);

    expect(applied).toHaveLength(1);
    expect(pathsOf(applied[0])).toEqual(
      [
        manifestPath("base"),
        "/repo/packages/base/CHANGELOG.md",
        manifestPath("prettier"),
        "/repo/packages/prettier/CHANGELOG.md",
        manifestPath("consumer"),
        "/repo/packages/consumer/CHANGELOG.md",
      ].toSorted(),
    );
    const consumer = applied[0]?.find(
      (change) => change.path === manifestPath("consumer"),
    );
    expect(consumer?.content).toContain('"@xonovex/base": "0.2.0"');
    expect(consumer?.content).toContain('"version": "2.0.1"');
    const changelog = applied[0]?.find(
      (change) => change.path === "/repo/packages/base/CHANGELOG.md",
    );
    expect(changelog?.content).toContain("## 0.2.0");
    expect(changelog?.content).toContain(
      "- Updated dependency `@xonovex/prettier` to `0.2.0`",
    );
  });

  it("moves the reference of a never-released dependent in the same write and keeps its version", () => {
    const applied: FileChange[][] = [];
    const output = memoryOutput();
    const released = options(applied);

    runLockstep(
      options(applied, {
        output,
        git: {...released.git, tags: () => []},
      }),
    );

    expect(applied).toHaveLength(1);
    expect(pathsOf(applied[0])).toEqual(
      [
        manifestPath("base"),
        "/repo/packages/base/CHANGELOG.md",
        manifestPath("prettier"),
        "/repo/packages/prettier/CHANGELOG.md",
        manifestPath("consumer"),
      ].toSorted(),
    );
    const consumer = applied[0]?.find(
      (change) => change.path === manifestPath("consumer"),
    );
    expect(consumer?.content).toContain('"@xonovex/base": "0.2.0"');
    expect(consumer?.content).toContain('"version": "2.0.0"');
    expect(output.stdoutLines().join("\n")).toContain(
      "@xonovex/consumer: 1 reference(s) updated (version unchanged)",
    );
  });

  it("writes only the manifests without changelogs or dependents", () => {
    const applied: FileChange[][] = [];

    runLockstep(options(applied, {noChangelog: true, noDependents: true}));

    expect(pathsOf(applied[0])).toEqual(
      [manifestPath("base"), manifestPath("prettier")].toSorted(),
    );
  });

  it("prints the changelog entries and applies nothing on a dry run", () => {
    const applied: FileChange[][] = [];
    const output = memoryOutput();

    runLockstep(options(applied, {dryRun: true, output}));

    expect(applied).toEqual([]);
    expect(output.stdoutLines().join("\n")).toContain("## 0.2.0");
  });

  it("writes no changelog for a member already at the target version", () => {
    const applied: FileChange[][] = [];

    runLockstep(
      options(applied, {
        exact: "0.1.0",
        noDependents: true,
      }),
    );

    expect(pathsOf(applied[0])).toEqual(
      [manifestPath("base"), manifestPath("prettier")].toSorted(),
    );
  });
});
