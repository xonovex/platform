import {dirname, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {
  BUDGET_MANIFEST_FILE,
  DRIFT_LINT_MODE_ENV,
  evaluateBudgets,
  readBudgetManifest,
  resolveDriftLintMode,
  seedBudgets,
} from "@xonovex/script-moon-common/drift-budgets";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {findWorkspaceRoot} from "@xonovex/script-moon-common/workspace";
import {
  collectCommandCatalogFiles,
  collectSkillCatalogFiles,
} from "@xonovex/script-moon-skill-catalog-common/drift-files";
import {evaluateDuplication} from "./drift-duplication.js";

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

export const main = (
  argv: readonly string[],
  fs: FileSystem = nodeFileSystem,
): number => {
  if (argv.includes("-h") || argv.includes("--help")) {
    console.log(HELP);
    return 0;
  }
  const known = new Set(["--seed", "--repo-root"]);
  const unknown = argv.find(
    (argument) => argument.startsWith("-") && !known.has(argument),
  );
  if (unknown !== undefined) {
    process.stderr.write(`Error: unrecognized argument: ${unknown}\n`);
    return 2;
  }

  const rootIndex = argv.indexOf("--repo-root");
  const repositoryRoot =
    rootIndex === -1
      ? findWorkspaceRoot(
          dirname(fileURLToPath(import.meta.url)),
          undefined,
          fs,
        )
      : resolve(argv[rootIndex + 1] ?? ".");

  const skillFiles = collectSkillCatalogFiles(repositoryRoot, fs);
  const files = [
    ...skillFiles,
    ...collectCommandCatalogFiles(repositoryRoot, fs),
  ];
  if (files.length === 0) {
    process.stderr.write(`Error: no catalog files under ${repositoryRoot}\n`);
    return 2;
  }

  const budgetPath = join(repositoryRoot, BUDGET_MANIFEST_FILE);
  if (argv.includes("--seed")) {
    const manifest = seedBudgets(files);
    fs.writeFile(budgetPath, `${JSON.stringify(manifest, null, 2)}\n`);
    console.log(
      `Seeded ${String(Object.keys(manifest).length)} budgets into ${BUDGET_MANIFEST_FILE}`,
    );
    return 0;
  }

  const budgets = readBudgetManifest(budgetPath, fs);
  const manifestErrors = [budgets.error].filter(
    (error): error is string => error !== undefined,
  );
  const findings = [
    ...evaluateBudgets(files, budgets.manifest).map(({message}) => message),
    ...evaluateDuplication(skillFiles).map(({message}) => message),
  ];

  const mode = resolveDriftLintMode(process.env[DRIFT_LINT_MODE_ENV]);
  for (const error of manifestErrors) console.log(`[FAIL] ${error}`);
  for (const finding of findings) {
    console.log(`[${mode === "enforce" ? "FAIL" : "DRIFT"}] ${finding}`);
  }
  console.log();
  console.log(
    `Drift (${mode} mode): ${String(files.length)} files, ` +
      `${String(findings.length)} finding(s)`,
  );
  if (manifestErrors.length > 0) return 1;
  return mode === "enforce" && findings.length > 0 ? 1 : 0;
};
