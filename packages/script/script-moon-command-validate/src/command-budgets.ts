import {readdirSync, readFileSync} from "node:fs";
import {join, relative, sep} from "node:path";
import {
  DRIFT_LINT_MODE_ENV,
  evaluateBudgets,
  readBudgetManifest,
  resolveDriftLintMode,
  type BudgetedFile,
} from "@xonovex/script-moon-common/drift-budgets";
import {isDirectory, isFile} from "@xonovex/script-moon-common/fs";
import {issue, type ValidationIssue} from "./validation.js";

export const BUDGET_MANIFEST_FILE = "budgets.json";

const MARKDOWN_RE = /\.md$/;

// commandFiles reads every command document as a budgeted file keyed by its
// repository-relative path, matching the manifest written by moon-skill-validate-drift.
export const commandFiles = (
  packageDirectory: string,
  repositoryRoot: string,
): readonly BudgetedFile[] => {
  const commandDirectory = join(packageDirectory, "commands");
  if (!isDirectory(commandDirectory)) return [];
  return readdirSync(commandDirectory)
    .toSorted()
    .filter((entry) => MARKDOWN_RE.test(entry))
    .map((entry) => join(commandDirectory, entry))
    .filter((path) => isFile(path))
    .map((path) => ({
      kind: "command" as const,
      path: relative(repositoryRoot, path).split(sep).join("/"),
      text: readFileSync(path, "utf8"),
    }));
};

// checkCommandBudgets reports ratchet findings as warnings until the drift mode
// is switched to enforce.
export const checkCommandBudgets = (
  packageDirectory: string,
  repositoryRoot: string,
  mode = resolveDriftLintMode(process.env[DRIFT_LINT_MODE_ENV]),
): readonly ValidationIssue[] => {
  const {error, manifest} = readBudgetManifest(
    join(repositoryRoot, BUDGET_MANIFEST_FILE),
  );
  const severity = mode === "enforce" ? "error" : "warning";
  const manifestIssues =
    error === undefined
      ? []
      : [issue("budget-manifest", BUDGET_MANIFEST_FILE, error)];
  return [
    ...manifestIssues,
    ...evaluateBudgets(
      commandFiles(packageDirectory, repositoryRoot),
      manifest,
    ).map(({message, path}) => issue("budget", path, message, severity)),
  ];
};
