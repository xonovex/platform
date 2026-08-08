import {join} from "node:path";
import {
  BUDGET_MANIFEST_FILE,
  evaluateBudgets,
  readBudgetManifest,
} from "@xonovex/script-moon-common/drift-budgets";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {collectGuideFiles} from "./drift-files.js";
import {
  evaluateVocabulary,
  readVocabularyManifest,
  VOCABULARY_MANIFEST_FILE,
} from "./drift-vocabulary.js";

export interface DriftLintReport {
  readonly findings: readonly string[];
  readonly manifestErrors: readonly string[];
}

// checkGuideDrift applies the per-file rules to one guide. Cross-file
// duplication spans the catalog and belongs to moon-skill-validate-drift instead.
export const checkGuideDrift = (
  guideDirectory: string,
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): DriftLintReport => {
  const budgets = readBudgetManifest(
    join(repositoryRoot, BUDGET_MANIFEST_FILE),
    fs,
  );
  const vocabulary = readVocabularyManifest(
    join(repositoryRoot, VOCABULARY_MANIFEST_FILE),
    fs,
  );
  const files = collectGuideFiles(guideDirectory, repositoryRoot, fs);
  return {
    findings: [
      ...evaluateBudgets(files, budgets.manifest).map(({message}) => message),
      ...evaluateVocabulary(files, vocabulary.manifest).map(
        ({message}) => message,
      ),
    ],
    manifestErrors: [budgets.error, vocabulary.error].filter(
      (error): error is string => error !== undefined,
    ),
  };
};
