import {mkdirSync, readdirSync, readFileSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {describe, expect, it} from "vitest";
import {applyFileChangesOnDisk} from "../../../src/adapters/file-transaction-node.js";
import {getGitVersion} from "../../../src/git.js";
import {
  runLockstep,
  type LockstepOptions,
} from "../../../src/lockstep-command.js";
import type {WorkspacePackage} from "../../../src/lockstep.js";
import {
  commitAll,
  git,
  gitRepositories,
  openGitRepository,
} from "../../util/git-repository.js";

const repository = gitRepositories();

type Manifest = Readonly<Record<string, unknown>>;
type Manifests = Readonly<Record<string, Manifest>>;

const configMembers = (version: string): Manifests => ({
  "eslint-config-base": {
    name: "@xonovex/eslint-config-base",
    version,
    devDependencies: {"@xonovex/prettier-config": version},
  },
  "prettier-config": {
    name: "@xonovex/prettier-config",
    version,
    dependencies: {"@xonovex/eslint-config-base": version},
  },
});

const configLine = (version: string): Manifests => ({
  ...configMembers(version),
  consumer: {
    name: "@xonovex/consumer",
    version: "1.4.0",
    devDependencies: {
      "@xonovex/eslint-config-base": version,
      "@xonovex/prettier-config": version,
    },
  },
  // Holds an exact reference to the consumer rather than to a member, so its
  // reference only moves once the consumer itself is patch-bumped.
  "downstream-consumer": {
    name: "@xonovex/downstream-consumer",
    version: "3.0.0",
    dependencies: {"@xonovex/consumer": "1.4.0"},
  },
});

// The consumers above keep one version, so only their release tags record that
// they were released.
const CONFIG_RELEASES = ["consumer-v1.4.0", "downstream-consumer-v3.0.0"];

// A launcher that was never released holds an exact reference to a member next
// to its own platform package. Of the two released dependents, one records its
// release as a version that moved between the commits, the other as a tag. The
// cli line is a launcher with two platform packages and a private release
// package, released when a case tags it.
const releaseLine = (version: string): Manifests => ({
  ...configMembers(version),
  launcher: {
    name: "@xonovex/launcher",
    version: "0.1.0",
    devDependencies: {"@xonovex/eslint-config-base": version},
    optionalDependencies: {"@xonovex/launcher-linux-x64": "0.1.0"},
  },
  "launcher-linux-x64": {name: "@xonovex/launcher-linux-x64", version: "0.1.0"},
  "versioned-consumer": {
    name: "@xonovex/versioned-consumer",
    version: version === "0.1.21" ? "1.0.0" : "1.1.0",
    devDependencies: {"@xonovex/prettier-config": version},
  },
  "tagged-consumer": {
    name: "@xonovex/tagged-consumer",
    version: "2.0.0",
    devDependencies: {"@xonovex/eslint-config-base": version},
  },
  cli: {
    name: "@xonovex/cli",
    version: "0.3.0",
    devDependencies: {"@xonovex/eslint-config-base": version},
    optionalDependencies: {
      "@xonovex/cli-linux-x64": "0.3.0",
      "@xonovex/cli-win32-x64": "0.3.0",
    },
  },
  "cli-linux-x64": {name: "@xonovex/cli-linux-x64", version: "0.3.0"},
  "cli-win32-x64": {name: "@xonovex/cli-win32-x64", version: "0.3.0"},
  "cli-github": {
    name: "@xonovex/cli-github",
    version: "0.0.0",
    private: true,
    optionalDependencies: {
      "@xonovex/cli-linux-x64": "0.3.0",
      "@xonovex/cli-win32-x64": "0.3.0",
    },
  },
});

const agentLine = (version: string): Manifests => ({
  "agent-cli-go": {
    name: "@xonovex/agent-cli-go",
    version,
    optionalDependencies: {"@xonovex/agent-cli-go-linux-x64": version},
  },
  "agent-cli-go-linux-x64": {
    name: "@xonovex/agent-cli-go-linux-x64",
    version,
  },
  "agent-cli-go-github": {
    name: "@xonovex/agent-cli-go-github",
    version: "0.0.0",
    private: true,
    optionalDependencies: {"@xonovex/agent-cli-go-linux-x64": version},
  },
});

const writeManifests = (root: string, manifests: Manifests): void => {
  for (const [directory, manifest] of Object.entries(manifests)) {
    const packageDirectory = join(root, "packages", directory);
    mkdirSync(packageDirectory, {recursive: true});
    writeFileSync(
      join(packageDirectory, "package.json"),
      `${JSON.stringify(manifest, null, 2)}\n`,
    );
  }
};

const createRepository = (
  line: (version: string) => Manifests,
  releaseTags: readonly string[] = [],
): string => {
  const root = repository("version-lockstep-");
  mkdirSync(join(root, ".moon"));
  writeManifests(root, line("0.1.21"));
  commitAll(root, "feat: introduce the line");
  writeManifests(root, line("0.1.22"));
  commitAll(root, "chore: release 0.1.22");
  for (const directory of readdirSync(join(root, "packages"))) {
    writeFileSync(
      join(root, "packages", directory, "source.ts"),
      "export {};\n",
    );
  }
  commitAll(root, "feat(line): add a shared capability");
  for (const tag of releaseTags) git(root, ["tag", tag]);
  return root;
};

const workspacePackages = (root: string): readonly WorkspacePackage[] =>
  readdirSync(join(root, "packages")).map((directory) => {
    const path = join(root, "packages", directory, "package.json");
    return {
      path,
      pkg: readPkg(path, nodeFileSystem),
      headVersion: getGitVersion(root, path, openGitRepository(root)),
    };
  });

const lockstepOptions = (
  root: string,
  names: readonly string[],
  overrides: Partial<LockstepOptions> = {},
): LockstepOptions => ({
  git: openGitRepository(root),
  fs: nodeFileSystem,
  output: memoryOutput(),
  applyChanges: applyFileChangesOnDisk,
  rootDir: root,
  packages: workspacePackages(root),
  names,
  bumpType: "minor",
  preid: undefined,
  exact: undefined,
  dryRun: false,
  noChangelog: false,
  noDependents: false,
  changelogPath: undefined,
  gitBase: undefined,
  includedTypes: undefined,
  ...overrides,
});

const manifestAt = (root: string, directory: string): Manifest =>
  readPkg(join(root, "packages", directory, "package.json"), nodeFileSystem);

const changelogAt = (root: string, directory: string): string =>
  readFileSync(join(root, "packages", directory, "CHANGELOG.md"), "utf8");

describe("runLockstep", () => {
  it("moves a dependency cycle to one version in a single write", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    const exitCode = runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"]),
    );

    expect(exitCode).toBe(0);
    expect(manifestAt(root, "eslint-config-base")).toMatchObject({
      version: "0.2.0",
      devDependencies: {"@xonovex/prettier-config": "0.2.0"},
    });
    expect(manifestAt(root, "prettier-config")).toMatchObject({
      version: "0.2.0",
      dependencies: {"@xonovex/eslint-config-base": "0.2.0"},
    });
  });

  it("records the shared version and its dependency updates in each changelog", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"]),
    );

    const changelog = changelogAt(root, "eslint-config-base");
    expect(changelog).toContain("# @xonovex/eslint-config-base");
    expect(changelog).toContain("## 0.2.0");
    expect(changelog).toContain("### Minor Changes");
    expect(changelog).toContain("add a shared capability");
    expect(changelog).toContain(
      "- Updated dependency `@xonovex/prettier-config` to `0.2.0`",
    );
    expect(changelogAt(root, "prettier-config")).toContain(
      "- Updated dependency `@xonovex/eslint-config-base` to `0.2.0`",
    );
  });

  it("lifts a member a peer already patch-bumped to the shared version", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);
    writeManifests(root, {
      "eslint-config-base": {
        name: "@xonovex/eslint-config-base",
        version: "0.1.23",
        devDependencies: {"@xonovex/prettier-config": "0.1.22"},
      },
    });

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"]),
    );

    expect(manifestAt(root, "eslint-config-base").version).toBe("0.2.0");
    expect(manifestAt(root, "prettier-config").version).toBe("0.2.0");
    expect(changelogAt(root, "eslint-config-base")).toContain("## 0.2.0");
  });

  it("patch-bumps an out-of-set dependent once and moves both references", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"]),
    );

    expect(manifestAt(root, "consumer")).toMatchObject({
      version: "1.4.1",
      devDependencies: {
        "@xonovex/eslint-config-base": "0.2.0",
        "@xonovex/prettier-config": "0.2.0",
      },
    });
    expect(changelogAt(root, "consumer")).toContain("## 1.4.1");
  });

  it("propagates a patch-bumped dependent to its own dependents", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"]),
    );

    expect(manifestAt(root, "downstream-consumer")).toMatchObject({
      version: "3.0.1",
      dependencies: {"@xonovex/consumer": "1.4.1"},
    });
    expect(changelogAt(root, "downstream-consumer")).toContain(
      "- Updated dependency `@xonovex/consumer` to `1.4.1`",
    );
  });

  it("keeps out-of-set references untouched with --no-dependents", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        noDependents: true,
      }),
    );

    expect(manifestAt(root, "consumer")).toMatchObject({
      version: "1.4.0",
      devDependencies: {"@xonovex/eslint-config-base": "0.1.22"},
    });
  });

  it("moves optionalDependencies of members and of a private consumer", () => {
    const root = createRepository(agentLine);

    runLockstep(
      lockstepOptions(root, ["agent-cli-go", "agent-cli-go-linux-x64"]),
    );

    expect(manifestAt(root, "agent-cli-go")).toMatchObject({
      version: "0.2.0",
      optionalDependencies: {"@xonovex/agent-cli-go-linux-x64": "0.2.0"},
    });
    expect(manifestAt(root, "agent-cli-go-linux-x64").version).toBe("0.2.0");
    expect(manifestAt(root, "agent-cli-go-github")).toMatchObject({
      version: "0.0.0",
      optionalDependencies: {"@xonovex/agent-cli-go-linux-x64": "0.2.0"},
    });
  });

  it("repeats without stacking a second changelog entry", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);
    const names = ["eslint-config-base", "prettier-config"];

    runLockstep(lockstepOptions(root, names));
    runLockstep(lockstepOptions(root, names));

    expect(manifestAt(root, "eslint-config-base").version).toBe("0.2.0");
    const changelog = changelogAt(root, "eslint-config-base");
    expect(changelog.match(/## 0\.2\.0/g)).toHaveLength(1);
  });

  it("writes nothing during a dry run", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        dryRun: true,
      }),
    );

    expect(manifestAt(root, "eslint-config-base")).toMatchObject({
      version: "0.1.22",
      devDependencies: {"@xonovex/prettier-config": "0.1.22"},
    });
    expect(manifestAt(root, "consumer").version).toBe("1.4.0");
    expect(
      readdirSync(join(root, "packages", "eslint-config-base")),
    ).not.toContain("CHANGELOG.md");
  });

  it("honours an explicit changelog base ref and filename", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);
    const base = git(root, ["rev-parse", "HEAD~1"]);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        gitBase: base,
        changelogPath: "CHANGES.md",
      }),
    );

    const changelog = readFileSync(
      join(root, "packages", "eslint-config-base", "CHANGES.md"),
      "utf8",
    );
    expect(changelog).toContain("## 0.2.0");
    expect(changelog).toContain("add a shared capability");
  });

  it("skips changelog generation on request", () => {
    const root = createRepository(configLine, CONFIG_RELEASES);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        noChangelog: true,
      }),
    );

    expect(manifestAt(root, "eslint-config-base").version).toBe("0.2.0");
    expect(
      readdirSync(join(root, "packages", "eslint-config-base")),
    ).not.toContain("CHANGELOG.md");
  });

  it("keeps a never-released dependent at its version and moves its exact reference", () => {
    const root = createRepository(releaseLine, ["tagged-consumer-v2.0.0"]);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        bumpType: "patch",
      }),
    );

    expect(manifestAt(root, "launcher")).toEqual({
      name: "@xonovex/launcher",
      version: "0.1.0",
      devDependencies: {"@xonovex/eslint-config-base": "0.1.23"},
      optionalDependencies: {"@xonovex/launcher-linux-x64": "0.1.0"},
    });
    expect(readdirSync(join(root, "packages", "launcher"))).not.toContain(
      "CHANGELOG.md",
    );
    expect(manifestAt(root, "launcher-linux-x64").version).toBe("0.1.0");
  });

  it("still patch-bumps a dependent whose version moved or that carries a release tag", () => {
    const root = createRepository(releaseLine, ["tagged-consumer-v2.0.0"]);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        bumpType: "patch",
      }),
    );

    expect(manifestAt(root, "versioned-consumer")).toMatchObject({
      version: "1.1.1",
      devDependencies: {"@xonovex/prettier-config": "0.1.23"},
    });
    expect(changelogAt(root, "versioned-consumer")).toContain("## 1.1.1");
    expect(manifestAt(root, "tagged-consumer")).toMatchObject({
      version: "2.0.1",
      devDependencies: {"@xonovex/eslint-config-base": "0.1.23"},
    });
    expect(changelogAt(root, "tagged-consumer")).toContain("## 2.0.1");
  });

  it("moves members that reference each other with the line, released or not", () => {
    const root = createRepository(releaseLine);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        bumpType: "patch",
      }),
    );

    expect(manifestAt(root, "eslint-config-base")).toMatchObject({
      version: "0.1.23",
      devDependencies: {"@xonovex/prettier-config": "0.1.23"},
    });
    expect(manifestAt(root, "prettier-config")).toMatchObject({
      version: "0.1.23",
      dependencies: {"@xonovex/eslint-config-base": "0.1.23"},
    });
    expect(manifestAt(root, "tagged-consumer").version).toBe("2.0.0");
  });

  it("names the never-released dependent and writes nothing on a dry run", () => {
    const root = createRepository(releaseLine, ["tagged-consumer-v2.0.0"]);
    const output = memoryOutput();

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        bumpType: "patch",
        dryRun: true,
        output,
      }),
    );

    const printed = output.stdoutLines().join("\n");
    expect(printed).toContain(
      "@xonovex/launcher: 1 reference(s) updated (version unchanged)",
    );
    expect(printed).toContain(
      "@xonovex/tagged-consumer: 1 reference(s) updated (2.0.0 -> 2.0.1)",
    );
    expect(manifestAt(root, "launcher")).toMatchObject({
      devDependencies: {"@xonovex/eslint-config-base": "0.1.22"},
    });
    expect(git(root, ["status", "--porcelain"])).toBe("");
  });

  it("moves a released launcher with its platform packages in one write", () => {
    const root = createRepository(releaseLine, ["cli-v0.3.0"]);

    runLockstep(
      lockstepOptions(root, ["eslint-config-base", "prettier-config"], {
        bumpType: "patch",
      }),
    );

    expect(manifestAt(root, "cli")).toEqual({
      name: "@xonovex/cli",
      version: "0.3.1",
      devDependencies: {"@xonovex/eslint-config-base": "0.1.23"},
      optionalDependencies: {
        "@xonovex/cli-linux-x64": "0.3.1",
        "@xonovex/cli-win32-x64": "0.3.1",
      },
    });
    for (const directory of ["cli", "cli-linux-x64", "cli-win32-x64"]) {
      expect(manifestAt(root, directory).version).toBe("0.3.1");
      expect(changelogAt(root, directory)).toContain("## 0.3.1");
    }
    expect(manifestAt(root, "cli-github")).toMatchObject({
      version: "0.0.0",
      optionalDependencies: {
        "@xonovex/cli-linux-x64": "0.3.1",
        "@xonovex/cli-win32-x64": "0.3.1",
      },
    });
    expect(manifestAt(root, "launcher").version).toBe("0.1.0");
    expect(manifestAt(root, "launcher-linux-x64").version).toBe("0.1.0");
  });

  it("refuses a lockstep set that splits a release line and writes nothing", () => {
    const root = createRepository(releaseLine, ["cli-v0.3.0"]);

    expect(() =>
      runLockstep(lockstepOptions(root, ["cli", "cli-linux-x64"])),
    ).toThrow("splits the release line of @xonovex/cli");
    expect(git(root, ["status", "--porcelain"])).toBe("");
  });
});
