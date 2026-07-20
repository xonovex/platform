import {describe, expect, it} from "vitest";
import {main as validateLinks} from "./validate-links.js";
import {main as validateSkill} from "./validate-skill.js";

describe("skill validator entrypoints", () => {
  it("renders validator help without reading a skill", () => {
    expect(validateSkill(["--help"])).toBe(0);
  });

  it("renders link-validator help without scanning the repository", () => {
    expect(validateLinks(["--help"])).toBe(0);
  });
});
