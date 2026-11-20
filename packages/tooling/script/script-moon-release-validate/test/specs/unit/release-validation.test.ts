import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {GitRepository} from "@xonovex/script-moon-common/git";
import {describe, expect, it} from "vitest";
import {readReleaseSnapshot} from "../../../src/release-reader.js";
import {
  validateRelease,
  type ReleaseValidationResult,
} from "../../../src/release-validation.js";

const ROOT = "/repo";
const VERSION = "1.2.3";
const PACKAGE_PATH = "packages/runtime/plugin/plugin-test";
const EXTRA_PATH = "packages/runtime/plugin/plugin-extra";
const SCRIPT_PATH = "packages/tooling/script/script-test";

type Git = Pick<GitRepository, "trackedFiles" | "ignoredPaths">;

// An empty repository: nothing committed and nothing ignored.
const emptyRepository: Git = {trackedFiles: () => [], ignoredPaths: () => []};

const writeText = (fs: FileSystem, path: string, content: string): void => {
  fs.writeFile(join(ROOT, path), content);
};

const writeJson = (fs: FileSystem, path: string, value: unknown): void => {
  writeText(fs, path, `${JSON.stringify(value)}\n`);
};

const pluginPackage = (
  fs: FileSystem,
  path: string,
  name: string,
  description: string,
): void => {
  const manifest = {
    name: `xonovex-${name}`,
    version: VERSION,
    description,
    skills: "./skills/",
  };
  writeJson(fs, `${path}/package.json`, {
    name: `@xonovex/plugin-${name}`,
    version: VERSION,
    description,
  });
  writeJson(fs, `${path}/plugin.json`, manifest);
  writeJson(fs, `${path}/.claude-plugin/plugin.json`, manifest);
  writeJson(fs, `${path}/.codex-plugin/plugin.json`, manifest);
  writeText(fs, `${path}/skills/using-${name}/SKILL.md`, `# ${name}\n`);
};

const SKILL_ENTRY = {
  name: "xonovex-test",
  source: `./${PACKAGE_PATH}`,
  description: "Test skill",
};
const EXTRA_ENTRY = {
  name: "xonovex-extra",
  source: `./${EXTRA_PATH}`,
  description: "Test extra skill",
};

// A complete lockstep release with two plugin packages and one script project.
const createFixture = (): FileSystem => {
  const fs = memoryFileSystem({directories: [join(ROOT, ".moon")]});
  writeJson(fs, ".claude-plugin/marketplace.json", {
    metadata: {version: VERSION},
    plugins: [SKILL_ENTRY, EXTRA_ENTRY],
  });
  writeJson(fs, ".agents/plugins/marketplace.json", {
    plugins: [SKILL_ENTRY, EXTRA_ENTRY],
  });
  pluginPackage(fs, PACKAGE_PATH, "test", "Test skill");
  pluginPackage(fs, EXTRA_PATH, "extra", "Test extra skill");
  writeJson(fs, "package-lock.json", {
    packages: {
      [PACKAGE_PATH]: {version: VERSION},
      [EXTRA_PATH]: {version: VERSION},
    },
  });
  writeText(fs, "README.md", "# Test repository\n");
  for (const group of [
    "",
    "runtime",
    "runtime/plugin",
    "tooling",
    "tooling/script",
  ]) {
    writeText(fs, `packages/${group}/AGENTS.md`, `# ${group}\n`);
    writeText(
      fs,
      `packages/${group}/CLAUDE.md`,
      "See @AGENTS.md for complete documentation.\n",
    );
  }
  writeText(
    fs,
    `${SCRIPT_PATH}/moon.yml`,
    `language: typescript
tags: [typescript-script]
tasks:
  ts-coverage:
    env:
      TS_COVERAGE_MIN_LINES: "85"
      TS_COVERAGE_MIN_FUNCTIONS: "90"
      TS_COVERAGE_MIN_BRANCHES: "70"
      TS_COVERAGE_MIN_STATEMENTS: "85"
`,
  );
  writeText(
    fs,
    ".moon/tasks/tag-typescript.yml",
    `tasks:
  ci-check:
    deps: [build, lint, typecheck, test, format-check]
`,
  );
  writeText(
    fs,
    ".moon/tasks/tag-typescript-script.yml",
    `extends: ./tag-typescript.yml
tasks:
  ci-check:
    deps: [build, lint, typecheck, test, format-check, coverage]
`,
  );
  writeText(fs, ".moon/tasks/README.md", "not a task file\n");
  writeText(
    fs,
    ".github/workflows/release.yml",
    `on:
  workflow_dispatch:
  pull_request:
jobs:
  release:
    steps:
      - name: Publish
        if: github.event_name == 'pull_request'
        run: publish
      - name: Publish (dry run)
        if: github.event_name == 'workflow_dispatch'
        run: dry-run
      - uses: example/report@sha
`,
  );
  return fs;
};

const validate = (
  fs: FileSystem,
  git: Git = emptyRepository,
): ReleaseValidationResult =>
  validateRelease(readReleaseSnapshot(ROOT, fs, git));

describe("release input validation", () => {
  it("accepts a complete lockstep release fixture", () => {
    expect(validate(createFixture())).toMatchObject({
      failures: [],
      pluginPackages: 2,
    });
  });

  it("reports a typescript project that inherits the template coverage floor", () => {
    const fs = createFixture();
    writeText(
      fs,
      `${SCRIPT_PATH}/moon.yml`,
      `language: typescript
tags: [typescript-script]
tasks:
  ts-coverage:
    env:
      TS_COVERAGE_MIN_LINES: "85"
      TS_COVERAGE_MIN_STATEMENTS: "85"
`,
    );

    expect(validate(fs).failures).toEqual([
      expect.stringContaining(
        `${SCRIPT_PATH}/moon.yml task 'ts-coverage' sets no TS_COVERAGE_MIN_FUNCTIONS, TS_COVERAGE_MIN_BRANCHES`,
      ),
    ]);
  });

  it("reports malformed marketplace JSON and checks nothing else", () => {
    const fs = createFixture();
    writeText(fs, ".claude-plugin/marketplace.json", "{");

    expect(validate(fs)).toMatchObject({
      failures: [
        expect.stringContaining(
          ".claude-plugin/marketplace.json could not be read or parsed",
        ),
      ],
      pluginPackages: 0,
    });
  });

  it("reports an invalid package manifest at its field path", () => {
    const fs = createFixture();
    writeJson(fs, `${PACKAGE_PATH}/package.json`, {version: VERSION});

    expect(validate(fs).failures).toContain(
      `${PACKAGE_PATH}/package.json is invalid: name: Invalid input: expected string, received undefined`,
    );
  });

  it("reports an invalid lockfile package collection", () => {
    const fs = createFixture();
    writeJson(fs, "package-lock.json", {packages: []});

    expect(validate(fs).failures).toContain(
      "package-lock.json is invalid: packages: Invalid input: expected record, received array",
    );
  });

  it("reports a plugin package the lockfile records at another version", () => {
    const fs = createFixture();
    writeJson(fs, "package-lock.json", {
      packages: {[PACKAGE_PATH]: {version: VERSION}},
    });

    expect(validate(fs).failures).toEqual([
      `package-lock records ${EXTRA_PATH}@${VERSION}`,
    ]);
  });

  it("reports duplicate marketplace plugin names", () => {
    const fs = createFixture();
    writeJson(fs, ".claude-plugin/marketplace.json", {
      metadata: {version: VERSION},
      plugins: [SKILL_ENTRY, SKILL_ENTRY, EXTRA_ENTRY],
    });

    expect(validate(fs).failures).toContain(
      "Claude marketplace has duplicate plugin entries: xonovex-test (2)",
    );
  });

  it("names missing and unexpected Codex marketplace entries", () => {
    const fs = createFixture();
    writeJson(fs, ".agents/plugins/marketplace.json", {
      plugins: [
        {
          name: "xonovex-test",
          source: {path: `./${PACKAGE_PATH}`},
          description: "Test skill",
        },
        {
          name: "xonovex-retired",
          source: {path: "./packages/runtime/plugin/plugin-retired"},
        },
      ],
    });

    const {failures} = validate(fs);

    expect(failures).toContain(
      "Codex marketplace has unexpected plugin entries: xonovex-retired",
    );
    expect(failures).toContain(
      "Codex marketplace is missing plugin entries: xonovex-extra",
    );
  });

  it("rejects a skill manifest path that does not resolve to its skill", () => {
    const fs = createFixture();
    writeJson(fs, `${PACKAGE_PATH}/.codex-plugin/plugin.json`, {
      name: "xonovex-test",
      version: VERSION,
      description: "Test skill",
      skills: "./using-old",
    });

    const {failures} = validate(fs);

    expect(failures).toContain(
      `${PACKAGE_PATH} Codex manifest skill paths match ./skills/`,
    );
    expect(failures).toContain(
      `${PACKAGE_PATH} manifest skill path is direct: ./using-old`,
    );
  });

  it("reports a plugin package with no skills and a skill list manifest", () => {
    const fs = memoryFileSystem();
    const source = createFixture();
    for (const path of [
      ".claude-plugin/marketplace.json",
      ".agents/plugins/marketplace.json",
      "package-lock.json",
      "README.md",
      ".github/workflows/release.yml",
      ".moon/tasks/tag-typescript.yml",
      `${SCRIPT_PATH}/moon.yml`,
      `${PACKAGE_PATH}/package.json`,
      `${PACKAGE_PATH}/plugin.json`,
      `${PACKAGE_PATH}/.codex-plugin/plugin.json`,
      `${EXTRA_PATH}/package.json`,
      `${EXTRA_PATH}/plugin.json`,
      `${EXTRA_PATH}/.claude-plugin/plugin.json`,
      `${EXTRA_PATH}/.codex-plugin/plugin.json`,
      `${EXTRA_PATH}/skills/using-extra/SKILL.md`,
    ]) {
      writeText(fs, path, source.readText(join(ROOT, path)));
    }
    writeJson(fs, `${PACKAGE_PATH}/.claude-plugin/plugin.json`, {
      name: "xonovex-test",
      version: VERSION,
      description: "Test skill",
      skills: ["./skills/"],
    });

    const {failures} = validate(fs);

    expect(failures).toContain(`${PACKAGE_PATH} contains skills`);
    expect(failures).toContain(
      `${PACKAGE_PATH} Claude manifest exposes a skill directory`,
    );
    expect(failures).toContain(
      `${PACKAGE_PATH} manifest skill path resolves: ./skills/`,
    );
  });

  it("reports each missing plugin manifest", () => {
    const fs = createFixture();
    const other = memoryFileSystem();
    for (const path of [
      ".claude-plugin/marketplace.json",
      ".agents/plugins/marketplace.json",
    ]) {
      writeText(other, path, fs.readText(join(ROOT, path)));
    }
    other.makeDirectory(join(ROOT, PACKAGE_PATH));

    const {failures} = validate(other);

    expect(failures).toEqual(
      expect.arrayContaining([
        `${PACKAGE_PATH} has package.json`,
        `${PACKAGE_PATH} has a neutral plugin manifest`,
        `${PACKAGE_PATH} has a Claude plugin manifest`,
        `${PACKAGE_PATH} has a Codex plugin manifest`,
        expect.stringContaining("package-lock.json could not be read"),
        expect.stringContaining("README.md could not be read"),
        expect.stringContaining(
          ".github/workflows/release.yml could not be read",
        ),
        ".moon/tasks must define tag task files",
        "packages must define project task files",
      ]),
    );
  });

  it("reports a manifest path that exists but cannot be read", () => {
    const fs = createFixture();
    fs.makeDirectory(join(ROOT, EXTRA_PATH, "package.json.d"));
    const unreadable: FileSystem = {
      ...fs,
      readText: (path) => {
        if (path === join(ROOT, EXTRA_PATH, "plugin.json")) {
          throw new Error("permission denied");
        }
        return fs.readText(path);
      },
    };

    expect(validate(unreadable).failures).toContain(
      `${EXTRA_PATH}/plugin.json could not be read or parsed: permission denied`,
    );
  });

  it("reports drift between a plugin package and its manifests", () => {
    const fs = createFixture();
    writeJson(fs, `${PACKAGE_PATH}/package.json`, {
      name: "@xonovex/plugin-test",
      version: "1.0.0",
      description: "",
      dependencies: {"@xonovex/plugin-extra": "1.0.0"},
    });

    const {failures} = validate(fs);

    expect(failures).toEqual(
      expect.arrayContaining([
        "@xonovex/plugin-test package version matches 1.2.3",
        "@xonovex/plugin-test package declares a description",
        "@xonovex/plugin-test package and Claude manifest descriptions match",
        "@xonovex/plugin-test dependency @xonovex/plugin-extra matches 1.2.3",
      ]),
    );
  });

  it("reports a release version that is not semantic", () => {
    const fs = createFixture();
    writeJson(fs, ".claude-plugin/marketplace.json", {
      plugins: [SKILL_ENTRY, EXTRA_ENTRY],
    });

    expect(validate(fs).failures).toContain(
      "marketplace release must use a semantic version",
    );
  });

  it("reports a plugin directory that cannot be listed", () => {
    const fs = createFixture();
    const bare = memoryFileSystem();
    for (const path of [
      ".claude-plugin/marketplace.json",
      ".agents/plugins/marketplace.json",
    ]) {
      writeText(bare, path, fs.readText(join(ROOT, path)));
    }
    bare.makeDirectory(join(ROOT, "packages"));

    const result = validate(bare);

    expect(result.pluginPackages).toBe(0);
    expect(result.failures).toContainEqual(
      expect.stringMatching(
        /^packages\/runtime\/plugin could not be enumerated: /u,
      ),
    );
  });

  it("checks the README links and claims", () => {
    const fs = createFixture();
    writeText(
      fs,
      "README.md",
      "# Test\n\n[Agents](packages/AGENTS.md#top) [Gone](docs/gone.md) [Site](https://example.com)\n",
    );

    expect(validate(fs).failures).toEqual([
      "README.md link resolves: docs/gone.md",
    ]);
  });

  it("reports a committed-file listing git cannot give", () => {
    const git: Git = {
      trackedFiles: () => {
        throw new Error("not a git repository");
      },
      ignoredPaths: () => {
        throw new Error("not a git repository");
      },
    };

    expect(validate(createFixture(), git).failures).toEqual(
      expect.arrayContaining(["git could not list the committed files"]),
    );
  });

  it("checks the bin targets of the committed package manifests", () => {
    const fs = createFixture();
    writeJson(fs, `${SCRIPT_PATH}/package.json`, {
      name: "@xonovex/script-test",
      version: "0.0.0",
      private: true,
      bin: {"moon-test": "./bin.js"},
    });
    writeText(fs, `${SCRIPT_PATH}/tsconfig.json`, "{}\n");

    const failures = (paths: readonly string[]): readonly string[] =>
      validate(fs, {trackedFiles: () => paths, ignoredPaths: () => []})
        .failures;

    expect(failures([`${SCRIPT_PATH}/bin.js`])).toEqual([]);
    expect(failures([])).toEqual([
      expect.stringContaining("bin.js, which is not committed"),
    ]);
  });

  it("reports an AGENTS.md without its CLAUDE.md pointer", () => {
    const fs = createFixture();
    writeText(fs, "AGENTS.md", "# Root\n");

    expect(validate(fs).failures).toEqual([
      expect.stringContaining("CLAUDE.md"),
    ]);
  });
});
