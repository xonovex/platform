import {execFileSync} from "node:child_process";
import {
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {afterEach, describe, expect, it} from "vitest";
import {getGitVersion} from "./git.js";
import {runLockstep, type LockstepOptions} from "./lockstep-command.js";
import type {WorkspacePackage} from "./lockstep.js";

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
type Manifests = Readonly<Record<string, Manifest>>;

const configLine = (version: string): Manifests => ({
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

const commit = (root: string, subject: string): void => {
  execFileSync(gitExecutable, ["add", "."], {cwd: root, env: gitEnvironment});
  execFileSync(gitExecutable, ["commit", "--quiet", "-m", subject], {
    cwd: root,
    env: gitEnvironment,
  });
};

const createRepository = (line: (version: string) => Manifests): string => {
  const root = mkdtempSync(join(tmpdir(), "version-lockstep-"));
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
  writeManifests(root, line("0.1.21"));
  commit(root, "feat: introduce the line");
  writeManifests(root, line("0.1.22"));
  commit(root, "chore: release 0.1.22");
  for (const directory of readdirSync(join(root, "packages"))) {
    writeFileSync(
      join(root, "packages", directory, "source.ts"),
      "export {};\n",
    );
  }
  commit(root, "feat(line): add a shared capability");
  return root;
};

const workspacePackages = (root: string): readonly WorkspacePackage[] =>
  readdirSync(join(root, "packages")).map((directory) => {
    const path = join(root, "packages", directory, "package.json");
    return {path, pkg: readPkg(path), headVersion: getGitVersion(root, path)};
  });

const lockstepOptions = (
  root: string,
  names: readonly string[],
  overrides: Partial<LockstepOptions> = {},
): LockstepOptions => ({
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
  readPkg(join(root, "packages", directory, "package.json"));

const changelogAt = (root: string, directory: string): string =>
  readFileSync(join(root, "packages", directory, "CHANGELOG.md"), "utf8");

afterEach(() => {
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
});

describe("runLockstep", () => {
  it("moves a dependency cycle to one version in a single write", () => {
    const root = createRepository(configLine);

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
    const root = createRepository(configLine);

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
    const root = createRepository(configLine);
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
    const root = createRepository(configLine);

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
    const root = createRepository(configLine);

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
    const root = createRepository(configLine);

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
    const root = createRepository(configLine);
    const names = ["eslint-config-base", "prettier-config"];

    runLockstep(lockstepOptions(root, names));
    runLockstep(lockstepOptions(root, names));

    expect(manifestAt(root, "eslint-config-base").version).toBe("0.2.0");
    const changelog = changelogAt(root, "eslint-config-base");
    expect(changelog.match(/## 0\.2\.0/g)).toHaveLength(1);
  });

  it("writes nothing during a dry run", () => {
    const root = createRepository(configLine);

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
    const root = createRepository(configLine);
    const base = execFileSync(gitExecutable, ["rev-parse", "HEAD~1"], {
      cwd: root,
      encoding: "utf8",
      env: gitEnvironment,
    }).trim();

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
    const root = createRepository(configLine);

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
});
