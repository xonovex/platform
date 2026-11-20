import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {describe, expect, it} from "vitest";
import {
  planDependentUpdates,
  type DependentUpdateOptions,
} from "../../../src/dependents-command.js";
import {emptyHistory, manifestsAtHead} from "../../util/version-git.js";

const manifestPath = (directory: string): string =>
  `/repo/packages/${directory}/package.json`;

const MANIFESTS = {
  core: {name: "@xonovex/core", version: "1.0.0"},
  consumer: {
    name: "@xonovex/consumer",
    version: "2.0.0",
    dependencies: {"@xonovex/core": "1.0.0"},
  },
  internal: {
    name: "@xonovex/internal",
    version: "0.0.0",
    private: true,
    devDependencies: {"@xonovex/core": "1.0.0"},
  },
  unrelated: {name: "@xonovex/unrelated", version: "1.0.0"},
} as const;

// HEAD holds every manifest unchanged, so a referencing public package is due
// a patch bump.
const options = (
  overrides: Partial<DependentUpdateOptions> = {},
): DependentUpdateOptions => ({
  rootDir: "/repo",
  packagePaths: Object.keys(MANIFESTS).map((name) => manifestPath(name)),
  packagePath: manifestPath("core"),
  packageName: "@xonovex/core",
  newVersion: "1.1.0",
  dryRun: false,
  noChangelog: false,
  changelogPath: undefined,
  gitBase: "base",
  includedTypes: undefined,
  git: emptyHistory({
    readFile: manifestsAtHead(
      Object.fromEntries(
        Object.entries(MANIFESTS).map(([name, manifest]) => [
          `packages/${name}/package.json`,
          manifest,
        ]),
      ),
    ),
  }),
  fs: memoryFileSystem({
    files: Object.fromEntries(
      Object.entries(MANIFESTS).map(([name, manifest]) => [
        manifestPath(name),
        JSON.stringify(manifest),
      ]),
    ),
  }),
  output: memoryOutput(),
  ...overrides,
});

describe("planDependentUpdates", () => {
  it("rewrites references, patch-bumps a public dependent and plans its changelog", () => {
    const result = planDependentUpdates(options());

    expect(result.updated).toBe(2);
    expect(result.changes.map((change) => change.path).toSorted()).toEqual(
      [
        manifestPath("consumer"),
        "/repo/packages/consumer/CHANGELOG.md",
        manifestPath("internal"),
      ].toSorted(),
    );
    const consumer = result.changes.find(
      (change) => change.path === manifestPath("consumer"),
    );
    expect(consumer?.content).toContain('"@xonovex/core": "1.1.0"');
    expect(consumer?.content).toContain('"version": "2.0.1"');
    const changelog = result.changes.find((change) =>
      change.path.endsWith("CHANGELOG.md"),
    );
    expect(changelog?.content).toContain(
      "- Updated dependency `@xonovex/core` to `1.1.0`",
    );
  });

  it("counts the dependents but plans no file on a dry run", () => {
    const result = planDependentUpdates(options({dryRun: true}));

    expect(result.updated).toBe(2);
    expect(result.changes).toEqual([]);
  });

  it("plans only the manifests when changelogs are off", () => {
    const result = planDependentUpdates(options({noChangelog: true}));

    expect(result.changes.map((change) => change.path).toSorted()).toEqual(
      [manifestPath("consumer"), manifestPath("internal")].toSorted(),
    );
  });

  it("names a dependent without a name by its path", () => {
    const path = manifestPath("anonymous");
    const result = planDependentUpdates(
      options({
        packagePaths: [path],
        fs: memoryFileSystem({
          files: {
            [path]: JSON.stringify({
              version: "1.0.0",
              dependencies: {"@xonovex/core": "1.0.0"},
            }),
          },
        }),
        git: emptyHistory({
          readFile: manifestsAtHead({
            "packages/anonymous/package.json": {version: "1.0.0"},
          }),
        }),
      }),
    );

    expect(
      result.changes.find((change) => change.path.endsWith("CHANGELOG.md"))
        ?.content,
    ).toContain(`# ${path}`);
  });
});
