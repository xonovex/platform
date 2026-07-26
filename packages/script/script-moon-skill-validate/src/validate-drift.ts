import {writeFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {findWorkspaceRoot} from "@xonovex/script-moon-common";
import {
  DRIFT_LINT_MODE_ENV,
  evaluateBudgets,
  readBudgetManifest,
  resolveDriftLintMode,
  seedBudgets,
} from "@xonovex/script-moon-common/drift-budgets";
import {evaluateDuplication} from "./drift-duplication.js";
import {
  collectCommandCatalogFiles,
  collectSkillCatalogFiles,
} from "./drift-files.js";
import {BUDGET_MANIFEST_FILE, VOCABULARY_MANIFEST_FILE} from "./drift-lints.js";
import {
  evaluateVocabulary,
  readVocabularyManifest,
} from "./drift-vocabulary.js";

const HELP = [
  "usage: moon-skill-drift [-h] [--seed] [--repo-root PATH]",
  "",
  "Report catalog drift: file budgets, vocabulary ownership, and invariants",
  "restated across files.",
  "",
  "options:",
  "  -h, --help         show this help message and exit",
  "  --seed             write current file sizes to budgets.json and exit",
  "  --repo-root PATH   repository root (defaults to the detected workspace)",
  "",
  `Set ${DRIFT_LINT_MODE_ENV}=enforce to exit non-zero on findings.`,
].join("\n");

export const main = (argv: readonly string[]): number => {
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
      ? findWorkspaceRoot(dirname(fileURLToPath(import.meta.url)))
      : resolve(argv[rootIndex + 1] ?? ".");

  const skillFiles = collectSkillCatalogFiles(repositoryRoot);
  const files = [...skillFiles, ...collectCommandCatalogFiles(repositoryRoot)];
  if (files.length === 0) {
    process.stderr.write(`Error: no catalog files under ${repositoryRoot}\n`);
    return 2;
  }

  const budgetPath = join(repositoryRoot, BUDGET_MANIFEST_FILE);
  if (argv.includes("--seed")) {
    const manifest = seedBudgets(files);
    writeFileSync(budgetPath, `${JSON.stringify(manifest, null, 2)}\n`, "utf8");
    console.log(
      `Seeded ${String(Object.keys(manifest).length)} budgets into ${BUDGET_MANIFEST_FILE}`,
    );
    return 0;
  }

  const budgets = readBudgetManifest(budgetPath);
  const vocabulary = readVocabularyManifest(
    join(repositoryRoot, VOCABULARY_MANIFEST_FILE),
  );
  const manifestErrors = [budgets.error, vocabulary.error].filter(
    (error): error is string => error !== undefined,
  );
  const findings = [
    ...evaluateBudgets(files, budgets.manifest).map(({message}) => message),
    ...evaluateVocabulary(files, vocabulary.manifest).map(
      ({message}) => message,
    ),
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
