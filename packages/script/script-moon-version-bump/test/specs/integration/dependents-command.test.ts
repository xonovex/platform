import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {afterEach, describe, expect, it} from "vitest";
import {
  planDependentUpdates,
  type DependentUpdateOptions,
} from "../../../src/dependents-command.js";
import {runGit} from "../../../src/git.js";

const directories: string[] = [];
const gitExecutable = resolveExecutable("git");

// Git exports GIT_DIR, GIT_INDEX_FILE, and GIT_AUTHOR_* to the hooks it runs,
// and those take precedence over a repository's own config. Running these
// fixtures under a hook would otherwise commit with the caller's identity and
// resolve paths against the outer repository, so drop every GIT_* variable.
const gitEnvironment = Object.fromEntries(
  Object.entries(process.env).filter(([name]) => !name.startsWith("GIT_")),
);

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

const commit = (root: string, subject: string): void => {
  execFileSync(gitExecutable, ["add", "."], {cwd: root, env: gitEnvironment});
  execFileSync(gitExecutable, ["commit", "--quiet", "-m", subject], {
    cwd: root,
    env: gitEnvironment,
  });
};

const createRepository = (): {
  readonly root: string;
  readonly corePath: string;
  readonly consumerPath: string;
  readonly privatePath: string;
} => {
  const root = mkdtempSync(join(tmpdir(), "version-dependents-"));
  directories.push(root);
  mkdirSync(join(root, ".moon"));
  execFileSync(gitExecutable, ["init", "--quiet"], {
    cwd: root,
    env: gitEnvironment,
  });
  execFileSync(gitExecutable, ["config", "user.name", "Test Author"], {
    cwd: root,
    env: gitEnvironment,
  });
  execFileSync(gitExecutable, ["config", "user.email", "test@example.com"], {
    cwd: root,
    env: gitEnvironment,
  });
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
  commit(root, "feat: introduce the packages");
  writeFileSync(
    join(root, "packages", "consumer", "source.ts"),
    "export {};\n",
  );
  commit(root, "feat(consumer): add a capability");
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

afterEach(() => {
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
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
