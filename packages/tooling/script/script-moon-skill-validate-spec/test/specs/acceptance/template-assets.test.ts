import {
  cpSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {afterEach, describe, expect, it} from "vitest";
import {scanSkill} from "../../../src/skill-scan.js";
import {checkCatalogEvidence} from "../../../src/validate.js";

const templateRoot = resolve(
  import.meta.dirname,
  "../../../../../../runtime/plugin/plugin-agentic/skills/authoring-skills/assets",
);
const temporaryDirectories: string[] = [];

const materializeCatalogEvidence = (templateName: string): string => {
  const temporaryDirectory = mkdtempSync(join(tmpdir(), "skill-template-"));
  temporaryDirectories.push(temporaryDirectory);
  const skillDirectory = join(temporaryDirectory, "using-example");
  cpSync(join(templateRoot, templateName), skillDirectory, {recursive: true});

  const evalsPath = join(skillDirectory, "evals.json");
  const evals = JSON.parse(readFileSync(evalsPath, "utf8")) as Record<
    string,
    unknown
  >;
  writeFileSync(
    evalsPath,
    `${JSON.stringify({...evals, skill_name: "using-example"}, undefined, 2)}\n`,
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

      const findings = checkCatalogEvidence(
        scanSkill(skillDirectory, nodeFileSystem),
        {name: "using-example"},
      );

      expect(findings.filter(({level}) => level === "fail")).toEqual([]);
      expect(
        findings
          .filter(({level}) => level === "pass")
          .map(({message}) => message),
      ).toEqual([
        "catalog: 3 output eval(s) are structurally valid",
        "catalog: trigger evals cover 8 positive and 8 negative routes with train/validation splits",
        "catalog: source provenance and review date are present",
        "credentials: examples avoid plaintext token anti-patterns",
      ]);
    });
  },
);
