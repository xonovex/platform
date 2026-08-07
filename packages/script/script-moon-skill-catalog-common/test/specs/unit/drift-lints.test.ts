import {mkdirSync, mkdtempSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
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

// catalogFixture builds a miniature repository with one guide and one command.
const catalogFixture = (
  skillText: string,
  referenceText = "reference prose",
): {readonly guideDirectory: string; readonly repositoryRoot: string} => {
  const repositoryRoot = mkdtempSync(join(tmpdir(), "drift-lints-"));
  const guideDirectory = join(
    repositoryRoot,
    "packages",
    "skill",
    "skill-demo",
    "demo-guide",
  );
  mkdirSync(join(guideDirectory, "references"), {recursive: true});
  writeFileSync(join(guideDirectory, "SKILL.md"), skillText, "utf8");
  writeFileSync(
    join(guideDirectory, "references", "topic.md"),
    referenceText,
    "utf8",
  );
  const commandDirectory = join(
    repositoryRoot,
    "packages",
    "command",
    "command-demo",
    "commands",
  );
  mkdirSync(commandDirectory, {recursive: true});
  writeFileSync(join(commandDirectory, "run.md"), "command prose", "utf8");
  return {guideDirectory, repositoryRoot};
};

describe("drift files", () => {
  it("keys files by repository-relative POSIX path", () => {
    expect(manifestPath("/repo", "/repo/packages/skill/a/b.md")).toBe(
      "packages/skill/a/b.md",
    );
  });

  it("collects a guide's SKILL.md and references with their kinds", () => {
    const {guideDirectory, repositoryRoot} = catalogFixture("skill prose");

    const files = collectGuideFiles(guideDirectory, repositoryRoot);

    expect(files.map(({kind, path}) => [kind, path])).toEqual([
      ["skill", "packages/skill/skill-demo/demo-guide/SKILL.md"],
      ["reference", "packages/skill/skill-demo/demo-guide/references/topic.md"],
    ]);
  });

  it("walks the skill and command catalogs", () => {
    const {repositoryRoot} = catalogFixture("skill prose");

    expect(collectSkillCatalogFiles(repositoryRoot)).toHaveLength(2);
    expect(collectCommandCatalogFiles(repositoryRoot)).toEqual([
      {
        kind: "command",
        path: "packages/command/command-demo/commands/run.md",
        text: "command prose",
      },
    ]);
  });

  it("returns nothing when the catalogs are absent", () => {
    const empty = mkdtempSync(join(tmpdir(), "drift-empty-"));

    expect(collectSkillCatalogFiles(empty)).toEqual([]);
    expect(collectCommandCatalogFiles(empty)).toEqual([]);
  });
});

describe("guide drift", () => {
  it("passes a guide within its caps", () => {
    const {guideDirectory, repositoryRoot} = catalogFixture("skill prose");

    expect(checkGuideDrift(guideDirectory, repositoryRoot)).toEqual({
      findings: [],
      manifestErrors: [],
    });
  });

  it("reports a guide over the skill cap", () => {
    const {guideDirectory, repositoryRoot} = catalogFixture(words(950));

    const {findings} = checkGuideDrift(guideDirectory, repositoryRoot);

    expect(findings).toHaveLength(1);
    expect(findings[0]).toContain("900-word cap");
  });

  it("reports a guide redefining an owned term", () => {
    const {guideDirectory, repositoryRoot} = catalogFixture(
      "skill prose",
      "## Handoff\n",
    );
    writeFileSync(
      join(repositoryRoot, "vocabulary.json"),
      JSON.stringify({Handoff: "contract.md"}),
      "utf8",
    );

    const {findings} = checkGuideDrift(guideDirectory, repositoryRoot);

    expect(findings).toHaveLength(1);
    expect(findings[0]).toContain("contract.md owns");
  });

  it("surfaces a malformed manifest separately from findings", () => {
    const {guideDirectory, repositoryRoot} = catalogFixture("skill prose");
    writeFileSync(join(repositoryRoot, "budgets.json"), "{", "utf8");

    const {findings, manifestErrors} = checkGuideDrift(
      guideDirectory,
      repositoryRoot,
    );

    expect(findings).toEqual([]);
    expect(manifestErrors).toHaveLength(1);
    expect(manifestErrors[0]).toContain("not valid JSON");
  });
});
