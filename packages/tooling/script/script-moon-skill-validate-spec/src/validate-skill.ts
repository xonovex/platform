import {resolve} from "node:path";
import type {DriftLintMode} from "@xonovex/script-moon-common/drift-budgets";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import {renderReport} from "./report.js";
import {parseSkillDocument} from "./skill-document.js";
import {scanSkill} from "./skill-scan.js";
import {locateSkill} from "./skill-target.js";
import {validateSkill} from "./validate.js";

const HELP_TEXT = [
  "usage: moon-skill-validate-spec [-h] [--strict] [target]",
  "",
  "Validate a SKILL.md against the Agent Skills spec and authoring best practices.",
  "",
  "positional arguments:",
  "  target      skill directory or path to a SKILL.md (defaults to the current directory)",
  "",
  "options:",
  "  -h, --help  show this help message and exit",
  "  --strict     treat authoring warnings as validation failures",
].join("\n");

export interface ValidateSpecDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
  readonly driftMode: DriftLintMode;
}

/** Validates the skill that argv names, or the one at cwd, and returns the exit code. */
export const main = (
  argv: readonly string[],
  cwd: string,
  dependencies: ValidateSpecDependencies,
): number => {
  const {driftMode, fs, output} = dependencies;
  if (argv.includes("-h") || argv.includes("--help")) {
    output.stdout(HELP_TEXT);
    return 0;
  }
  const unknownOption = argv.find(
    (argument) => argument.startsWith("-") && argument !== "--strict",
  );
  if (unknownOption !== undefined) {
    output.stderr(`Error: unrecognized argument: ${unknownOption}`);
    return 2;
  }
  const strict = argv.includes("--strict");
  const positional = argv.find((argument) => !argument.startsWith("-"));

  const location = locateSkill(resolve(cwd, positional ?? "."), fs);
  if (location.kind === "missing") {
    output.stderr(`Error: ${location.message}`);
    return 2;
  }
  const {skillDir, skillPath} = location;
  const document = parseSkillDocument(fs.readText(skillPath));
  if (document.kind === "invalid") {
    output.stderr(`Error: ${document.message}`);
    return 2;
  }

  const findings = validateSkill({
    document,
    driftMode,
    isFile: (path) => fs.isFile(path),
    skillDir,
    snapshot: scanSkill(skillDir, fs),
  });
  const report = renderReport(findings, skillPath, skillDir, strict);
  for (const line of report.lines) output.stdout(line);
  return report.exitCode;
};
