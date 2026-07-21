import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {dirname, resolve} from "node:path";
import {describe, expect, it} from "vitest";
import {validateRelease} from "./release-validation.js";

const VERSION = "1.2.3";
const PACKAGE_PATH = "packages/skill/skill-test";

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
  const manifest = {
    name: "skill-test",
    version: VERSION,
    description: "Test skill",
  };
  const marketplace = {
    metadata: {version: VERSION},
    plugins: [
      {
        name: "skill-test",
        source: `./${PACKAGE_PATH}`,
        description: "Test skill",
      },
    ],
  };

  mkdirSync(resolve(root, "packages/command"), {recursive: true});
  writeJson(root, ".claude-plugin/marketplace.json", marketplace);
  writeJson(root, ".agents/plugins/marketplace.json", marketplace);
  writeJson(root, `${PACKAGE_PATH}/package.json`, manifest);
  writeJson(root, `${PACKAGE_PATH}/.claude-plugin/plugin.json`, manifest);
  writeJson(root, `${PACKAGE_PATH}/.codex-plugin/plugin.json`, manifest);
  writeJson(root, "package-lock.json", {
    packages: {[PACKAGE_PATH]: {version: VERSION}},
  });
  writeText(root, "README.md", "# Test repository\n");
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

      expect(result).toMatchObject({failures: [], pluginPackages: 1});
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
      name: "skill-test",
      source: `./${PACKAGE_PATH}`,
      description: "Test skill",
    };
    writeJson(root, ".claude-plugin/marketplace.json", {
      metadata: {version: VERSION},
      plugins: [duplicate, duplicate],
    });

    try {
      const result = validateRelease(root);

      expect(result.failures).toContain(
        "Claude marketplace contains every command and skill package exactly once",
      );
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });
});
