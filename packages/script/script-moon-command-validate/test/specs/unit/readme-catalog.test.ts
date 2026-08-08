import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  checkReadmeCatalog,
  parseReadmeCommands,
} from "../../../src/readme-catalog.js";

const PACKAGE_DIR = "/repo/packages/command/pkg";

const readme = `# Utility Commands

Manage project instructions and create skills.

## Installation

| Plugin      | Description |
| ----------- | ----------- |
| \`not-here\` | An install row, not a command |

## Commands

### Instructions

| Command               | Description         |
| --------------------- | ------------------- |
| \`instructions-init\` | Create an AGENTS.md |

### Skills

| Command          | Description  |
| ---------------- | ------------ |
| \`skill-create\` | Create one   |

## Notes

| Command      | Description                  |
| ------------ | ---------------------------- |
| \`after-end\` | Outside the Commands section |
`;

const createPackage = (source: string): FileSystem =>
  memoryFileSystem({
    directories: [join(PACKAGE_DIR, "commands")],
    files: {[join(PACKAGE_DIR, "README.md")]: source},
  });

describe("parseReadmeCommands", () => {
  it("reads every table row under the Commands heading", () => {
    expect(parseReadmeCommands(readme)).toEqual(
      new Set(["instructions-init", "skill-create"]),
    );
  });

  it("ignores tables outside the Commands section", () => {
    const listed = parseReadmeCommands(readme);
    expect(listed?.has("not-here")).toBe(false);
    expect(listed?.has("after-end")).toBe(false);
  });

  it("reports no catalog when the README has no Commands heading", () => {
    expect(parseReadmeCommands("# Title\n\nNo tables here.\n")).toBeUndefined();
  });

  it("reads an empty catalog from a Commands heading with no table", () => {
    expect(
      parseReadmeCommands("# Title\n\n## Commands\n\nNone yet.\n"),
    ).toEqual(new Set());
  });
});

describe("checkReadmeCatalog", () => {
  it("accepts a README that matches the shipped commands", () => {
    const fs = createPackage(readme);

    expect(
      checkReadmeCatalog(
        PACKAGE_DIR,
        ["instructions-init", "skill-create"],
        PACKAGE_DIR,
        fs,
      ),
    ).toEqual([]);
  });

  it("reports a README entry whose command file was removed", () => {
    const fs = createPackage(readme);

    const issues = checkReadmeCatalog(
      PACKAGE_DIR,
      ["skill-create"],
      PACKAGE_DIR,
      fs,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({
      code: "readme.command-missing",
      severity: "error",
      path: "README.md",
    });
    expect(issues[0]?.message).toContain("instructions-init");
  });

  it("reports a shipped command the README never lists", () => {
    const fs = createPackage(readme);

    const issues = checkReadmeCatalog(
      PACKAGE_DIR,
      ["instructions-init", "skill-create", "skill-extract"],
      PACKAGE_DIR,
      fs,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({code: "readme.command-unlisted"});
    expect(issues[0]?.message).toContain("skill-extract");
  });

  it("skips a README that declares no command catalog", () => {
    const fs = createPackage("# Commands\n\nProse only.\n");

    expect(
      checkReadmeCatalog(PACKAGE_DIR, ["skill-create"], PACKAGE_DIR, fs),
    ).toEqual([]);
  });

  it("skips a package that ships no README", () => {
    const fs = memoryFileSystem({directories: [PACKAGE_DIR]});

    expect(
      checkReadmeCatalog(PACKAGE_DIR, ["skill-create"], PACKAGE_DIR, fs),
    ).toEqual([]);
  });
});
