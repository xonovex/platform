import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  checkDependencyUsage,
  type DependencyUsageInput,
} from "../../../src/dependency-usage.js";

const directories: string[] = [];

const createPackage = (
  readme: string,
  commands: Readonly<Record<string, string>> = {},
): string => {
  const directory = mkdtempSync(join(tmpdir(), "command-dependency-"));
  directories.push(directory);
  mkdirSync(join(directory, "commands"), {recursive: true});
  writeFileSync(join(directory, "README.md"), readme);
  for (const [name, body] of Object.entries(commands)) {
    writeFileSync(join(directory, "commands", `${name}.md`), body);
  }
  return directory;
};

const usageInput = (
  directory: string,
  overrides: Partial<DependencyUsageInput> = {},
): DependencyUsageInput => ({
  packageDirectory: directory,
  repositoryRoot: directory,
  commandNames: [],
  pluginDependencies: [],
  packageDependencies: [],
  delegatedPlugins: new Set<string>(),
  ...overrides,
});

afterEach(() => {
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
});

describe("checkDependencyUsage", () => {
  it("accepts a dependency a command delegates to", () => {
    const directory = createPackage("# Commands\n");

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["xonovex-skill-plan"],
          packageDependencies: ["@xonovex/skill-plan"],
          delegatedPlugins: new Set(["xonovex-skill-plan"]),
        }),
      ),
    ).toEqual([]);
  });

  it("accepts a routing dependency the README names", () => {
    const directory = createPackage(
      "# Commands\n\nTest-first plans route to tdd / bdd guides.\n",
    );

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["xonovex-skill-tdd", "xonovex-skill-bdd"],
        }),
      ),
    ).toEqual([]);
  });

  it("accepts a routing dependency a command document names", () => {
    const directory = createPackage("# Commands\n", {
      "plan-research": "Run a read-only code-quality audit.\n",
    });

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          commandNames: ["plan-research"],
          pluginDependencies: ["xonovex-skill-code-quality"],
        }),
      ),
    ).toEqual([]);
  });

  it("reports a dependency nothing delegates to or names", () => {
    const directory = createPackage("# Commands\n\nNothing relevant here.\n");

    const issues = checkDependencyUsage(
      usageInput(directory, {
        pluginDependencies: ["xonovex-skill-versioning"],
        packageDependencies: ["@xonovex/skill-versioning"],
      }),
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
    expect(issues[0]?.message).toContain("xonovex-skill-versioning");
  });

  it("does not accept an install line that only names the plugin id", () => {
    const directory = createPackage(
      "# Commands\n\ncodex plugin add xonovex-skill-versioning@xonovex-marketplace\n",
    );

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["xonovex-skill-versioning"],
        }),
      ).map(({code}) => code),
    ).toEqual(["dependency.plugin-unused"]);
  });

  it("does not accept a hyphenated compound as a mention", () => {
    const directory = createPackage("# Commands\n\nSee `skill-create`.\n");

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["xonovex-skill-create"],
        }),
      ).map(({code}) => code),
    ).toEqual(["dependency.plugin-unused"]);
  });

  it("matches a mention regardless of case", () => {
    const directory = createPackage("# Commands\n\n## Content\n");

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["xonovex-skill-content"],
        }),
      ),
    ).toEqual([]);
  });

  it("ignores dependencies that are not skill plugins", () => {
    const directory = createPackage("# Commands\n");

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["some-other-plugin"],
          packageDependencies: ["zod"],
        }),
      ),
    ).toEqual([]);
  });

  it("tolerates a package with no README", () => {
    const directory = mkdtempSync(join(tmpdir(), "command-dependency-none-"));
    directories.push(directory);

    expect(
      checkDependencyUsage(
        usageInput(directory, {
          pluginDependencies: ["xonovex-skill-plan"],
          delegatedPlugins: new Set(["xonovex-skill-plan"]),
        }),
      ),
    ).toEqual([]);
  });
});
