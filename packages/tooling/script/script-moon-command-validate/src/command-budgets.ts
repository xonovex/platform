import {join, relative, sep} from "node:path";
import {
  BUDGET_MANIFEST_FILE,
  DRIFT_LINT_MODE_ENV,
  evaluateBudgets,
  readBudgetManifest,
  resolveDriftLintMode,
  type BudgetedFile,
} from "@xonovex/script-moon-common/drift-budgets";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {issue, type ValidationIssue} from "./validation.js";

// commandFiles reads every command document as a budgeted file keyed by its
// repository-relative path, matching the manifest written by moon-skill-validate-drift.
export const commandFiles = (
  packageDirectory: string,
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly BudgetedFile[] => {
  const commandDirectory = join(packageDirectory, "commands");
  if (!fs.isDirectory(commandDirectory)) return [];
  return fs
    .readDirectory(commandDirectory)
    .toSorted()
    .filter((entry) => entry.endsWith(".md"))
    .map((entry) => join(commandDirectory, entry))
    .filter((path) => fs.isFile(path))
    .map((path) => ({
      kind: "command" as const,
      path: relative(repositoryRoot, path).split(sep).join("/"),
      text: fs.readText(path),
    }));
};

// checkCommandBudgets reports ratchet findings as warnings until the drift mode
// is switched to enforce.
export const checkCommandBudgets = (
  packageDirectory: string,
  repositoryRoot: string,
  mode = resolveDriftLintMode(process.env[DRIFT_LINT_MODE_ENV]),
  fs: FileSystem = nodeFileSystem,
): readonly ValidationIssue[] => {
  const {error, manifest} = readBudgetManifest(
    join(repositoryRoot, BUDGET_MANIFEST_FILE),
    fs,
  );
  const severity = mode === "enforce" ? "error" : "warning";
  const manifestIssues =
    error === undefined
      ? []
      : [issue("budget-manifest", BUDGET_MANIFEST_FILE, error)];
  return [
    ...manifestIssues,
    ...evaluateBudgets(
      commandFiles(packageDirectory, repositoryRoot, fs),
      manifest,
    ).map(({message, path}) => issue("budget", path, message, severity)),
  ];
};
