import {describe, expect, it} from "vitest";
import {
  claudePointer,
  instructionDocFailures,
} from "../../../src/instruction-docs.js";

const paired = (path: string) => ({
  path,
  hasAgentsFile: true,
  claudeText: `${claudePointer}\n`,
});

describe("instruction doc pairing", () => {
  it("accepts a group whose AGENTS.md has a pointing CLAUDE.md beside it", () => {
    expect(
      instructionDocFailures(
        [paired("packages/skill"), paired("packages/script")],
        ["packages/skill", "packages/script"],
      ),
    ).toEqual([]);
  });

  it("reports an AGENTS.md with no CLAUDE.md beside it", () => {
    const failures = instructionDocFailures(
      [{path: "packages/script", hasAgentsFile: true, claudeText: undefined}],
      ["packages/script"],
    );

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain(
      "packages/script/AGENTS.md has no CLAUDE.md beside it",
    );
  });

  it("reports a CLAUDE.md that does not reference its sibling", () => {
    const failures = instructionDocFailures(
      [
        {
          path: "packages/config",
          hasAgentsFile: true,
          claudeText: "# Config\n\n- Pin every plugin exactly.\n",
        },
      ],
      ["packages/config"],
    );

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain("packages/config/CLAUDE.md must point at");
  });

  it("reports a package group that documents nothing", () => {
    const failures = instructionDocFailures(
      [{path: "packages/shared", hasAgentsFile: false, claudeText: undefined}],
      ["packages/shared"],
    );

    expect(failures).toEqual([
      "packages/shared must document the group in AGENTS.md",
    ]);
  });

  it("reports a required group that is absent from the scan", () => {
    expect(instructionDocFailures([], ["packages/asset"])).toEqual([
      "packages/asset must document the group in AGENTS.md",
    ]);
  });

  it("ignores a directory that carries no AGENTS.md and is not a group", () => {
    expect(
      instructionDocFailures(
        [
          {
            path: "packages/skill/skill-git",
            hasAgentsFile: false,
            claudeText: undefined,
          },
        ],
        [],
      ),
    ).toEqual([]);
  });
});
