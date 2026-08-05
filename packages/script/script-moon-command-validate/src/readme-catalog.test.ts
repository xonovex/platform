import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {checkReadmeCatalog, parseReadmeCommands} from "./readme-catalog.js";

const directories: string[] = [];

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

const createPackage = (source: string): string => {
  const directory = mkdtempSync(join(tmpdir(), "command-readme-"));
  directories.push(directory);
  mkdirSync(join(directory, "commands"), {recursive: true});
  writeFileSync(join(directory, "README.md"), source);
  return directory;
};

afterEach(() => {
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
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
    const directory = createPackage(readme);

    expect(
      checkReadmeCatalog(
        directory,
        ["instructions-init", "skill-create"],
        directory,
      ),
    ).toEqual([]);
  });

  it("reports a README entry whose command file was removed", () => {
    const directory = createPackage(readme);

    const issues = checkReadmeCatalog(directory, ["skill-create"], directory);

    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({
      code: "readme.command-missing",
      severity: "error",
      path: "README.md",
    });
    expect(issues[0]?.message).toContain("instructions-init");
  });

  it("reports a shipped command the README never lists", () => {
    const directory = createPackage(readme);

    const issues = checkReadmeCatalog(
      directory,
      ["instructions-init", "skill-create", "skill-extract"],
      directory,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]).toMatchObject({code: "readme.command-unlisted"});
    expect(issues[0]?.message).toContain("skill-extract");
  });

  it("skips a README that declares no command catalog", () => {
    const directory = createPackage("# Commands\n\nProse only.\n");

    expect(checkReadmeCatalog(directory, ["skill-create"], directory)).toEqual(
      [],
    );
  });

  it("skips a package that ships no README", () => {
    const directory = mkdtempSync(join(tmpdir(), "command-readme-none-"));
    directories.push(directory);

    expect(checkReadmeCatalog(directory, ["skill-create"], directory)).toEqual(
      [],
    );
  });
});
