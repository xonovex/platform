import {join} from "node:path";
import {
  evaluateBudgets,
  readBudgetManifest,
} from "@xonovex/script-moon-common/drift-budgets";
import {collectGuideFiles} from "./drift-files.js";
import {
  evaluateVocabulary,
  readVocabularyManifest,
} from "./drift-vocabulary.js";

export const BUDGET_MANIFEST_FILE = "budgets.json";
export const VOCABULARY_MANIFEST_FILE = "vocabulary.json";

export interface DriftLintReport {
  readonly findings: readonly string[];
  readonly manifestErrors: readonly string[];
}

// checkGuideDrift applies the per-file rules to one guide. Cross-file
// duplication spans the catalog and belongs to moon-skill-validate-drift instead.
export const checkGuideDrift = (
  guideDirectory: string,
  repositoryRoot: string,
): DriftLintReport => {
  const budgets = readBudgetManifest(
    join(repositoryRoot, BUDGET_MANIFEST_FILE),
  );
  const vocabulary = readVocabularyManifest(
    join(repositoryRoot, VOCABULARY_MANIFEST_FILE),
  );
  const files = collectGuideFiles(guideDirectory, repositoryRoot);
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
