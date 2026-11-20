import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  collectSkillCatalogFiles,
  collectSkillFiles,
  manifestPath,
} from "../../../src/drift-files.js";
import {checkSkillDrift} from "../../../src/drift-lints.js";

const words = (count: number) =>
  Array.from({length: count}, () => "word").join(" ");

const REPOSITORY_ROOT = "/repo";
const SKILL_DIRECTORY = join(
  REPOSITORY_ROOT,
  "packages",
  "runtime",
  "plugin",
  "plugin-demo",
  "skills",
  "using-demo",
);

// A miniature repository with one skill, which is the smallest tree the
// catalog walks recognize.
const catalogFixture = (
  skillText: string,
  referenceText = "reference prose",
  extra: Readonly<Record<string, string>> = {},
) =>
  memoryFileSystem({
    files: {
      [join(SKILL_DIRECTORY, "SKILL.md")]: skillText,
      [join(SKILL_DIRECTORY, "references", "topic.md")]: referenceText,
      ...extra,
    },
  });

describe("drift files", () => {
  it("keys files by repository-relative POSIX path", () => {
    expect(manifestPath("/repo", "/repo/packages/runtime/plugin/a/b.md")).toBe(
      "packages/runtime/plugin/a/b.md",
    );
  });

  it("collects the SKILL.md of a skill and references with their kinds", () => {
    const files = collectSkillFiles(
      SKILL_DIRECTORY,
      REPOSITORY_ROOT,
      catalogFixture("skill prose"),
    );

    expect(files.map(({kind, path}) => [kind, path])).toEqual([
      [
        "skill",
        "packages/runtime/plugin/plugin-demo/skills/using-demo/SKILL.md",
      ],
      [
        "reference",
        "packages/runtime/plugin/plugin-demo/skills/using-demo/references/topic.md",
      ],
    ]);
  });

  it("walks the skill catalog", () => {
    const fs = catalogFixture("skill prose");

    expect(collectSkillCatalogFiles(REPOSITORY_ROOT, fs)).toHaveLength(2);
  });

  it("returns nothing when the catalogs are absent", () => {
    const empty = memoryFileSystem({directories: [REPOSITORY_ROOT]});

    expect(collectSkillCatalogFiles(REPOSITORY_ROOT, empty)).toEqual([]);
  });

  it("ignores a catalog entry that holds no SKILL.md", () => {
    const fs = memoryFileSystem({
      files: {
        [join(
          REPOSITORY_ROOT,
          "packages",
          "runtime",
          "plugin",
          "plugin-demo",
          "README.md",
        )]: "prose",
        [join(
          REPOSITORY_ROOT,
          "packages",
          "runtime",
          "plugin",
          "plugin-demo",
          "skills",
          "drafting-notes",
          "references",
          "notes.md",
        )]: "a directory without a SKILL.md",
      },
    });

    expect(collectSkillCatalogFiles(REPOSITORY_ROOT, fs)).toEqual([]);
  });
});

describe("skill drift", () => {
  it("passes a skill within its caps", () => {
    expect(
      checkSkillDrift(
        SKILL_DIRECTORY,
        REPOSITORY_ROOT,
        catalogFixture("skill prose"),
      ),
    ).toEqual({findings: [], manifestErrors: []});
  });

  it("reports a skill over the skill cap", () => {
    const {findings} = checkSkillDrift(
      SKILL_DIRECTORY,
      REPOSITORY_ROOT,
      catalogFixture(words(950)),
    );

    expect(findings).toHaveLength(1);
    expect(findings[0]).toContain("900-word cap");
  });

  it("surfaces a malformed manifest separately from findings", () => {
    const {findings, manifestErrors} = checkSkillDrift(
      SKILL_DIRECTORY,
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
