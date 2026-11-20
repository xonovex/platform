import {join, resolve} from "node:path";
import {
  BUDGET_MANIFEST_FILE,
  DRIFT_LINT_MODE_ENV,
  evaluateBudgets,
  readBudgetManifest,
  seedBudgets,
  type DriftLintMode,
} from "@xonovex/script-moon-common/drift-budgets";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {collectSkillCatalogFiles} from "@xonovex/script-moon-skill-catalog-common/drift-files";
import {evaluateDuplication} from "./drift-duplication.js";
import {renderDriftReport} from "./drift-report.js";

const HELP = [
  "usage: moon-skill-validate-drift [-h] [--seed] [--repo-root PATH]",
  "",
  "Report catalog drift: file budgets and invariants restated across files.",
  "",
  "options:",
  "  -h, --help         show this help message and exit",
  "  --seed             write current file sizes to budgets.json and exit",
  "  --repo-root PATH   repository root (defaults to the detected workspace)",
  "",
  `Set ${DRIFT_LINT_MODE_ENV}=enforce to exit non-zero on findings.`,
].join("\n");

const KNOWN_OPTIONS = new Set(["--seed", "--repo-root"]);

export interface ValidateDriftDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
  readonly driftMode: DriftLintMode;
  /** Where the installed script lives, from which the workspace root is found. */
  readonly scriptDirectory: string;
}

/** Reports or seeds the catalog drift budgets and returns the exit code. */
export const main = (
  argv: readonly string[],
  cwd: string,
  dependencies: ValidateDriftDependencies,
): number => {
  const {driftMode, fs, output, scriptDirectory} = dependencies;
  if (argv.includes("-h") || argv.includes("--help")) {
    output.stdout(HELP);
    return 0;
  }
  const unknown = argv.find(
    (argument) => argument.startsWith("-") && !KNOWN_OPTIONS.has(argument),
  );
  if (unknown !== undefined) {
    output.stderr(`Error: unrecognized argument: ${unknown}`);
    return 2;
  }

  const rootIndex = argv.indexOf("--repo-root");
  const repositoryRoot =
    rootIndex === -1
      ? findWorkspaceRoot(scriptDirectory, undefined, fs)
      : resolve(cwd, argv[rootIndex + 1] ?? ".");

  const files = collectSkillCatalogFiles(repositoryRoot, fs);
  if (files.length === 0) {
    output.stderr(`Error: no catalog files under ${repositoryRoot}`);
    return 2;
  }

  const budgetPath = join(repositoryRoot, BUDGET_MANIFEST_FILE);
  if (argv.includes("--seed")) {
    const manifest = seedBudgets(files);
    fs.writeFile(budgetPath, `${JSON.stringify(manifest, null, 2)}\n`);
    output.stdout(
      `Seeded ${String(Object.keys(manifest).length)} budgets into ${BUDGET_MANIFEST_FILE}`,
    );
    return 0;
  }

  const budgets = readBudgetManifest(budgetPath, fs);
  const report = renderDriftReport({
    fileCount: files.length,
    manifestErrors: [budgets.error].filter(
      (error): error is string => error !== undefined,
    ),
    findings: [
      ...evaluateBudgets(files, budgets.manifest).map(({message}) => message),
      ...evaluateDuplication(files).map(({message}) => message),
    ],
    mode: driftMode,
  });
  for (const line of report.lines) output.stdout(line);
  return report.exitCode;
};
