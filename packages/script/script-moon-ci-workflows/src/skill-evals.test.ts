import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  discoverSkills,
  outputEvalBatches,
  selectChangedSkills,
  triggerEvalBatches,
} from "./skill-evals.js";

const temporaryDirectories: string[] = [];

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("discoverSkills", () => {
  it("finds the single guide in each skill package", () => {
    const root = mkdtempSync(join(tmpdir(), "skill-eval-catalog-"));
    temporaryDirectories.push(root);
    const guide = join(root, "skill-testing", "testing-guide");
    mkdirSync(guide, {recursive: true});
    writeFileSync(join(guide, "SKILL.md"), "# Testing\n", "utf8");

    expect(discoverSkills(root)).toEqual([
      {package: "testing", guide: "testing-guide"},
    ]);
  });
});

describe("selectChangedSkills", () => {
  const skills = [
    {package: "astro", guide: "astro-guide"},
    {package: "testing", guide: "testing-guide"},
  ];

  it("selects only directly changed skills", () => {
    expect(
      selectChangedSkills(skills, [
        "packages/skill/skill-testing/testing-guide/SKILL.md",
      ]),
    ).toEqual([{package: "testing", guide: "testing-guide"}]);
  });

  it("selects the catalog when shared orchestration changes", () => {
    expect(
      selectChangedSkills(skills, [
        "packages/script/script-moon-ci-workflows/src/skill-evals.ts",
      ]),
    ).toEqual(skills);
  });
});

describe("triggerEvalBatches", () => {
  it("keeps validation queries and bounds each batch", () => {
    const queries = [
      {query: "train", split: "train"},
      ...Array.from({length: 9}, (_, index) => ({
        query: `validation-${String(index)}`,
        split: "validation",
      })),
    ];

    expect(triggerEvalBatches(JSON.stringify(queries))).toHaveLength(2);
    expect(triggerEvalBatches(JSON.stringify(queries))[0]).toHaveLength(8);
    expect(triggerEvalBatches(JSON.stringify(queries))[1]).toHaveLength(1);
  });
});

describe("outputEvalBatches", () => {
  it("preserves document metadata while bounding evals", () => {
    const document = {
      skill_name: "testing-guide",
      tier: "moderate",
      evals: Array.from({length: 7}, (_, index) => ({id: index + 1})),
    };

    const batches = outputEvalBatches(JSON.stringify(document));

    expect(batches).toHaveLength(2);
    expect(batches[0]).toMatchObject({
      skill_name: "testing-guide",
      tier: "moderate",
    });
    expect(batches[0]?.evals).toHaveLength(6);
    expect(batches[1]?.evals).toHaveLength(1);
  });
});
