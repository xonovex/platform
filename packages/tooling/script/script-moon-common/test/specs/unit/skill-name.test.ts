import {describe, expect, it} from "vitest";
import {
  boldSkillNameMentions,
  isSkillName,
  SKILL_NAME_RULE,
  skillNameMentions,
} from "../../../src/skill-name.js";

describe("skill names", () => {
  it("accepts a gerund followed by its object", () => {
    expect(isSkillName("using-git")).toBe(true);
    expect(isSkillName("writing-c99-opinionated")).toBe(true);
  });

  it("rejects a name without a leading gerund or without an object", () => {
    expect(isSkillName("git-skill")).toBe(false);
    expect(isSkillName("debugging")).toBe(false);
    expect(isSkillName("Using-git")).toBe(false);
    expect(isSkillName("using-git-")).toBe(false);
  });

  it("states the rule as a plain regular expression", () => {
    expect(SKILL_NAME_RULE).toBe("^[a-z0-9]+ing(-[a-z0-9]+)+$");
  });

  it("finds whole tokens that have the form of a skill name", () => {
    expect(
      skillNameMentions(
        "Hand off to Using-Git, not to a non-blocking-queue or git-skill.",
      ),
    ).toEqual(["using-git"]);
  });

  it("finds the skill names that Markdown sets in bold", () => {
    expect(
      boldSkillNameMentions(
        "**using-git** owns it; **Batching** and `writing-lua` do not count, **writing-lua** does.",
      ),
    ).toEqual(["using-git", "writing-lua"]);
  });

  it("reads a bold name before the word operation as an operation", () => {
    expect(
      boldSkillNameMentions(
        "Load **authoring-skills** and perform its **evaluating-outputs** operation.",
      ),
    ).toEqual(["authoring-skills"]);
  });
});
