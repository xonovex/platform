import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  checkDependencyUsage,
  type DependencyUsageInput,
} from "../../../src/dependency-usage.js";

const PLUGINS = "/repo/packages/runtime/plugin";
const PACKAGE_DIR = join(PLUGINS, "plugin-test");

// The fixture holds the plugin under test with one skill, and the core plugin
// with the skills `using-git` and `applying-tdd`.
const createPlugins = (skillBody: string, referenceBody?: string): FileSystem =>
  memoryFileSystem({
    directories: [PACKAGE_DIR],
    files: {
      [join(PACKAGE_DIR, "skills", "using-test", "SKILL.md")]: skillBody,
      ...(referenceBody === undefined
        ? {}
        : {
            [join(PACKAGE_DIR, "skills", "using-test", "references", "a.md")]:
              referenceBody,
          }),
      [join(PLUGINS, "plugin-core", "skills", "using-git", "SKILL.md")]:
        "# Git\n",
      [join(PLUGINS, "plugin-core", "skills", "applying-tdd", "SKILL.md")]:
        "# TDD\n",
    },
  });

const usageInput = (
  overrides: Partial<DependencyUsageInput> = {},
): DependencyUsageInput => ({
  packageDirectory: PACKAGE_DIR,
  repositoryRoot: "/repo",
  pluginDependencies: [],
  packageDependencies: [],
  ...overrides,
});

const codes = (
  fs: FileSystem,
  overrides: Partial<DependencyUsageInput>,
): readonly string[] =>
  checkDependencyUsage(usageInput(overrides), fs).map(({code}) => code);

describe("checkDependencyUsage", () => {
  it("accepts a dependency when a skill names one of its skills", () => {
    const fs = createPlugins("# Test\n\nCommits belong to **using-git**.\n");

    expect(
      codes(fs, {
        pluginDependencies: ["xonovex-core"],
        packageDependencies: ["@xonovex/plugin-core"],
      }),
    ).toEqual([]);
  });

  it("reports a dependency when no skill names one of its skills", () => {
    const fs = createPlugins("# Test\n\nRun the tests.\n");

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-core"],
          packageDependencies: ["@xonovex/plugin-core"],
        }),
        fs,
      ).map(({code, path}) => [code, path]),
    ).toEqual([
      [
        "dependency.plugin-unused",
        "packages/runtime/plugin/plugin-test/.claude-plugin/plugin.json",
      ],
      [
        "dependency.package-unused",
        "packages/runtime/plugin/plugin-test/package.json",
      ],
    ]);
  });

  it("accepts a dependency when a reference hands off to one of its skills", () => {
    const fs = createPlugins(
      "# Test\n",
      "# A\n\nTests follow **applying-tdd**.\n",
    );

    expect(codes(fs, {pluginDependencies: ["xonovex-core"]})).toEqual([]);
  });

  it("does not accept a plain skill name as a reason", () => {
    const fs = createPlugins(
      "# Test\n\nCommits belong to using-git.\n",
      "# A\n\nSee `applying-tdd`.\n",
    );

    expect(codes(fs, {pluginDependencies: ["xonovex-core"]})).toEqual([
      "dependency.plugin-unused",
    ]);
  });

  it("does not accept the name of the plugin as a reason", () => {
    const fs = createPlugins("# Test\n\nInstall xonovex-core first.\n");

    expect(codes(fs, {pluginDependencies: ["xonovex-core"]})).toEqual([
      "dependency.plugin-unused",
    ]);
  });

  it("does not accept a longer name that contains a skill name", () => {
    const fs = createPlugins("# Test\n\nUse **using-github** for this.\n");

    expect(codes(fs, {pluginDependencies: ["xonovex-core"]})).toEqual([
      "dependency.plugin-unused",
    ]);
  });

  it("reports a dependency on a plugin that has no skills", () => {
    const fs = createPlugins("# Test\n\nCommits belong to **using-git**.\n");

    expect(codes(fs, {pluginDependencies: ["xonovex-absent"]})).toEqual([
      "dependency.plugin-unused",
    ]);
  });

  it("ignores dependencies that are not skill plugins", () => {
    const fs = createPlugins("# Test\n");

    expect(
      codes(fs, {
        pluginDependencies: ["other-plugin"],
        packageDependencies: ["zod", "@xonovex/library"],
      }),
    ).toEqual([]);
  });
});
