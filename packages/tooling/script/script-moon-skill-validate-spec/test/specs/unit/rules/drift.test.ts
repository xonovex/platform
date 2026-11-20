import {describe, expect, it} from "vitest";
import {checkDrift} from "../../../../src/rules/drift.js";

describe("checkDrift", () => {
  it("keeps a budget finding advisory in warn mode", () => {
    expect(
      checkDrift({findings: ["budget: over"], manifestErrors: []}, "warn"),
    ).toEqual([{level: "drift", message: "budget: over"}]);
  });

  it("fails a budget finding in enforce mode", () => {
    expect(
      checkDrift({findings: ["budget: over"], manifestErrors: []}, "enforce"),
    ).toEqual([{level: "fail", message: "budget: over"}]);
  });

  it("fails a manifest error in either mode", () => {
    expect(
      checkDrift({findings: [], manifestErrors: ["bad manifest"]}, "warn"),
    ).toEqual([{level: "fail", message: "bad manifest"}]);
  });

  it("passes a skill within its budgets", () => {
    expect(checkDrift({findings: [], manifestErrors: []}, "warn")).toEqual([
      {
        level: "pass",
        message: "drift: file budgets and vocabulary ownership hold",
      },
    ]);
  });
});
