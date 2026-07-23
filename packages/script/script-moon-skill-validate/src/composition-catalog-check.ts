import {readdirSync, readFileSync} from "node:fs";
import {join, relative} from "node:path";
import {
  compositionCatalogIssues,
  compositionCatalogSnapshotErrors,
  parseCompositionCatalog,
  type InstalledSkill,
} from "@xonovex/core/skill-composition";
import {isDirectory, isFile} from "@xonovex/script-moon-common/fs";
import {valid} from "semver";
import {type LinkReport} from "./reference-file-links.js";

const installedSkillInventory = (
  repoRoot: string,
  report: LinkReport,
): {readonly skills: readonly InstalledSkill[]; readonly valid: boolean} => {
  const root = join(repoRoot, "packages", "skill");
  if (!isDirectory(root)) return {skills: [], valid: false};
  const installed: InstalledSkill[] = [];
  let inventoryValid = true;
  for (const packageName of readdirSync(root).toSorted()) {
    const packagePath = join(root, packageName);
    if (!packageName.startsWith("skill-") || !isDirectory(packagePath)) {
      continue;
    }
    const packageJsonPath = join(packagePath, "package.json");
    if (!isFile(packageJsonPath)) continue;
    let packageJson: unknown;
    try {
      packageJson = JSON.parse(
        readFileSync(packageJsonPath, "utf8"),
      ) as unknown;
    } catch (error) {
      inventoryValid = false;
      report.addFail(
        `composition catalog: invalid JSON in ${relative(repoRoot, packageJsonPath)}: ${error instanceof Error ? error.message : String(error)}`,
      );
      continue;
    }
    if (
      typeof packageJson !== "object" ||
      packageJson === null ||
      !("version" in packageJson) ||
      typeof packageJson.version !== "string" ||
      valid(packageJson.version) === null
    ) {
      inventoryValid = false;
      report.addFail(
        `composition catalog: ${relative(repoRoot, packageJsonPath)} needs a semantic version`,
      );
      continue;
    }
    for (const guide of readdirSync(packagePath).toSorted()) {
      const guidePath = join(packagePath, guide);
      if (!isFile(join(guidePath, "SKILL.md"))) continue;
      const sourcesPath = join(guidePath, "SOURCES.md");
      if (!isFile(sourcesPath)) {
        inventoryValid = false;
        report.addFail(
          `composition catalog: ${relative(repoRoot, sourcesPath)} is missing selection provenance`,
        );
      }
      installed.push({
        guide,
        implementationVersion: packageJson.version,
        packagePath: relative(repoRoot, packagePath),
        plugin: `xonovex-${packageName}`,
        sourcesPath: relative(repoRoot, sourcesPath),
      });
    }
  }
  return {skills: installed, valid: inventoryValid};
};

export const checkCompositionCatalog = (
  repoRoot: string,
  report: LinkReport,
): void => {
  const path = join(repoRoot, "packages", "skill", "composition-catalog.json");
  if (!isFile(path)) {
    report.addFail(
      "composition catalog: packages/skill/composition-catalog.json is missing",
    );
    return;
  }
  const sourceText = readFileSync(path, "utf8");
  let input: unknown;
  try {
    input = JSON.parse(sourceText) as unknown;
  } catch (error) {
    report.addFail(
      `composition catalog: invalid JSON: ${error instanceof Error ? error.message : String(error)}`,
    );
    return;
  }
  const parsed = parseCompositionCatalog(input, sourceText);
  if (!parsed.success) {
    for (const error of parsed.errors) {
      report.addFail(`composition catalog: ${error}`);
    }
    return;
  }
  const snapshotPath = join(
    repoRoot,
    "packages",
    "skill",
    "skill-workflow",
    "workflow-guide",
    "assets",
    "composition-catalog.json",
  );
  const snapshotText = isFile(snapshotPath)
    ? readFileSync(snapshotPath, "utf8")
    : undefined;
  const snapshotErrors = compositionCatalogSnapshotErrors(
    sourceText,
    snapshotText,
  );
  for (const error of snapshotErrors) {
    report.addFail(`composition catalog: ${error}`);
  }
  const installed = installedSkillInventory(repoRoot, report);
  const issues = compositionCatalogIssues(parsed.data, installed.skills);
  for (const error of issues.errors) {
    report.addFail(`composition catalog: ${error}`);
  }
  for (const failure of issues.preferredFailures) {
    report.addWarn?.(
      `composition catalog: preferred requirement degraded: ${failure.message}`,
    );
  }
  if (
    issues.errors.length === 0 &&
    snapshotErrors.length === 0 &&
    installed.valid &&
    installed.skills.length > 0
  ) {
    report.addPass(
      `composition catalog: ${String(installed.skills.length)} installed skill(s) have one validated classification`,
    );
  }
};
