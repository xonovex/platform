import {readFileSync} from "node:fs";
import {z} from "zod";
import {isFile} from "./fs.js";

// Drift findings are advisory until the mode is switched to enforce.
export type DriftLintMode = "warn" | "enforce";

export const DRIFT_LINT_MODE_ENV = "XONOVEX_LINT_MODE";

export const resolveDriftLintMode = (
  value: string | undefined,
): DriftLintMode => (value?.trim() === "enforce" ? "enforce" : "warn");

// Absolute ceilings for files absent from the budget manifest, from the p90 of
// the catalog at the time the ratchet was introduced.
export const DRIFT_WORD_CAPS = {
  command: 250,
  reference: 650,
  skill: 900,
} as const;

export type DriftFileKind = keyof typeof DRIFT_WORD_CAPS;

// The repository-relative name of the budget manifest, owned here alongside the
// schema that parses it so every reader resolves the same file.
export const BUDGET_MANIFEST_FILE = "budgets.json";

export const BudgetManifestSchema = z.record(z.string(), z.int().positive());

export type BudgetManifest = z.infer<typeof BudgetManifestSchema>;

export interface DriftFinding {
  readonly path: string;
  readonly message: string;
}

export interface BudgetedFile {
  readonly kind: DriftFileKind;
  // path is repository-relative and is the manifest key.
  readonly path: string;
  readonly text: string;
}

const FENCED_BLOCK_RE = /^```[\s\S]+?^```[^\S\n]*$/gm;

// countProseWords measures rendered prose: fenced code is excluded so that
// examples do not consume a file's word budget.
export const countProseWords = (text: string): number => {
  const prose = text.replaceAll(FENCED_BLOCK_RE, "");
  return prose.split(/\s+/u).filter((token) => token.length > 0).length;
};

export const readBudgetManifest = (
  path: string,
): {readonly manifest: BudgetManifest; readonly error?: string} => {
  if (!isFile(path)) return {manifest: {}};
  let parsed: unknown;
  try {
    parsed = JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    return {
      error: `budget manifest is not valid JSON: ${detail}`,
      manifest: {},
    };
  }
  const result = BudgetManifestSchema.safeParse(parsed);
  if (!result.success) {
    return {
      error: `budget manifest entries must map a path to a positive word budget`,
      manifest: {},
    };
  }
  return {manifest: result.data};
};

// evaluateBudgets ratchets: a file listed in the manifest may not exceed its
// recorded budget, and an unlisted file may not exceed its kind's absolute cap.
export const evaluateBudgets = (
  files: readonly BudgetedFile[],
  manifest: BudgetManifest,
): readonly DriftFinding[] =>
  files.flatMap((file) => {
    const words = countProseWords(file.text);
    const budget = manifest[file.path];
    if (budget !== undefined) {
      return words > budget
        ? [
            {
              message:
                `budget: ${file.path} is ${String(words)} words, over its ` +
                `${String(budget)}-word budget; bump the budget in the same change`,
              path: file.path,
            },
          ]
        : [];
    }
    const cap = DRIFT_WORD_CAPS[file.kind];
    return words > cap
      ? [
          {
            message:
              `budget: ${file.path} is ${String(words)} words, over the ` +
              `${String(cap)}-word cap for a new ${file.kind} file`,
            path: file.path,
          },
        ]
      : [];
  });

// byCodePoint orders manifest keys the way the stored manifest is written. A
// locale-aware comparison folds case and punctuation, which reorders every
// nested path on a reseed and varies with the machine's locale.
const byCodePoint = (left: string, right: string): number => {
  if (left === right) return 0;
  return left < right ? -1 : 1;
};

// seedBudgets records current sizes so the ratchet starts from the corrected
// catalog rather than from the caps.
export const seedBudgets = (files: readonly BudgetedFile[]): BudgetManifest =>
  Object.fromEntries(
    files
      .map((file) => [file.path, countProseWords(file.text)] as const)
      .filter(([, words]) => words > 0)
      .toSorted(([left], [right]) => byCodePoint(left, right)),
  );
