import {describe, expect, it} from "vitest";
import {parseArgs} from "./args.js";

describe("parseArgs", () => {
  it("parses positional and value options", () => {
    expect(
      parseArgs([
        "guide",
        "--max-age=30",
        "--version-max-age=14",
        "--fetch",
        "--pull",
        "--json",
        "--mark-reviewed",
        "Primary source",
      ]),
    ).toEqual({
      target: "guide",
      all: undefined,
      maxAge: 30,
      versionMaxAge: 14,
      fetch: true,
      markReviewed: "Primary source",
      json: true,
      pull: true,
      help: false,
    });
  });

  it("supports optional values", () => {
    expect(parseArgs(["--all", "--mark-reviewed"])).toMatchObject({
      all: ".",
      markReviewed: "",
    });
  });

  it.each([
    ["--max-age"],
    ["--max-age", "-1"],
    ["--max-age=1.5"],
    ["--version-max-age"],
    ["--version-max-age=-1"],
    ["--unknown"],
    ["one", "two"],
  ])("rejects invalid arguments: %o", (...argv) => {
    expect(() => parseArgs(argv)).toThrow();
  });

  it("records help without exiting the process", () => {
    expect(parseArgs(["--help"]).help).toBe(true);
  });
});
