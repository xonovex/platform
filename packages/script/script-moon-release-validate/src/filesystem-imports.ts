export interface SourceFile {
  /** Repository-relative path, which is how a failure names the file. */
  readonly path: string;
  readonly text: string;
}

const FS_IMPORT_RE =
  /^\s*(?:import|export)[^;]*?from\s+["']node:fs(?:\/promises)?["']/mu;

/**
 * The modules allowed to reach node:fs directly, each with the reason it cannot
 * go through the FileSystem port. Every other module under src takes the port as
 * a parameter, which is what keeps the unit tier off a real disk. Adding an entry
 * is a reviewed decision rather than a silent import, which is the whole point of
 * naming them here rather than exempting a directory.
 */
const ALLOWED: ReadonlyMap<string, string> = new Map([
  [
    "packages/script/script-moon-common/src/file-system.ts",
    "is the node adapter the port is defined against",
  ],
  [
    "packages/config/eslint-config-base/src/gitignore.ts",
    "is a config package, which cannot depend on script-moon-common without inverting config -> shared -> agent",
  ],
  [
    "packages/agent/agent-cli-go/src/bin.ts",
    "is the launcher that locates the built Go binary before any port exists",
  ],
  [
    "packages/script/script-moon-npm-check/src/index.ts",
    "is a composition root that supplies the real filesystem to its checks",
  ],
  [
    "packages/script/script-moon-npm-publish/src/cli.ts",
    "defines the defaultDependencies its port is driven from",
  ],
  [
    "packages/script/script-moon-release-validate/src/release-validation.ts",
    "is a composition root that walks the repository it validates",
  ],
  [
    "packages/script/script-moon-skill-eval-common/src/trigger-process.ts",
    "stages an isolated workspace for a harness process it spawns",
  ],
  [
    "packages/script/script-moon-skill-eval-outputs/src/output-harness.ts",
    "stages an isolated workspace for a harness process it spawns",
  ],
  [
    "packages/script/script-moon-skill-eval-outputs/src/evaluate.ts",
    "defines the discard in defaultDependencies, which removes a previous run's evidence",
  ],
  [
    "packages/script/script-moon-skill-eval-routing/src/routing-evaluate.ts",
    "defines the discard in defaultDependencies, which removes a previous run's evidence",
  ],
  [
    "packages/script/script-moon-skill-eval-triggers/src/evaluate.ts",
    "defines the discard in defaultDependencies, which removes a previous run's evidence",
  ],
  [
    "packages/script/script-moon-version-bump/src/file-transaction.ts",
    "defines the defaultIo whose exclusive write and rename the transaction rests on",
  ],
  [
    "packages/script/script-moon-version-bump/src/package-changes.ts",
    "reads a changelog relative to the package being bumped",
  ],
]);

/**
 * Reports a module under src that imports node:fs without being named above. The
 * port is only a boundary while every reader goes through it, and a direct import
 * puts a real disk back into whatever tier reaches that module.
 */
export const filesystemImportFailures = (
  files: readonly SourceFile[],
): readonly string[] =>
  files
    .filter((file) => !ALLOWED.has(file.path) && FS_IMPORT_RE.test(file.text))
    .map(
      (file) =>
        `${file.path} imports node:fs directly: take the FileSystem port from @xonovex/script-moon-common/file-system as a defaulted last parameter, or add the file to the allowlist in filesystem-imports.ts with the reason it cannot`,
    )
    .toSorted();

/**
 * The allowlist entries naming a module that is present but no longer imports
 * node:fs, so the allowance has outlived its reason. An entry naming a module
 * absent from the scanned set is left alone, because the same allowlist is
 * applied to fixture repositories that ship none of these files.
 */
export const staleAllowlistFailures = (
  files: readonly SourceFile[],
): readonly string[] => {
  const scanned = new Map(files.map((file) => [file.path, file.text]));
  return [...ALLOWED.keys()]
    .filter((path) => {
      const text = scanned.get(path);
      return text !== undefined && !FS_IMPORT_RE.test(text);
    })
    .map(
      (path) =>
        `${path} is allowed to import node:fs but no longer does: drop it from the allowlist in filesystem-imports.ts`,
    )
    .toSorted();
};
