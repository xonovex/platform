import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {validateCommandPackage} from "../../../src/command-validation.js";

const ROOT = "/repo";
const PACKAGE_DIR = join(ROOT, "packages", "command", "command-test");

const write = (fs: FileSystem, path: string, content: string): void => {
  fs.writeFile(join(ROOT, path), content);
};

const commandSource = (namespace = "xonovex-test"): string => `---
description: Run one test
allowed-tools:
  - Read
  - Skill
argument-hint: "<subject>"
---

# /${namespace}:run - Run

## Arguments

- \`subject\` (required): Subject.

## Delegation

Load the \`test-guide\` skill (plugin \`xonovex-skill-test\`) and perform its
**Run** operation with these arguments.
`;

// The parts a case can leave out, so a missing piece is expressed by not building
// it rather than by deleting it from a tree that was already built.
interface FixtureParts {
  readonly commands?: boolean;
  readonly skill?: boolean;
}

const fixture = ({
  commands = true,
  skill = true,
}: FixtureParts = {}): FileSystem => {
  const fs = memoryFileSystem();
  write(
    fs,
    "packages/command/command-test/package.json",
    JSON.stringify({
      dependencies: {"@xonovex/skill-test": "1.0.0"},
    }),
  );
  write(
    fs,
    "packages/command/command-test/.claude-plugin/plugin.json",
    JSON.stringify({
      name: "xonovex-test",
      dependencies: ["xonovex-skill-test"],
    }),
  );
  if (commands) {
    write(fs, "packages/command/command-test/commands/run.md", commandSource());
  }
  if (skill) {
    write(
      fs,
      "packages/skill/skill-test/test-guide/SKILL.md",
      "# Test\n\n- **Run**: run one test, see [run](references/run.md)\n",
    );
    write(
      fs,
      "packages/skill/skill-test/test-guide/references/run.md",
      "# Run\n",
    );
  }
  return fs;
};

describe("validateCommandPackage", () => {
  it("accepts a thin command with declared hard delegation", () => {
    const fs = fixture();

    expect(validateCommandPackage(PACKAGE_DIR, ROOT, fs).report).toEqual({
      commands: 1,
      issues: [],
    });
  });

  it("reports an em dash or ellipsis in any package prose file", () => {
    const fs = fixture();
    write(
      fs,
      "packages/command/command-test/commands/run.md",
      commandSource().replace("Subject.", "Subject \u2014 the thing to run."),
    );
    write(
      fs,
      "packages/command/command-test/README.md",
      "# Test\n\nUse when\u2026\n",
    );
    write(
      fs,
      "packages/command/command-test/.claude-plugin/plugin.json",
      String.raw`{"name": "xonovex-test", "dependencies": ["xonovex-skill-test"], "description": "a \u2014 b"}`,
    );

    const findings = validateCommandPackage(PACKAGE_DIR, ROOT, fs)
      .report.issues.filter(({code}) => code === "prose.punctuation")
      .map(({message, path}) => [path, message.split(",", 1)[0]]);

    expect(findings).toEqual([
      [
        join(
          "packages",
          "command",
          "command-test",
          ".claude-plugin",
          "plugin.json",
        ),
        "escaped em dash on line 1",
      ],
      [
        join("packages", "command", "command-test", "README.md"),
        "ellipsis on line 3",
      ],
      [
        join("packages", "command", "command-test", "commands", "run.md"),
        "em dash on line 13",
      ],
    ]);
  });

  it("reports namespace, filename, manifest, and skill drift", () => {
    const fs = fixture({skill: false});
    write(
      fs,
      "packages/command/command-test/package.json",
      JSON.stringify({dependencies: {}}),
    );
    write(
      fs,
      "packages/command/command-test/.claude-plugin/plugin.json",
      JSON.stringify({name: "other", dependencies: []}),
    );
    write(
      fs,
      "packages/command/command-test/commands/wrong.md",
      commandSource(),
    );

    const codes = validateCommandPackage(
      PACKAGE_DIR,
      ROOT,
      fs,
    ).report.issues.map(({code}) => code);

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

  it("rejects a delegated operation that the owner does not register", () => {
    const fs = fixture();
    write(fs, "packages/skill/skill-test/test-guide/SKILL.md", "# Test\n");

    expect(
      validateCommandPackage(PACKAGE_DIR, ROOT, fs).report.issues,
    ).toContainEqual(
      expect.objectContaining({code: "delegation.operation-missing"}),
    );
  });

  it("reports a missing commands directory", () => {
    const fs = fixture({commands: false});

    expect(validateCommandPackage(PACKAGE_DIR, ROOT, fs).report.issues).toEqual(
      [expect.objectContaining({code: "package.commands-missing"})],
    );
  });

  it("rejects drift between argument-hint and the Arguments section", () => {
    const fs = fixture();
    write(
      fs,
      "packages/command/command-test/commands/run.md",
      commandSource()
        .replace(
          'argument-hint: "<subject>"',
          'argument-hint: "<subject> [--extra]"',
        )
        .replace(
          "- `subject` (required): Subject.",
          "- `documented` (required): Subject.",
        ),
    );

    const codes = validateCommandPackage(
      PACKAGE_DIR,
      ROOT,
      fs,
    ).report.issues.map(({code}) => code);

    expect(codes).toEqual(
      expect.arrayContaining([
        "command.argument-undocumented",
        "command.argument-missing-hint",
      ]),
    );
  });

  it("allows a positional and a flag alias but rejects a repeated argument", () => {
    const fs = fixture();
    write(
      fs,
      "packages/command/command-test/commands/run.md",
      commandSource().replace(
        'argument-hint: "<subject>"',
        'argument-hint: "<subject> [--subject <subject>]"',
      ),
    );
    expect(
      validateCommandPackage(PACKAGE_DIR, ROOT, fs).report.issues.map(
        ({code}) => code,
      ),
    ).not.toContain("command.argument-duplicate");

    write(
      fs,
      "packages/command/command-test/commands/run.md",
      commandSource()
        .replace(
          'argument-hint: "<subject>"',
          'argument-hint: "<subject> [--extra] [--extra]"',
        )
        .replace(
          "- `subject` (required): Subject.",
          "- `subject` (required): Subject.\n- `--extra` (optional): Extra mode.",
        ),
    );
    expect(
      validateCommandPackage(PACKAGE_DIR, ROOT, fs).report.issues.map(
        ({code}) => code,
      ),
    ).toContain("command.argument-duplicate");
  });

  it("rejects obsolete machine-readable soft requirements", () => {
    const fs = fixture();
    write(
      fs,
      "packages/command/command-test/commands/run.md",
      commandSource().replace(
        "## Delegation",
        [
          "## Requirements",
          "",
          "- `assurance:evidence@^1.0.0` (preferred): Specialist evidence.",
          "",
          "## Delegation",
        ].join("\n"),
      ),
    );
    expect(
      validateCommandPackage(PACKAGE_DIR, ROOT, fs).report.issues,
    ).toContainEqual(
      expect.objectContaining({
        code: "command.requirements-unsupported",
        severity: "error",
      }),
    );
  });

  it("validates local Markdown targets and fragments", () => {
    const fs = fixture();
    write(
      fs,
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
      fs,
      "packages/command/command-test/docs/guide.md",
      "# Guide\n\n## Details\n",
    );

    expect(
      validateCommandPackage(PACKAGE_DIR, ROOT, fs).report.issues.map(
        ({code}) => code,
      ),
    ).toEqual(["link.missing-target", "link.missing-fragment"]);
  });
});
