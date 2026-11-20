import {join} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {
  BUDGET_MANIFEST_FILE,
  evaluateBudgets,
  readBudgetManifest,
} from "@xonovex/script-moon-common/drift-budgets";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {collectSkillFiles} from "./drift-files.js";

export interface DriftLintReport {
  readonly findings: readonly string[];
  readonly manifestErrors: readonly string[];
}

// checkSkillDrift applies the per-file rules to one skill. Cross-file
// duplication spans the catalog and belongs to moon-skill-validate-drift instead.
export const checkSkillDrift = (
  skillDirectory: string,
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): DriftLintReport => {
  const budgets = readBudgetManifest(
    join(repositoryRoot, BUDGET_MANIFEST_FILE),
    fs,
  );
  const files = collectSkillFiles(skillDirectory, repositoryRoot, fs);
  return {
    findings: evaluateBudgets(files, budgets.manifest).map(
      ({message}) => message,
    ),
    manifestErrors: [budgets.error].filter(
      (error): error is string => error !== undefined,
    ),
  };
};
