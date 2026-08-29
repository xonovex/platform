import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  collectCommandCatalogFiles,
  collectGuideFiles,
  collectSkillCatalogFiles,
  manifestPath,
} from "../../../src/drift-files.js";
import {checkGuideDrift} from "../../../src/drift-lints.js";

const words = (count: number) =>
  Array.from({length: count}, () => "word").join(" ");

const REPOSITORY_ROOT = "/repo";
const GUIDE_DIRECTORY = join(
  REPOSITORY_ROOT,
  "packages",
  "skill",
  "skill-demo",
  "demo-guide",
);

// A miniature repository with one guide and one command, which is the smallest
// tree the catalog walks recognize.
const catalogFixture = (
  skillText: string,
  referenceText = "reference prose",
  extra: Readonly<Record<string, string>> = {},
) =>
  memoryFileSystem({
    files: {
      [join(GUIDE_DIRECTORY, "SKILL.md")]: skillText,
      [join(GUIDE_DIRECTORY, "references", "topic.md")]: referenceText,
      [join(
        REPOSITORY_ROOT,
        "packages",
        "command",
        "command-demo",
        "commands",
        "run.md",
      )]: "command prose",
      ...extra,
    },
  });

describe("drift files", () => {
  it("keys files by repository-relative POSIX path", () => {
    expect(manifestPath("/repo", "/repo/packages/skill/a/b.md")).toBe(
      "packages/skill/a/b.md",
    );
  });

  it("collects a guide's SKILL.md and references with their kinds", () => {
    const files = collectGuideFiles(
      GUIDE_DIRECTORY,
      REPOSITORY_ROOT,
      catalogFixture("skill prose"),
    );

    expect(files.map(({kind, path}) => [kind, path])).toEqual([
      ["skill", "packages/skill/skill-demo/demo-guide/SKILL.md"],
      ["reference", "packages/skill/skill-demo/demo-guide/references/topic.md"],
    ]);
  });

  it("walks the skill and command catalogs", () => {
    const fs = catalogFixture("skill prose");

    expect(collectSkillCatalogFiles(REPOSITORY_ROOT, fs)).toHaveLength(2);
    expect(collectCommandCatalogFiles(REPOSITORY_ROOT, fs)).toEqual([
      {
        kind: "command",
        path: "packages/command/command-demo/commands/run.md",
        text: "command prose",
      },
    ]);
  });

  it("returns nothing when the catalogs are absent", () => {
    const empty = memoryFileSystem({directories: [REPOSITORY_ROOT]});

    expect(collectSkillCatalogFiles(REPOSITORY_ROOT, empty)).toEqual([]);
    expect(collectCommandCatalogFiles(REPOSITORY_ROOT, empty)).toEqual([]);
  });

  it("ignores a catalog entry that is not a guide directory", () => {
    const fs = memoryFileSystem({
      files: {
        [join(REPOSITORY_ROOT, "packages", "skill", "skill-demo", "README.md")]:
          "prose",
      },
    });

    expect(collectSkillCatalogFiles(REPOSITORY_ROOT, fs)).toEqual([]);
  });
});

describe("guide drift", () => {
  it("passes a guide within its caps", () => {
    expect(
      checkGuideDrift(
        GUIDE_DIRECTORY,
        REPOSITORY_ROOT,
        catalogFixture("skill prose"),
      ),
    ).toEqual({findings: [], manifestErrors: []});
  });

  it("reports a guide over the skill cap", () => {
    const {findings} = checkGuideDrift(
      GUIDE_DIRECTORY,
      REPOSITORY_ROOT,
      catalogFixture(words(950)),
    );

    expect(findings).toHaveLength(1);
    expect(findings[0]).toContain("900-word cap");
  });

  it("surfaces a malformed manifest separately from findings", () => {
    const {findings, manifestErrors} = checkGuideDrift(
      GUIDE_DIRECTORY,
      REPOSITORY_ROOT,
      catalogFixture("skill prose", "reference prose", {
        [join(REPOSITORY_ROOT, "budgets.json")]: "{",
      }),
    );

    expect(findings).toEqual([]);
    expect(manifestErrors).toHaveLength(1);
    expect(manifestErrors[0]).toContain("not valid JSON");
  });
});
