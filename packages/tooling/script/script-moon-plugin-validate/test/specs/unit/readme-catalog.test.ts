import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {
  checkReadmeCatalog,
  parseReadmeSkills,
} from "../../../src/readme-catalog.js";

const PACKAGE_DIR = "/repo/packages/runtime/plugin/pkg";

const readme = `# Utility Plugin

Manage project instructions and create skills.

## Installation

| Plugin      | Description |
| ----------- | ----------- |
| [not-here](skills/not-here/SKILL.md) | An install row, not a skill |

## Skills

### Instructions

| Skill | Use when |
| --- | --- |
| [authoring-instructions](skills/authoring-instructions/SKILL.md) | Writing AGENTS.md |

### Skills

| Skill | Use when |
| --- | --- |
| [authoring-skills](skills/authoring-skills/SKILL.md) | Writing a skill |
| [mismatched](skills/other-name/SKILL.md) | The text differs from the path |
| \`plain-name\` | Not a link |

## Notes

| Skill | Use when |
| --- | --- |
| [after-end](skills/after-end/SKILL.md) | Outside the Skills section |
`;

const createPackage = (source: string): FileSystem =>
  memoryFileSystem({
    directories: [PACKAGE_DIR],
    files: {[join(PACKAGE_DIR, "README.md")]: source},
  });

describe("parseReadmeSkills", () => {
  it("reads every linked table row under the Skills heading", () => {
    expect(parseReadmeSkills(readme)).toEqual(
      new Set(["authoring-instructions", "authoring-skills"]),
    );
  });

  it("ignores tables outside the Skills section", () => {
    const listed = parseReadmeSkills(readme);
    expect(listed.has("not-here")).toBe(false);
    expect(listed.has("after-end")).toBe(false);
  });

  it("ignores a row whose link text differs from its path or that is no link", () => {
    const listed = parseReadmeSkills(readme);
    expect(listed.has("mismatched")).toBe(false);
    expect(listed.has("other-name")).toBe(false);
    expect(listed.has("plain-name")).toBe(false);
  });

  it("lists nothing when the README has no Skills heading", () => {
    expect(parseReadmeSkills("# Title\n\nNo tables here.\n")).toEqual(
      new Set(),
    );
  });

  it("lists nothing for a Skills heading with no table", () => {
    expect(parseReadmeSkills("# Title\n\n## Skills\n\nNone yet.\n")).toEqual(
      new Set(),
    );
  });
});

describe("checkReadmeCatalog", () => {
  it("accepts a README that matches the shipped skills", () => {
    const fs = createPackage(readme);

    expect(
      checkReadmeCatalog(
        PACKAGE_DIR,
        ["authoring-instructions", "authoring-skills"],
        PACKAGE_DIR,
        fs,
      ),
    ).toEqual([]);
  });

  it("reports a README entry whose skill directory was removed", () => {
    const fs = createPackage(readme);

    const issues = checkReadmeCatalog(
      PACKAGE_DIR,
      ["authoring-skills"],
      PACKAGE_DIR,
      fs,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({
      code: "readme.skill-missing",
      severity: "error",
      path: "README.md",
    });
    expect(issues[0]?.message).toContain("authoring-instructions");
  });

  it("reports a shipped skill the README never lists", () => {
    const fs = createPackage(readme);

    const issues = checkReadmeCatalog(
      PACKAGE_DIR,
      ["authoring-instructions", "authoring-skills", "using-git"],
      PACKAGE_DIR,
      fs,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({code: "readme.skill-unlisted"});
    expect(issues[0]?.message).toContain("using-git");
  });

  it("reports every shipped skill when the README has no Skills heading", () => {
    const fs = createPackage("# Plugin\n\nProse only.\n");

    expect(
      checkReadmeCatalog(PACKAGE_DIR, ["using-git"], PACKAGE_DIR, fs).map(
        ({code}) => code,
      ),
    ).toEqual(["readme.skill-unlisted"]);
  });

  it("skips a package that ships no README", () => {
    const fs = memoryFileSystem({directories: [PACKAGE_DIR]});

    expect(
      checkReadmeCatalog(PACKAGE_DIR, ["using-git"], PACKAGE_DIR, fs),
    ).toEqual([]);
  });
});
