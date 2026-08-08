import {join, relative, sep} from "node:path";
import type {BudgetedFile} from "@xonovex/script-moon-common/drift-budgets";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";

// An absent directory reads as empty, so a catalog missing a whole tier is a
// quiet no-op rather than a throw.
const listing = (path: string, fs: FileSystem): readonly string[] =>
  fs.isDirectory(path) ? fs.readDirectory(path).toSorted() : [];

// manifestPath keys every manifest entry by a repository-relative POSIX path so
// the manifests stay readable and platform independent.
export const manifestPath = (repositoryRoot: string, path: string): string =>
  relative(repositoryRoot, path).split(sep).join("/");

const readBudgeted = (
  repositoryRoot: string,
  path: string,
  kind: BudgetedFile["kind"],
  fs: FileSystem,
): readonly BudgetedFile[] =>
  fs.isFile(path)
    ? [
        {
          kind,
          path: manifestPath(repositoryRoot, path),
          text: fs.readText(path),
        },
      ]
    : [];

// collectGuideFiles returns the SKILL.md and reference files of one guide.
export const collectGuideFiles = (
  guideDirectory: string,
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] => [
  ...readBudgeted(
    repositoryRoot,
    join(guideDirectory, "SKILL.md"),
    "skill",
    fs,
  ),
  ...listing(join(guideDirectory, "references"), fs)
    .filter((entry) => entry.endsWith(".md"))
    .flatMap((entry) =>
      readBudgeted(
        repositoryRoot,
        join(guideDirectory, "references", entry),
        "reference",
        fs,
      ),
    ),
];

// collectSkillCatalogFiles walks every guide under packages/skill.
export const collectSkillCatalogFiles = (
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] => {
  const skillRoot = join(repositoryRoot, "packages", "skill");
  return listing(skillRoot, fs).flatMap((packageName) =>
    listing(join(skillRoot, packageName), fs)
      .filter((entry) => entry.endsWith("-guide"))
      .flatMap((guide) =>
        collectGuideFiles(
          join(skillRoot, packageName, guide),
          repositoryRoot,
          fs,
        ),
      ),
  );
};

// collectCommandCatalogFiles walks every command document under packages/command.
export const collectCommandCatalogFiles = (
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] => {
  const commandRoot = join(repositoryRoot, "packages", "command");
  return listing(commandRoot, fs).flatMap((packageName) =>
    collectCommandPackageFiles(
      join(commandRoot, packageName),
      repositoryRoot,
      fs,
    ),
  );
};

// collectCommandPackageFiles returns the command documents of one package.
export const collectCommandPackageFiles = (
  packageDirectory: string,
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] =>
  listing(join(packageDirectory, "commands"), fs)
    .filter((entry) => entry.endsWith(".md"))
    .flatMap((entry) =>
      readBudgeted(
        repositoryRoot,
        join(packageDirectory, "commands", entry),
        "command",
        fs,
      ),
    );
