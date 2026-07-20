import {describe, expect, it} from "vitest";
import {main as validateSkill} from "./index.js";
import {main as validateLinks} from "./links-cli.js";

describe("skill validator entrypoints", () => {
  it("renders validator help without reading a skill", () => {
    expect(validateSkill(["--help"])).toBe(0);
  });

  it("renders link-validator help without scanning the repository", () => {
    expect(validateLinks(["--help"])).toBe(0);
  });
});
