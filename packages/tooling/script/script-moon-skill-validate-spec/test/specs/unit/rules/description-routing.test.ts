import {describe, expect, it} from "vitest";
import {descriptionRoutingErrors} from "../../../../src/rules/description-routing.js";

describe("descriptionRoutingErrors", () => {
  it("accepts positive routing and non-routing uses of skip", () => {
    expect(
      descriptionRoutingErrors(
        "Use when publishing packages with idempotent publishing (skip if the version exists). Triggers on release tasks.",
        "using-release",
      ),
    ).toEqual([]);
  });

  it("rejects negative routing clauses and other skill names", () => {
    expect(
      descriptionRoutingErrors(
        "Use when refining stories. Triggers on INVEST. Skip executable specs (applying-bdd).",
        "writing-user-stories",
      ),
    ).toEqual([
      "description: routing must use positive triggers only; move skip/out-of-scope guidance to the body",
      "description: names other skill(s) applying-bdd; put by-name handoffs in the body",
    ]);
  });

  it("does not treat the skill's own name as a cross-skill handoff", () => {
    expect(
      descriptionRoutingErrors(
        "Use when validating using-release metadata. Triggers on using-release files.",
        "using-release",
      ),
    ).toEqual([]);
  });
});
