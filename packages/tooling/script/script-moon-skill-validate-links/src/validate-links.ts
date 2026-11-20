import {resolve} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import {catalogDocuments, scanCatalog} from "./catalog-scan.js";
import {checkCrossPackageLinks} from "./cross-package-links.js";
import {renderLinkReport} from "./report.js";
import {checkSkillDependencies} from "./skill-dependencies.js";
import {checkNamedSkillHandoffs} from "./skill-handoffs.js";

const HELP_TEXT = [
  "usage: moon-skill-validate-links [-h] [repo-root]",
  "",
  "Validate links, skill packaging, and paired acyclic hard dependencies.",
  "",
  "positional arguments:",
  "  repo-root   repository root to scan (defaults to the current directory)",
  "",
  "options:",
  "  -h, --help  show this help message and exit",
].join("\n");

export interface ValidateLinksDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
}

/** Validates the links and the plugin packaging of a repository and returns the exit code. */
export const main = (
  argv: readonly string[],
  cwd: string,
  dependencies: ValidateLinksDependencies,
): number => {
  const {fs, output} = dependencies;
  if (argv.includes("-h") || argv.includes("--help")) {
    output.stdout(HELP_TEXT);
    return 0;
  }

  // Resolve the root against the working directory: source paths and link
  // targets must both be absolute, or every relative link reads as one that
  // crosses a package boundary.
  const repoRoot = resolve(
    cwd,
    argv.find((arg) => !arg.startsWith("-")) ?? ".",
  );
  const snapshot = scanCatalog(repoRoot, fs);
  const documents = catalogDocuments(snapshot);
  const report = renderLinkReport([
    ...checkCrossPackageLinks(
      documents,
      repoRoot,
      (path) => fs.isFile(path) || fs.isDirectory(path),
    ),
    ...checkNamedSkillHandoffs(documents, repoRoot),
    ...checkSkillDependencies(snapshot.plugins, repoRoot),
  ]);
  for (const line of report.lines) output.stdout(line);
  return report.exitCode;
};
