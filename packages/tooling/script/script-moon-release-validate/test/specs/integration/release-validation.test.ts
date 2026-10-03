import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join, resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {describe, expect, it} from "vitest";
import {validateRelease} from "../../../src/release-validation.js";

const VERSION = "1.2.3";
const PACKAGE_PATH = "packages/runtime/plugin/plugin-test";
const COMMAND_PATH = "packages/runtime/plugin/plugin-commands";
const SCRIPT_PATH = "packages/tooling/script/script-test";

const writeText = (root: string, path: string, content: string): void => {
  const target = resolve(root, path);
  mkdirSync(dirname(target), {recursive: true});
  writeFileSync(target, content);
};

const writeJson = (root: string, path: string, value: unknown): void => {
  writeText(root, path, `${JSON.stringify(value)}\n`);
};

const createFixture = (): string => {
  const root = mkdtempSync(join(tmpdir(), "release-validation-"));
  // validateRelease asks git for the tracked and the ignored paths. An empty
  // repository answers both with an empty list, which is what these cases assert
  // against, and it makes the answer a property of the fixture rather than of the
  // checkout the test happens to run inside.
  execFileSync(resolveExecutable("git"), ["init", "--quiet"], {cwd: root});
  const packageManifest = {
    name: "@xonovex/plugin-test",
    version: VERSION,
    description: "Test skill",
  };
  const pluginManifest = {
    name: "xonovex-test",
    version: VERSION,
    description: "Test skill",
    skills: "./skills/",
  };
  const skillEntry = {
    name: "xonovex-test",
    source: `./${PACKAGE_PATH}`,
    description: "Test skill",
  };
  const commandPackageManifest = {
    name: "@xonovex/plugin-commands",
    version: VERSION,
    description: "Test command",
  };
  const commandPluginManifest = {
    name: "xonovex-commands",
    version: VERSION,
    description: "Test command",
    skills: "./skills/",
  };
  const commandEntry = {
    name: "xonovex-commands",
    source: `./${COMMAND_PATH}`,
    description: "Test command",
  };
  const claudeMarketplace = {
    metadata: {version: VERSION},
    plugins: [skillEntry, commandEntry],
  };

  writeJson(root, ".claude-plugin/marketplace.json", claudeMarketplace);
  writeJson(root, ".agents/plugins/marketplace.json", {
    plugins: [skillEntry, commandEntry],
  });
  writeJson(root, `${PACKAGE_PATH}/package.json`, packageManifest);
  writeJson(root, `${PACKAGE_PATH}/plugin.json`, pluginManifest);
  writeJson(root, `${PACKAGE_PATH}/.claude-plugin/plugin.json`, {
    ...pluginManifest,
    skills: "./skills/",
  });
  writeJson(root, `${PACKAGE_PATH}/.codex-plugin/plugin.json`, {
    ...pluginManifest,
    skills: "./skills/",
  });
  writeText(
    root,
    `${PACKAGE_PATH}/skills/test-guide/SKILL.md`,
    "# Test skill\n",
  );
  writeJson(root, `${COMMAND_PATH}/package.json`, commandPackageManifest);
  writeJson(root, `${COMMAND_PATH}/plugin.json`, commandPluginManifest);
  writeJson(root, `${COMMAND_PATH}/.claude-plugin/plugin.json`, {
    ...commandPluginManifest,
    skills: "./skills/",
  });
  writeJson(root, `${COMMAND_PATH}/.codex-plugin/plugin.json`, {
    ...commandPluginManifest,
    skills: "./skills/",
  });
  writeText(
    root,
    `${COMMAND_PATH}/skills/command-guide/SKILL.md`,
    "# Command skill\n",
  );
  writeJson(root, "package-lock.json", {
    packages: {
      [PACKAGE_PATH]: {version: VERSION},
      [COMMAND_PATH]: {version: VERSION},
    },
  });
  writeText(root, "README.md", "# Test repository\n");
  for (const group of [
    "",
    "runtime",
    "runtime/plugin",
    "tooling",
    "tooling/script",
  ]) {
    writeText(root, `packages/${group}/AGENTS.md`, `# ${group}\n`);
    writeText(
      root,
      `packages/${group}/CLAUDE.md`,
      "See @AGENTS.md for complete documentation.\n",
    );
  }
  writeText(
    root,
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
    root,
    ".moon/tasks/tag-typescript.yml",
    `tasks:
  ci-check:
    deps: [build, lint, typecheck, test, format-check]
`,
  );
  writeText(
    root,
    ".moon/tasks/tag-typescript-script.yml",
    `extends: ./tag-typescript.yml
tasks:
  ci-check:
    deps: [build, lint, typecheck, test, format-check, coverage]
`,
  );
  writeText(
    root,
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
  return root;
};

describe("release input validation", () => {
  it("accepts a complete lockstep release fixture", () => {
    const root = createFixture();

    try {
      const result = validateRelease(root);

      expect(result).toMatchObject({failures: [], pluginPackages: 2});
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("reports a typescript project that inherits the template coverage floor", () => {
    const root = createFixture();
    writeText(
      root,
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

    try {
      const result = validateRelease(root);

      expect(result.failures).toEqual([
        expect.stringContaining(
          `${SCRIPT_PATH}/moon.yml task 'ts-coverage' sets no TS_COVERAGE_MIN_FUNCTIONS, TS_COVERAGE_MIN_BRANCHES`,
        ),
      ]);
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("reports malformed marketplace JSON without throwing", () => {
    const root = createFixture();
    writeText(root, ".claude-plugin/marketplace.json", "{");

    try {
      const result = validateRelease(root);

      expect(result.failures).toEqual([
        expect.stringContaining(
          ".claude-plugin/marketplace.json could not be read or parsed",
        ),
      ]);
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("reports an invalid package manifest at its field path", () => {
    const root = createFixture();
    writeJson(root, `${PACKAGE_PATH}/package.json`, {version: VERSION});

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        `${PACKAGE_PATH}/package.json is invalid: name: Invalid input: expected string, received undefined`,
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("reports an invalid lockfile package collection", () => {
    const root = createFixture();
    writeJson(root, "package-lock.json", {packages: []});

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        "package-lock.json is invalid: packages: Invalid input: expected record, received array",
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("reports duplicate marketplace plugin names", () => {
    const root = createFixture();
    const duplicate = {
      name: "xonovex-test",
      source: `./${PACKAGE_PATH}`,
      description: "Test skill",
    };
    const command = {
      name: "xonovex-commands",
      source: `./${COMMAND_PATH}`,
      description: "Test command",
    };
    writeJson(root, ".claude-plugin/marketplace.json", {
      metadata: {version: VERSION},
      plugins: [duplicate, duplicate, command],
    });

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        "Claude marketplace has duplicate plugin entries: xonovex-test (2)",
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("names unexpected Codex marketplace entries", () => {
    const root = createFixture();
    writeJson(root, ".agents/plugins/marketplace.json", {
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

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        "Codex marketplace has unexpected plugin entries: xonovex-retired",
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("rejects a skill manifest path that does not resolve to its guide", () => {
    const root = createFixture();
    writeJson(root, `${PACKAGE_PATH}/.codex-plugin/plugin.json`, {
      name: "xonovex-test",
      version: VERSION,
      description: "Test skill",
      skills: "./old-guide",
    });

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        `${PACKAGE_PATH} Codex manifest skill paths match ./skills/`,
      );
      expect(result.failures).toContain(
        `${PACKAGE_PATH} manifest skill path is direct: ./old-guide`,
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });
});
