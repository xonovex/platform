import {describe, expect, it} from "vitest";
import {main as validateLinks} from "../../../src/validate-links.js";

describe("link validator entrypoint", () => {
  it("renders link-validator help without scanning the repository", () => {
    expect(validateLinks(["--help"])).toBe(0);
  });
});
