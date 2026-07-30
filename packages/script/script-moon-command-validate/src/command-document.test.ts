import {describe, expect, it} from "vitest";
import {parseCommandDocument} from "./command-document.js";

const validCommand = `---
description: Review one subject
allowed-tools:
  - Read
  - Skill
argument-hint: "<subject> [--perspective <selection>...]"
---

# /xonovex-test:review - Review

## Arguments

- \`subject\` (required): Subject.
- \`--perspective\`, \`--role\` (repeatable, optional): Perspective.
- \`--one, --two, --three\` (optional): Grouped flags.

## Delegation

Load the \`test-guide\` skill (plugin \`xonovex-skill-test\`) and perform its
**Review** operation with these arguments.
`;

describe("parseCommandDocument", () => {
  it("parses the public command contract and delegation", () => {
    const result = parseCommandDocument("commands/review.md", validCommand);

    expect(result.issues).toEqual([]);
    expect(result.document).toMatchObject({
      command: "review",
      namespace: "xonovex-test",
      delegation: {
        operation: "Review",
        plugin: "xonovex-skill-test",
        skill: "test-guide",
      },
    });
    expect(result.document?.documentedArguments).toEqual(
      new Set(["subject", "perspective", "role", "one", "two", "three"]),
    );
  });

  it("reports the existing unquoted-colon frontmatter failure class", () => {
    const result = parseCommandDocument(
      "commands/bad.md",
      validCommand.replace(
        "description: Review one subject",
        "description: Optimize in phases: inspect and apply",
      ),
    );

    expect(result.document).toBeUndefined();
    expect(result.issues).toEqual([
      expect.objectContaining({code: "frontmatter.invalid-yaml"}),
    ]);
  });

  it("reports missing structure and a delegation without Skill permission", () => {
    const result = parseCommandDocument(
      "commands/review.md",
      validCommand
        .replace("  - Skill\n", "")
        .replace("## Arguments", "## Input"),
    );

    expect(result.issues).toEqual(
      expect.arrayContaining([
        expect.objectContaining({code: "command.heading"}),
        expect.objectContaining({code: "command.delegation-tool"}),
      ]),
    );
  });

  it("rejects machine-readable soft requirements", () => {
    const withRequirements = validCommand.replace(
      "## Delegation",
      [
        "## Requirements",
        "",
        "- `assurance:evidence@^1.0.0` (preferred): Specialist evidence.",
        "",
        "## Delegation",
      ].join("\n"),
    );

    const result = parseCommandDocument("commands/review.md", withRequirements);

    expect(result.issues).toContainEqual(
      expect.objectContaining({code: "command.requirements-unsupported"}),
    );
  });
});
