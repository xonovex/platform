import {resolve} from "node:path";
import {parseArgs} from "node:util";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import {validatePluginPackage} from "./plugin-validation.js";
import {renderReport} from "./report.js";

const HELP = `usage: moon-plugin-validate [package-dir] [options]

Validate a plugin package: its skill catalog, links, dependencies, and prose punctuation.

options:
  --repo-root <path>  repository root (defaults to three levels above package)
  -h, --help          show this help`;

const errorMessage = (error: unknown): string =>
  error instanceof Error ? error.message : String(error);

interface PluginTarget {
  readonly packageDirectory: string;
  readonly repositoryRoot: string;
}

type ParsedTarget =
  | {readonly kind: "target"; readonly target: PluginTarget}
  | {readonly kind: "invalid"; readonly message: string};

const parseTarget = (argv: readonly string[], cwd: string): ParsedTarget => {
  let parsed: ReturnType<typeof parseArgs>;
  try {
    parsed = parseArgs({
      args: [...argv],
      allowPositionals: true,
      strict: true,
      options: {"repo-root": {type: "string"}},
    });
  } catch (error) {
    return {kind: "invalid", message: errorMessage(error)};
  }
  if (parsed.positionals.length > 1) {
    return {kind: "invalid", message: "expected at most one package directory"};
  }
  const packageDirectory = resolve(cwd, parsed.positionals[0] ?? ".");
  const repoRoot = parsed.values["repo-root"];
  return {
    kind: "target",
    target: {
      packageDirectory,
      repositoryRoot:
        typeof repoRoot === "string"
          ? resolve(cwd, repoRoot)
          : resolve(packageDirectory, "../../../.."),
    },
  };
};

export interface PluginValidateDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
}

/** Validates one plugin package and returns the exit code. */
export const main = (
  argv: readonly string[],
  cwd: string,
  dependencies: PluginValidateDependencies,
): number => {
  const {fs, output} = dependencies;
  if (argv.includes("-h") || argv.includes("--help")) {
    output.stdout(HELP);
    return 0;
  }
  const parsed = parseTarget(argv, cwd);
  if (parsed.kind === "invalid") {
    output.stderr(`plugin validation input error: ${parsed.message}`);
    return 2;
  }
  const {packageDirectory, repositoryRoot} = parsed.target;
  let report: ReturnType<typeof renderReport>;
  try {
    report = renderReport(
      validatePluginPackage(packageDirectory, repositoryRoot, fs),
    );
  } catch (error) {
    output.stderr(
      `plugin validation failed unexpectedly: ${errorMessage(error)}`,
    );
    return 2;
  }
  for (const {stream, text} of report.lines) output[stream](text);
  return report.exitCode;
};
