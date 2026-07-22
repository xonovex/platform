import {describe, expect, it} from "vitest";
import {roundRobinCandidates} from "./complete-trigger-evals.mjs";

describe("roundRobinCandidates", () => {
  it("takes one near miss from each sibling before reusing a sibling", () => {
    const candidates = roundRobinCandidates([
      {name: "alpha-guide", queries: ["alpha 1", "alpha 2"]},
      {name: "beta-guide", queries: ["beta 1", "beta 2"]},
      {name: "gamma-guide", queries: ["gamma 1", "gamma 2"]},
    ]);

    expect(candidates.map(({query}) => query)).toEqual([
      "alpha 1",
      "beta 1",
      "gamma 1",
      "alpha 2",
      "beta 2",
      "gamma 2",
    ]);
    expect(candidates.map(({owner}) => owner)).toEqual([
      "alpha-guide",
      "beta-guide",
      "gamma-guide",
      "alpha-guide",
      "beta-guide",
      "gamma-guide",
    ]);
  });
});
