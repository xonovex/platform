import {describe, expect, it} from "vitest";
import {
  alternateFutureSplits,
  roundRobinCandidates,
  selectSiblingNames,
} from "./complete-trigger-evals.mjs";

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

describe("alternateFutureSplits", () => {
  it("interleaves queries that will become train and validation cases", () => {
    expect(
      alternateFutureSplits([
        "train 1",
        "train 2",
        "train 3",
        "train 4",
        "train 5",
        "validation 1",
        "validation 2",
        "validation 3",
      ]),
    ).toEqual([
      "train 1",
      "validation 1",
      "train 2",
      "validation 2",
      "train 3",
      "validation 3",
      "train 4",
      "train 5",
    ]);
  });
});

describe("selectSiblingNames", () => {
  it("prefers semantically adjacent catalog families", () => {
    const skills = [
      {
        name: "android-wcag-guide",
        description: "Android Compose accessibility and TalkBack",
        body: "",
      },
      {
        name: "android-analytics-guide",
        description: "Android Compose analytics tracking",
        body: "",
      },
      {
        name: "accessibility-guide",
        description: "Cross-platform accessibility assurance",
        body: "",
      },
      {
        name: "terraform-guide",
        description: "Terraform infrastructure modules",
        body: "",
      },
    ];

    expect(selectSiblingNames(skills[0], skills)).toEqual([
      "android-analytics-guide",
      "accessibility-guide",
    ]);
  });

  it("prioritizes explicit by-name handoffs", () => {
    const skills = [
      {
        name: "alpha-guide",
        description: "Alpha authoring",
        body: "Use **beta-guide** for the neighboring concern.",
      },
      {name: "beta-guide", description: "Beta authoring", body: ""},
      {name: "gamma-guide", description: "Gamma runtime", body: ""},
    ];

    expect(selectSiblingNames(skills[0], skills)).toEqual(["beta-guide"]);
  });
});
