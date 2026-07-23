import {readFileSync, writeFileSync} from "node:fs";
import {join} from "node:path";

export interface CompositionCatalogSyncResult {
  readonly changed: boolean;
  readonly sourcePath: string;
  readonly snapshotPath: string;
}

const catalogPaths = (
  repoRoot: string,
): Pick<CompositionCatalogSyncResult, "sourcePath" | "snapshotPath"> => ({
  sourcePath: join(repoRoot, "packages", "skill", "composition-catalog.json"),
  snapshotPath: join(
    repoRoot,
    "packages",
    "skill",
    "skill-workflow",
    "workflow-guide",
    "assets",
    "composition-catalog.json",
  ),
});

export const syncCompositionCatalog = (
  repoRoot: string,
): CompositionCatalogSyncResult => {
  const paths = catalogPaths(repoRoot);
  const source = readFileSync(paths.sourcePath, "utf8");
  const snapshot = readFileSync(paths.snapshotPath, "utf8");
  if (source === snapshot) return {...paths, changed: false};
  writeFileSync(paths.snapshotPath, source);
  return {...paths, changed: true};
};
