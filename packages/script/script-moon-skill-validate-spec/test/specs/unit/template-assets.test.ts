import {
  cpSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {checkCatalogFiles} from "../../../src/catalog-files.js";

const templateRoot = resolve(
  import.meta.dirname,
  "../../../../../skill/skill-skill/skill-guide/assets",
);
const temporaryDirectories: string[] = [];

const materializeCatalogEvidence = (templateName: string): string => {
  const temporaryDirectory = mkdtempSync(join(tmpdir(), "skill-template-"));
  temporaryDirectories.push(temporaryDirectory);
  const skillDirectory = join(temporaryDirectory, "example-guide");
  cpSync(join(templateRoot, templateName), skillDirectory, {recursive: true});

  const evalsPath = join(skillDirectory, "evals.json");
  const evals = JSON.parse(readFileSync(evalsPath, "utf8")) as Record<
    string,
    unknown
  >;
  writeFileSync(
    evalsPath,
    `${JSON.stringify({...evals, skill_name: "example-guide"}, undefined, 2)}\n`,
  );

  const sourcesPath = join(skillDirectory, "SOURCES.md");
  writeFileSync(
    sourcesPath,
    readFileSync(sourcesPath, "utf8").replaceAll("{YYYY-MM-DD}", "2026-07-22"),
  );
  return skillDirectory;
};

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe.each(["guideline-skill-template", "workflow-skill-template"])(
  "%s",
  (templateName) => {
    it("provides strict catalog evidence after identity fields are filled", () => {
      const skillDirectory = materializeCatalogEvidence(templateName);

      const report = checkCatalogFiles(skillDirectory, {
        name: "example-guide",
      });

      expect(report.errors).toEqual([]);
      expect(report.passes).toEqual([
        "catalog: 3 output eval(s) are structurally valid",
        "catalog: trigger evals cover 8 positive and 8 negative routes with train/validation splits",
        "catalog: source provenance and review date are present",
        "credentials: examples avoid plaintext token anti-patterns",
      ]);
    });
  },
);
