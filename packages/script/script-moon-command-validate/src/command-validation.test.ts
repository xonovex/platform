import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {validateCommandPackage} from "./command-validation.js";

const temporaryDirectories: string[] = [];

const write = (root: string, path: string, content: string): void => {
  const target = join(root, path);
  mkdirSync(dirname(target), {recursive: true});
  writeFileSync(target, content);
};

const commandSource = (namespace = "xonovex-test"): string => `---
description: Run one test
allowed-tools:
  - Read
  - Skill
argument-hint: "<subject>"
---

# /${namespace}:run — Run

## Arguments

- \`subject\` (required): Subject.

## Delegation

Load the \`test-guide\` skill (plugin \`xonovex-skill-test\`) and perform its
**Run** operation with these arguments.
`;

const fixture = (): {packageDir: string; root: string} => {
  const root = mkdtempSync(join(tmpdir(), "command-validation-"));
  temporaryDirectories.push(root);
  const packageDir = join(root, "packages", "command", "command-test");
  write(
    root,
    "packages/command/command-test/package.json",
    JSON.stringify({
      dependencies: {"@xonovex/skill-test": "1.0.0"},
    }),
  );
  write(
    root,
    "packages/command/command-test/.claude-plugin/plugin.json",
    JSON.stringify({
      name: "xonovex-test",
      dependencies: ["xonovex-skill-test"],
    }),
  );
  write(root, "packages/command/command-test/commands/run.md", commandSource());
  write(root, "packages/skill/skill-test/test-guide/SKILL.md", "# Test\n");
  return {packageDir, root};
};

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("validateCommandPackage", () => {
  it("accepts a thin command with declared hard delegation", () => {
    const {packageDir, root} = fixture();

    expect(validateCommandPackage(packageDir, root).report).toEqual({
      commands: 1,
      issues: [],
    });
  });

  it("reports namespace, filename, manifest, and skill drift", () => {
    const {packageDir, root} = fixture();
    write(
      root,
      "packages/command/command-test/package.json",
      JSON.stringify({dependencies: {}}),
    );
    write(
      root,
      "packages/command/command-test/.claude-plugin/plugin.json",
      JSON.stringify({name: "other", dependencies: []}),
    );
    write(
      root,
      "packages/command/command-test/commands/wrong.md",
      commandSource(),
    );
    rmSync(join(root, "packages", "skill", "skill-test", "test-guide"), {
      recursive: true,
    });

    const codes = validateCommandPackage(packageDir, root).report.issues.map(
      ({code}) => code,
    );

    expect(codes).toEqual(
      expect.arrayContaining([
        "command.filename",
        "command.namespace",
        "delegation.plugin-dependency",
        "delegation.package-dependency",
        "delegation.skill-missing",
      ]),
    );
  });

  it("reports a missing commands directory", () => {
    const {packageDir, root} = fixture();
    rmSync(join(packageDir, "commands"), {recursive: true});

    expect(validateCommandPackage(packageDir, root).report.issues).toEqual([
      expect.objectContaining({code: "package.commands-missing"}),
    ]);
  });

  it("validates local Markdown targets and fragments", () => {
    const {packageDir, root} = fixture();
    write(
      root,
      "packages/command/command-test/README.md",
      [
        "# Commands",
        "",
        "[valid](docs/guide.md#details)",
        "[missing](docs/missing.md)",
        "[fragment](docs/guide.md#unknown)",
        "[cross-package](../../skill/missing.md)",
      ].join("\n"),
    );
    write(
      root,
      "packages/command/command-test/docs/guide.md",
      "# Guide\n\n## Details\n",
    );

    expect(
      validateCommandPackage(packageDir, root).report.issues.map(
        ({code}) => code,
      ),
    ).toEqual(["link.missing-target", "link.missing-fragment"]);
  });
});
