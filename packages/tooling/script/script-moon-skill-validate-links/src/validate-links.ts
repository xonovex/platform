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

  // The root stays as given. A relative root leaves the source paths relative
  // while link targets resolve to absolute paths, so every relative link counts
  // as crossing a package boundary; the composition-check task passes "." and
  // so checks every relative link of the catalog.
  const repoRoot = argv.find((arg) => !arg.startsWith("-")) ?? cwd;
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
