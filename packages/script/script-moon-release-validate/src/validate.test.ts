import {describe, expect, it} from "vitest";
import {
  createChecker,
  forbiddenClaims,
  markdownLinkTargets,
  tableIds,
} from "./validate.js";

describe("createChecker", () => {
  it("counts every check and collects only failures", () => {
    const checker = createChecker();
    checker.check(true, "passes");
    checker.check(false, "fails");
    checker.check(false, "also fails");
    expect(checker.checks()).toBe(3);
    expect(checker.failures()).toEqual(["fails", "also fails"]);
  });
});

describe("tableIds", () => {
  it("extracts prefixed ids from leading table cells only", () => {
    const content = [
      "| ID | Title |",
      "| --- | --- |",
      "| D-001 | first |",
      "| D-002 | mentions D-099 inline |",
      "| note | D-100 not in the leading cell |",
    ].join("\n");
    expect([...tableIds(content, "D-")]).toEqual(["D-001", "D-002"]);
  });
});

describe("markdownLinkTargets", () => {
  it("returns relative targets and skips external and anchor-only links", () => {
    const content =
      "[a](./local.md) [b](https://example.com) [c](mailto:x@y.z) " +
      "[d](#anchor) [e](../up.md#section)";
    expect(markdownLinkTargets(content)).toEqual(["./local.md", "../up.md"]);
  });
});

describe("forbiddenClaims", () => {
  it("matches overclaiming phrases and passes qualified statements", () => {
    const overclaim = "Installing a skill enforces the policy everywhere.";
    const qualified = "A skill documents the policy; enforcement is separate.";
    expect(forbiddenClaims.some((claim) => claim.test(overclaim))).toBe(true);
    expect(forbiddenClaims.some((claim) => claim.test(qualified))).toBe(false);
  });
});
