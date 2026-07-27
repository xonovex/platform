import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {dirname, resolve} from "node:path";
import {describe, expect, it} from "vitest";
import {validateRelease} from "./release-validation.js";

const VERSION = "1.2.3";
const PACKAGE_PATH = "packages/skill/skill-test";
const COMMAND_PATH = "packages/command/command-test";

const writeText = (root: string, path: string, content: string): void => {
  const target = resolve(root, path);
  mkdirSync(dirname(target), {recursive: true});
  writeFileSync(target, content);
};

const writeJson = (root: string, path: string, value: unknown): void => {
  writeText(root, path, `${JSON.stringify(value)}\n`);
};

const createFixture = (): string => {
  const root = mkdtempSync(resolve(".release-validation-"));
  const packageManifest = {
    name: "@xonovex/skill-test",
    version: VERSION,
    description: "Test skill",
  };
  const pluginManifest = {
    name: "xonovex-skill-test",
    version: VERSION,
    description: "Test skill",
  };
  const skillEntry = {
    name: "xonovex-skill-test",
    source: `./${PACKAGE_PATH}`,
    description: "Test skill",
  };
  const commandPackageManifest = {
    name: "@xonovex/command-test",
    version: VERSION,
    description: "Test command",
  };
  const commandPluginManifest = {
    name: "xonovex-test",
    version: VERSION,
    description: "Test command",
  };
  const commandEntry = {
    name: "xonovex-test",
    source: `./${COMMAND_PATH}`,
    description: "Test command",
  };
  const claudeMarketplace = {
    metadata: {version: VERSION},
    plugins: [skillEntry, commandEntry],
  };

  writeJson(root, ".claude-plugin/marketplace.json", claudeMarketplace);
  writeJson(root, ".agents/plugins/marketplace.json", {
    plugins: [skillEntry],
  });
  writeJson(root, `${PACKAGE_PATH}/package.json`, packageManifest);
  writeJson(root, `${PACKAGE_PATH}/.claude-plugin/plugin.json`, {
    ...pluginManifest,
    skills: ["./test-guide"],
  });
  writeJson(root, `${PACKAGE_PATH}/.codex-plugin/plugin.json`, {
    ...pluginManifest,
    skills: "./test-guide",
  });
  writeText(root, `${PACKAGE_PATH}/test-guide/SKILL.md`, "# Test skill\n");
  writeJson(root, `${COMMAND_PATH}/package.json`, commandPackageManifest);
  writeJson(
    root,
    `${COMMAND_PATH}/.claude-plugin/plugin.json`,
    commandPluginManifest,
  );
  writeJson(root, "package-lock.json", {
    packages: {
      [PACKAGE_PATH]: {version: VERSION},
      [COMMAND_PATH]: {version: VERSION},
    },
  });
  writeText(root, "README.md", "# Test repository\n");
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
      name: "xonovex-skill-test",
      source: `./${PACKAGE_PATH}`,
      description: "Test skill",
    };
    const command = {
      name: "xonovex-test",
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
        "Claude marketplace has duplicate plugin entries: xonovex-skill-test (2)",
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
          name: "xonovex-skill-test",
          source: {path: `./${PACKAGE_PATH}`},
          description: "Test skill",
        },
        {
          name: "xonovex-skill-retired",
          source: {path: "./packages/skill/skill-retired"},
        },
      ],
    });

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        "Codex marketplace has unexpected plugin entries: xonovex-skill-retired",
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });

  it("rejects a skill manifest path that does not resolve to its guide", () => {
    const root = createFixture();
    writeJson(root, `${PACKAGE_PATH}/.codex-plugin/plugin.json`, {
      name: "xonovex-skill-test",
      version: VERSION,
      description: "Test skill",
      skills: "./old-guide",
    });

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        `${PACKAGE_PATH} Codex manifest skill paths match ./test-guide`,
      );
      expect(result.failures).toContain(
        `${PACKAGE_PATH} manifest skill path resolves: ./old-guide`,
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });
});
