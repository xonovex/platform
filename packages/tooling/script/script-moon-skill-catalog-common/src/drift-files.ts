import {join, relative, sep} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import type {BudgetedFile} from "@xonovex/script-moon-common/drift-budgets";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {pluginSkillDirectories} from "@xonovex/script-moon-common/package-layout";

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

// collectSkillFiles returns the SKILL.md and reference files of one skill.
export const collectSkillFiles = (
  skillDirectory: string,
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] => [
  ...readBudgeted(
    repositoryRoot,
    join(skillDirectory, "SKILL.md"),
    "skill",
    fs,
  ),
  ...listing(join(skillDirectory, "references"), fs)
    .filter((entry) => entry.endsWith(".md"))
    .flatMap((entry) =>
      readBudgeted(
        repositoryRoot,
        join(skillDirectory, "references", entry),
        "reference",
        fs,
      ),
    ),
];

// collectSkillCatalogFiles walks every skill under packages/runtime/plugin.
export const collectSkillCatalogFiles = (
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] => {
  const skillRoot = join(repositoryRoot, "packages", "runtime", "plugin");
  return listing(skillRoot, fs).flatMap((packageName) =>
    pluginSkillDirectories(join(skillRoot, packageName), fs).flatMap(
      (skillDirectory) => collectSkillFiles(skillDirectory, repositoryRoot, fs),
    ),
  );
};
