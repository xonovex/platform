import {readdirSync, readFileSync} from "node:fs";
import {join, relative, sep} from "node:path";
import type {BudgetedFile} from "@xonovex/script-moon-common/drift-budgets";
import {isDirectory, isFile} from "@xonovex/script-moon-common/fs";

const readDirectory = (path: string): readonly string[] =>
  isDirectory(path) ? readdirSync(path).toSorted() : [];

// manifestPath keys every manifest entry by a repository-relative POSIX path so
// the manifests stay readable and platform independent.
export const manifestPath = (repositoryRoot: string, path: string): string =>
  relative(repositoryRoot, path).split(sep).join("/");

const readBudgeted = (
  repositoryRoot: string,
  path: string,
  kind: BudgetedFile["kind"],
): readonly BudgetedFile[] =>
  isFile(path)
    ? [
        {
          kind,
          path: manifestPath(repositoryRoot, path),
          text: readFileSync(path, "utf8"),
        },
      ]
    : [];

// collectGuideFiles returns the SKILL.md and reference files of one guide.
export const collectGuideFiles = (
  guideDirectory: string,
  repositoryRoot: string,
): readonly BudgetedFile[] => [
  ...readBudgeted(repositoryRoot, join(guideDirectory, "SKILL.md"), "skill"),
  ...readDirectory(join(guideDirectory, "references"))
    .filter((entry) => entry.endsWith(".md"))
    .flatMap((entry) =>
      readBudgeted(
        repositoryRoot,
        join(guideDirectory, "references", entry),
        "reference",
      ),
    ),
];

// collectSkillCatalogFiles walks every guide under packages/skill.
export const collectSkillCatalogFiles = (
  repositoryRoot: string,
): readonly BudgetedFile[] => {
  const skillRoot = join(repositoryRoot, "packages", "skill");
  return readDirectory(skillRoot).flatMap((packageName) =>
    readDirectory(join(skillRoot, packageName))
      .filter((entry) => entry.endsWith("-guide"))
      .flatMap((guide) =>
        collectGuideFiles(join(skillRoot, packageName, guide), repositoryRoot),
      ),
  );
};

// collectCommandCatalogFiles walks every command document under packages/command.
export const collectCommandCatalogFiles = (
  repositoryRoot: string,
): readonly BudgetedFile[] => {
  const commandRoot = join(repositoryRoot, "packages", "command");
  return readDirectory(commandRoot).flatMap((packageName) =>
    collectCommandPackageFiles(join(commandRoot, packageName), repositoryRoot),
  );
};

// collectCommandPackageFiles returns the command documents of one package.
export const collectCommandPackageFiles = (
  packageDirectory: string,
  repositoryRoot: string,
): readonly BudgetedFile[] =>
  readDirectory(join(packageDirectory, "commands"))
    .filter((entry) => entry.endsWith(".md"))
    .flatMap((entry) =>
      readBudgeted(
        repositoryRoot,
        join(packageDirectory, "commands", entry),
        "command",
      ),
    );
