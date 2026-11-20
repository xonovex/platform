import {mkdirSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {describe, expect, it} from "vitest";
import {
  planDependentUpdates,
  type DependentUpdateOptions,
} from "../../../src/dependents-command.js";
import {commitAll, git, gitRepositories} from "../../util/git-repository.js";

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
  git(root, ["tag", "consumer-v2.0.0"]);
  return {root, corePath, consumerPath, privatePath};
};

const dependentOptions = (
  root: string,
  packagePaths: readonly string[],
  overrides: Partial<DependentUpdateOptions> = {},
): DependentUpdateOptions => ({
  rootDir: root,
  packagePaths,
  git: nodeGitRepository(root),
  fs: nodeFileSystem,
  output: memoryOutput(),
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
      dependentOptions(root, [corePath, consumerPath, privatePath]),
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

  it("rewrites the reference of a never-released dependent and keeps its version", () => {
    const {root, corePath, consumerPath} = createRepository();
    git(root, ["tag", "--delete", "consumer-v2.0.0"]);

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, consumerPath]),
    );

    expect(result.updated).toBe(1);
    expect(result.changes.map((change) => change.path)).toEqual([consumerPath]);
    expect(result.changes[0]?.content).toContain('"@xonovex/core": "1.1.0"');
    expect(result.changes[0]?.content).toContain('"version": "2.0.0"');
  });

  it("plans no file changes for a dry run", () => {
    const {root, corePath, consumerPath} = createRepository();

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, consumerPath], {
        dryRun: true,
      }),
    );

    expect(result.updated).toBe(1);
    expect(result.changes).toEqual([]);
  });

  it("skips the changelog on request", () => {
    const {root, corePath, consumerPath} = createRepository();

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, consumerPath], {
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
      dependentOptions(root, [corePath, unrelatedPath]),
    );

    expect(result).toEqual({updated: 0, changes: []});
  });

  it("moves a released launcher dependent with its platform packages", () => {
    const {root, corePath} = createRepository();
    const launcherPath = writeManifest(root, "cli", {
      name: "@xonovex/cli",
      version: "0.3.0",
      devDependencies: {"@xonovex/core": "1.0.0"},
      optionalDependencies: {"@xonovex/cli-linux-x64": "0.3.0"},
    });
    const platformPath = writeManifest(root, "cli-linux-x64", {
      name: "@xonovex/cli-linux-x64",
      version: "0.3.0",
    });
    commitAll(root, "feat(cli): introduce the cli line");
    git(root, ["tag", "cli-v0.3.0"]);

    const result = planDependentUpdates(
      dependentOptions(root, [corePath, launcherPath, platformPath], {
        noChangelog: true,
      }),
    );

    expect(result.changes.map((change) => change.path)).toEqual([
      launcherPath,
      platformPath,
    ]);
    expect(result.changes[0]?.content).toContain('"version": "0.3.1"');
    expect(result.changes[0]?.content).toContain(
      '"@xonovex/cli-linux-x64": "0.3.1"',
    );
    expect(result.changes[1]?.content).toContain('"version": "0.3.1"');
  });
});
