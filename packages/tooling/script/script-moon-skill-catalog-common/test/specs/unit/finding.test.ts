import {describe, expect, it} from "vitest";
import {drift, fail, pass, warn} from "../../../src/finding.js";

describe("finding constructors", () => {
  it("tag each message with its level", () => {
    expect([pass("a"), warn("b"), drift("c"), fail("d")]).toEqual([
      {level: "pass", message: "a"},
      {level: "warn", message: "b"},
      {level: "drift", message: "c"},
      {level: "fail", message: "d"},
    ]);
  });
});
