import {describe, expect, it} from "vitest";
import {selectEvalMatrix} from "./list-eval-matrix.mjs";

const entries = ["alpha", "beta", "gamma", "omega"].map((name) => ({
  package: name,
  project: `skill-${name}`,
}));

describe("selectEvalMatrix", () => {
  it("selects changed skills and expands shared evaluator changes", () => {
    expect(
      selectEvalMatrix({
        entries,
        changedFiles: ["packages/skill/skill-beta/beta-guide/SKILL.md"],
        limit: 12,
        offset: 0,
      }),
    ).toEqual([entries[1]]);
    expect(
      selectEvalMatrix({
        entries,
        changedFiles: [
          "packages/script/script-moon-skill-eval-triggers/src/evaluate.ts",
        ],
        limit: 2,
        offset: 0,
      }),
    ).toEqual(entries.slice(0, 2));
  });

  it("rotates bounded scheduled slices", () => {
    expect(
      selectEvalMatrix({
        entries,
        changedFiles: undefined,
        limit: 2,
        offset: 3,
      }),
    ).toEqual([entries[3], entries[0]]);
  });
});
