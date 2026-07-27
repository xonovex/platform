import {describe, expect, it} from "vitest";
import {descriptionRoutingErrors} from "./description-routing.js";

describe("descriptionRoutingErrors", () => {
  it("accepts positive routing and non-routing uses of skip", () => {
    expect(
      descriptionRoutingErrors(
        "Use when publishing packages with idempotent publishing (skip if the version exists). Triggers on release tasks.",
        "release-guide",
      ),
    ).toEqual([]);
  });

  it("rejects negative routing clauses and other skill names", () => {
    expect(
      descriptionRoutingErrors(
        "Use when refining stories. Triggers on INVEST. Skip executable specs (bdd-guide).",
        "user-stories-guide",
      ),
    ).toEqual([
      "description: routing must use positive triggers only; move skip/out-of-scope guidance to the body",
      "description: names other skill(s) bdd-guide; put by-name handoffs in the body",
    ]);
  });

  it("does not treat the skill's own name as a cross-skill handoff", () => {
    expect(
      descriptionRoutingErrors(
        "Use when validating release-guide metadata. Triggers on release-guide files.",
        "release-guide",
      ),
    ).toEqual([]);
  });
});
