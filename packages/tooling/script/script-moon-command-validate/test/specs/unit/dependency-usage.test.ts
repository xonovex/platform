import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  checkDependencyUsage,
  type DependencyUsageInput,
} from "../../../src/dependency-usage.js";

const PACKAGE_DIR = "/repo/packages/runtime/plugin/pkg";

const createPackage = (
  readme: string,
  commands: Readonly<Record<string, string>> = {},
): FileSystem =>
  memoryFileSystem({
    directories: [join(PACKAGE_DIR, "commands")],
    files: {
      [join(PACKAGE_DIR, "README.md")]: readme,
      ...Object.fromEntries(
        Object.entries(commands).map(([name, body]) => [
          join(PACKAGE_DIR, "commands", `${name}.md`),
          body,
        ]),
      ),
    },
  });

const usageInput = (
  overrides: Partial<DependencyUsageInput> = {},
): DependencyUsageInput => ({
  packageDirectory: PACKAGE_DIR,
  repositoryRoot: PACKAGE_DIR,
  commandNames: [],
  pluginDependencies: [],
  packageDependencies: [],
  delegatedPlugins: new Set<string>(),
  ...overrides,
});

describe("checkDependencyUsage", () => {
  it("accepts a dependency a command delegates to", () => {
    const fs = createPackage("# Commands\n");

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-workflow"],
          packageDependencies: ["@xonovex/plugin-workflow"],
          delegatedPlugins: new Set(["xonovex-workflow"]),
        }),
        fs,
      ),
    ).toEqual([]);
  });

  it("accepts a routing dependency the README names", () => {
    const fs = createPackage(
      "# Commands\n\nThe core plugin supplies tdd and bdd guides.\n",
    );

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-core", "xonovex-core"],
        }),
        fs,
      ),
    ).toEqual([]);
  });

  it("accepts a routing dependency a command document names", () => {
    const fs = createPackage("# Commands\n", {
      "plan-research":
        "Use the core plugin for a read-only code-quality audit.\n",
    });

    expect(
      checkDependencyUsage(
        usageInput({
          commandNames: ["plan-research"],
          pluginDependencies: ["xonovex-core"],
        }),
        fs,
      ),
    ).toEqual([]);
  });

  it("reports a dependency nothing delegates to or names", () => {
    const fs = createPackage("# Commands\n\nNothing relevant here.\n");

    const issues = checkDependencyUsage(
      usageInput({
        pluginDependencies: ["xonovex-core"],
        packageDependencies: ["@xonovex/plugin-core"],
      }),
      fs,
    );

    expect(issues.map(({code}) => code)).toEqual([
      "dependency.plugin-unused",
      "dependency.package-unused",
    ]);
    expect(issues[0]).toMatchObject({
      severity: "error",
      path: join(".claude-plugin", "plugin.json"),
    });
    expect(issues[1]).toMatchObject({severity: "error", path: "package.json"});
    expect(issues[0]?.message).toContain("xonovex-core");
  });

  it("does not accept an install line that only names the plugin id", () => {
    const fs = createPackage(
      "# Commands\n\ncodex plugin add xonovex-core@xonovex-marketplace\n",
    );

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-core"],
        }),
        fs,
      ).map(({code}) => code),
    ).toEqual(["dependency.plugin-unused"]);
  });

  it("does not accept a hyphenated compound as a mention", () => {
    const fs = createPackage("# Commands\n\nSee `skill-create`.\n");

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-create"],
        }),
        fs,
      ).map(({code}) => code),
    ).toEqual(["dependency.plugin-unused"]);
  });

  it("matches a mention regardless of case", () => {
    const fs = createPackage("# Commands\n\n## Content\n");

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-content"],
        }),
        fs,
      ),
    ).toEqual([]);
  });

  it("ignores dependencies that are not skill plugins", () => {
    const fs = createPackage("# Commands\n");

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["some-other-plugin"],
          packageDependencies: ["zod"],
        }),
        fs,
      ),
    ).toEqual([]);
  });

  it("tolerates a package with no README", () => {
    const fs = memoryFileSystem({directories: [PACKAGE_DIR]});

    expect(
      checkDependencyUsage(
        usageInput({
          pluginDependencies: ["xonovex-workflow"],
          delegatedPlugins: new Set(["xonovex-workflow"]),
        }),
        fs,
      ),
    ).toEqual([]);
  });
});
