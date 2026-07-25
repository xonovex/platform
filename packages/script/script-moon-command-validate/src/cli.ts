import {resolve} from "node:path";
import {parseArgs} from "node:util";
import {validateCommandPackage} from "./command-validation.js";
import {type ValidationReport} from "./validation.js";

const HELP = `usage: moon-command-validate [package-dir] [options]

Validate command Markdown, delegation, arguments, and hard dependencies.

options:
  --repo-root <path>  repository root (defaults to three levels above package)
  -h, --help          show this help`;

const render = (report: ValidationReport): number => {
  for (const problem of report.issues) {
    const line = `[${problem.severity === "warning" ? "WARN" : "FAIL"}] ${problem.path}: ${problem.code}: ${problem.message}`;
    if (problem.severity === "warning") console.warn(line);
    else console.error(line);
  }
  const errors = report.issues.filter(({severity}) => severity === "error");
  const warnings = report.issues.filter(({severity}) => severity === "warning");
  if (errors.length > 0) {
    console.error(
      `Result: FAIL (${String(errors.length)} error(s), ${String(warnings.length)} warning(s), ${String(report.commands)} command(s))`,
    );
    return 1;
  }
  console.log(
    `Result: PASS (${String(report.commands)} command(s), ${String(warnings.length)} warning(s))`,
  );
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
  try {
    return render(validateCommandPackage(packageDir, repositoryRoot).report);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(
      `command validation failed unexpectedly: ${message}\n`,
    );
    return 2;
  }
};
