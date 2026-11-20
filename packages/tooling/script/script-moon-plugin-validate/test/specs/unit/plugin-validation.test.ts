import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {validatePluginPackage} from "../../../src/plugin-validation.js";

const ROOT = "/repo";
const PACKAGE_DIR = join(ROOT, "packages", "runtime", "plugin", "plugin-test");

const write = (fs: FileSystem, path: string, content: string): void => {
  fs.writeFile(join(PACKAGE_DIR, path), content);
};

const relativePath = (...parts: readonly string[]): string =>
  join("packages", "runtime", "plugin", "plugin-test", ...parts);

const README = `# Test

Run the tests of a package.

## Skills

| Skill | Use when |
| --- | --- |
| [using-test](skills/using-test/SKILL.md) | Running a test |
`;

const SKILL = "# Test\n\nCommits belong to **using-git**.\n";

const fixture = (): FileSystem => {
  const fs = memoryFileSystem();
  write(
    fs,
    "package.json",
    JSON.stringify({dependencies: {"@xonovex/plugin-core": "1.0.0"}}),
  );
  write(
    fs,
    ".claude-plugin/plugin.json",
    JSON.stringify({name: "xonovex-test", dependencies: ["xonovex-core"]}),
  );
  write(fs, "README.md", README);
  write(fs, "skills/using-test/SKILL.md", SKILL);
  fs.writeFile(
    join(ROOT, "packages/runtime/plugin/plugin-core/skills/using-git/SKILL.md"),
    "# Git\n",
  );
  return fs;
};

const codes = (fs: FileSystem): readonly string[] =>
  validatePluginPackage(PACKAGE_DIR, ROOT, fs).issues.map(({code}) => code);

describe("validatePluginPackage", () => {
  it("accepts a plugin whose catalog, links, and dependencies hold", () => {
    expect(validatePluginPackage(PACKAGE_DIR, ROOT, fixture())).toEqual({
      skills: 1,
      issues: [],
    });
  });

  it("counts only directories that hold a SKILL.md", () => {
    const fs = fixture();
    write(fs, "skills/notes/README.md", "# Not a skill\n");

    expect(validatePluginPackage(PACKAGE_DIR, ROOT, fs).skills).toBe(1);
  });

  it("reports a commands directory", () => {
    const fs = fixture();
    write(fs, "commands/run.md", "# Run\n");

    expect(validatePluginPackage(PACKAGE_DIR, ROOT, fs).issues).toContainEqual(
      expect.objectContaining({
        code: "package.commands-present",
        severity: "error",
        path: join("packages", "runtime", "plugin", "plugin-test", "commands"),
      }),
    );
  });

  it("reports README and skill directory drift", () => {
    const fs = fixture();
    write(fs, "skills/using-other/SKILL.md", "# Other\n");
    write(
      fs,
      "README.md",
      README.replace(
        "| [using-test]",
        "| [using-gone](skills/using-gone/SKILL.md) | Gone |\n| [using-test]",
      ),
    );

    expect(codes(fs)).toEqual([
      "link.missing-target",
      "readme.skill-missing",
      "readme.skill-unlisted",
    ]);
  });

  it("reports a skill dependency that no skill of the plugin uses", () => {
    const fs = fixture();
    write(fs, "skills/using-test/SKILL.md", "# Test\n");

    expect(codes(fs)).toEqual([
      "dependency.plugin-unused",
      "dependency.package-unused",
    ]);
  });

  it("reports an unreadable manifest", () => {
    const fs = fixture();
    write(fs, "package.json", "{");

    expect(codes(fs)).toEqual(["manifest.invalid"]);
  });

  it("reports an em dash or ellipsis in any package prose file", () => {
    const fs = fixture();
    write(fs, "skills/using-test/SKILL.md", "# Test\n\nRun — now.\n");
    write(fs, "README.md", `${README}\nUse when…\n`);
    write(
      fs,
      ".claude-plugin/plugin.json",
      '{"name": "xonovex-test", "dependencies": ["xonovex-core"], "description": "a \\u2014 b"}',
    );

    const findings = validatePluginPackage(PACKAGE_DIR, ROOT, fs)
      .issues.filter(({code}) => code === "prose.punctuation")
      .map(({message, path}) => [path, message.split(",", 1)[0]]);

    expect(findings).toEqual([
      [
        relativePath(".claude-plugin", "plugin.json"),
        "escaped em dash on line 1",
      ],
      [relativePath("README.md"), "ellipsis on line 11"],
      [relativePath("skills", "using-test", "SKILL.md"), "em dash on line 3"],
    ]);
  });

  it("validates local Markdown targets and fragments in the README and docs", () => {
    const fs = fixture();
    write(
      fs,
      "README.md",
      [
        README,
        "[valid](docs/guide.md#details)",
        "[missing](docs/missing.md)",
        "[fragment](docs/guide.md#unknown)",
        "[cross-package](../../skill/missing.md)",
        "[external](https://example.com/a.md)",
        "[anchor](#skills)",
      ].join("\n"),
    );
    write(
      fs,
      "docs/guide.md",
      "# Guide\n\n## Details\n\n[gone](other.md)\n\nCore plugin notes.\n",
    );

    expect(codes(fs)).toEqual([
      "link.missing-target",
      "link.missing-fragment",
      "link.missing-target",
    ]);
  });

  it("does not resolve links inside skills", () => {
    const fs = fixture();
    write(fs, "skills/using-test/SKILL.md", `${SKILL}\n[gone](missing.md)\n`);

    expect(codes(fs)).toEqual([]);
  });
});
