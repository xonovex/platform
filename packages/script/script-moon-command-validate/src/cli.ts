import {resolve} from "node:path";
import {parseArgs} from "node:util";
import {validateCommandPackage} from "./command-validation.js";
import {type ValidationReport} from "./validation.js";
import {validateWorkflowContract} from "./workflow-contract.js";
import {validateWorkflowSchemaAssets} from "./workflow-schema-assets.js";

const HELP = `usage: moon-command-validate [package-dir] [options]

Validate command Markdown and optional package-specific semantic contracts.

options:
  --assets <path>     validate workflow JSON Schemas and examples
  --contract <path>   validate a declarative workflow command contract
  --repo-root <path>  repository root (defaults to three levels above package)
  -h, --help          show this help`;

const render = (report: ValidationReport): number => {
  for (const problem of report.issues) {
    console.error(
      `[FAIL] ${problem.path}: ${problem.code}: ${problem.message}`,
    );
  }
  if (report.issues.length > 0) {
    console.error(
      `Result: FAIL (${String(report.issues.length)} issue(s), ${String(report.commands)} command(s))`,
    );
    return 1;
  }
  console.log(`Result: PASS (${String(report.commands)} command(s))`);
  return 0;
};

export const main = (argv: readonly string[]): number => {
  if (argv.includes("-h") || argv.includes("--help")) {
    console.log(HELP);
    return 0;
  }

  let parsed: ReturnType<typeof parseArgs>;
  try {
    parsed = parseArgs({
      args: [...argv],
      allowPositionals: true,
      strict: true,
      options: {
        assets: {type: "string"},
        contract: {type: "string"},
        "repo-root": {type: "string"},
      },
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`command validation input error: ${message}\n`);
    return 2;
  }
  if (parsed.positionals.length > 1) {
    process.stderr.write(
      "command validation input error: expected at most one package directory\n",
    );
    return 2;
  }

  const packageDir = resolve(parsed.positionals[0] ?? ".");
  const repositoryRoot =
    typeof parsed.values["repo-root"] === "string"
      ? resolve(parsed.values["repo-root"])
      : resolve(packageDir, "../../..");
  const contract =
    typeof parsed.values.contract === "string"
      ? resolve(parsed.values.contract)
      : undefined;
  const assets =
    typeof parsed.values.assets === "string"
      ? resolve(parsed.values.assets)
      : undefined;
  try {
    const report =
      contract === undefined
        ? validateCommandPackage(packageDir, repositoryRoot).report
        : validateWorkflowContract(contract, packageDir, repositoryRoot);
    if (assets === undefined) return render(report);
    const assetReport = validateWorkflowSchemaAssets(assets);
    return render({
      commands: report.commands,
      issues: [...report.issues, ...assetReport.issues],
    });
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(
      `command validation failed unexpectedly: ${message}\n`,
    );
    return 2;
  }
};
