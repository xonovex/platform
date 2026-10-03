import {mkdirSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  planDependentUpdates,
  type DependentUpdateOptions,
} from "../../../src/dependents-command.js";
import {runGit} from "../../../src/git.js";
import {commitAll, gitRepositories} from "../../util/git-repository.js";

const repository = gitRepositories();

type Manifest = Readonly<Record<string, unknown>>;

const writeManifest = (
  root: string,
  directory: string,
  manifest: Manifest,
): string => {
  const packageDirectory = join(root, "packages", directory);
  mkdirSync(packageDirectory, {recursive: true});
  const path = join(packageDirectory, "package.json");
  writeFileSync(path, `${JSON.stringify(manifest, null, 2)}\n`);
  return path;
};

const createRepository = (): {
  readonly root: string;
  readonly corePath: string;
  readonly consumerPath: string;
  readonly privatePath: string;
} => {
  const root = repository("version-dependents-");
  mkdirSync(join(root, ".moon"));
  const corePath = writeManifest(root, "core", {
    name: "@xonovex/core",
    version: "1.0.0",
  });
  const consumerPath = writeManifest(root, "consumer", {
    name: "@xonovex/consumer",
    version: "2.0.0",
    dependencies: {"@xonovex/core": "1.0.0"},
  });
  const privatePath = writeManifest(root, "internal", {
    name: "@xonovex/internal",
    version: "0.0.0",
    private: true,
    devDependencies: {"@xonovex/core": "1.0.0"},
  });
  commitAll(root, "feat: introduce the packages");
  writeFileSync(
    join(root, "packages", "consumer", "source.ts"),
    "export {};\n",
  );
  commitAll(root, "feat(consumer): add a capability");
  return {root, corePath, consumerPath, privatePath};
};

const dependentOptions = (
  root: string,
  packagePaths: readonly string[],
  packagePath: string,
  overrides: Partial<DependentUpdateOptions> = {},
): DependentUpdateOptions => ({
  rootDir: root,
  packagePaths,
  packagePath,
  git: runGit,
  packageName: "@xonovex/core",
  newVersion: "1.1.0",
  dryRun: false,
  noChangelog: false,
  changelogPath: undefined,
  gitBase: undefined,
  includedTypes: undefined,
  ...overrides,
});

describe("planDependentUpdates", () => {
  it("rewrites references, patch-bumps a public dependent and writes its changelog", () => {
    const {root, corePath, consumerPath, privatePath} = createRepository();

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, consumerPath, privatePath], corePath),
    );

    expect(result.updated).toBe(2);
    const consumerChange = result.changes.find(
      (change) => change.path === consumerPath,
    );
    expect(consumerChange?.content).toContain('"@xonovex/core": "1.1.0"');
    expect(consumerChange?.content).toContain('"version": "2.0.1"');
    const privateChange = result.changes.find(
      (change) => change.path === privatePath,
    );
    expect(privateChange?.content).toContain('"version": "0.0.0"');
    const changelog = result.changes.find((change) =>
      change.path.endsWith("CHANGELOG.md"),
    );
    expect(changelog?.content).toContain("## 2.0.1");
    expect(changelog?.content).toContain(
      "- Updated dependency `@xonovex/core` to `1.1.0`",
    );
  });

  it("plans no file changes for a dry run", () => {
    const {root, corePath, consumerPath} = createRepository();

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, consumerPath], corePath, {
        dryRun: true,
      }),
    );

    expect(result.updated).toBe(1);
    expect(result.changes).toEqual([]);
  });

  it("skips the changelog on request", () => {
    const {root, corePath, consumerPath} = createRepository();

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, consumerPath], corePath, {
        noChangelog: true,
      }),
    );

    expect(result.changes.map((change) => change.path)).toEqual([consumerPath]);
  });

  it("ignores packages that do not reference the bumped package", () => {
    const {root, corePath} = createRepository();
    const unrelatedPath = writeManifest(root, "unrelated", {
      name: "@xonovex/unrelated",
      version: "1.0.0",
    });

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, unrelatedPath], corePath),
    );

    expect(result).toEqual({updated: 0, changes: []});
  });
});
