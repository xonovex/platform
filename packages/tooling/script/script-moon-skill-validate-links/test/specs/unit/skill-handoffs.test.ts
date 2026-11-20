import {describe, expect, it} from "vitest";
import {checkNamedSkillHandoffs} from "../../../src/skill-handoffs.js";

const SKILL = "/repo/packages/runtime/plugin/plugin-x/skills/using-x/SKILL.md";

describe("checkNamedSkillHandoffs", () => {
  it("resolves catalog skill names and rejects missing handoffs", () => {
    expect(
      checkNamedSkillHandoffs(
        [{path: SKILL, text: "Use **using-x** and **using-missing**.\n"}],
        "/repo",
      ),
    ).toEqual([
      {
        level: "fail",
        message:
          "skill handoffs: packages/runtime/plugin/plugin-x/skills/using-x/SKILL.md names missing **using-missing**",
      },
    ]);
  });

  it("passes when every named handoff resolves", () => {
    expect(
      checkNamedSkillHandoffs(
        [{path: SKILL, text: "Use **using-x**.\n"}],
        "/repo",
      ),
    ).toEqual([
      {
        level: "pass",
        message: "skill handoffs: 1/1 named handoff(s) resolve",
      },
    ]);
  });

  it("reports nothing without a named handoff", () => {
    expect(
      checkNamedSkillHandoffs([{path: SKILL, text: "# X\n"}], "/repo"),
    ).toEqual([]);
  });
});
