// npm creates a node_modules/.bin link only when the bin target already exists
// as the package is linked, and building the target afterwards does not add the
// link, so npx falls through to the registry and 404s. Whoever links the package
// decides which tree the target has to be present in: a private package is only
// ever linked from the working tree by this workspace's install, so its target
// has to be committed, while a published package is linked from its unpacked
// tarball, so its target has to be inside the files allowlist.

export interface BinEntry {
  readonly name: string;
  readonly target: string;
}

export interface PackageBins {
  // manifestPath is workspace relative, for the failure message.
  readonly manifestPath: string;
  readonly packageDirectory: string;
  readonly isPrivate: boolean;
  readonly files: readonly string[] | undefined;
  readonly bins: readonly BinEntry[];
}

// A string bin takes the package name, which npm strips to the last segment for
// a scoped package.
export const binEntries = (
  bin: string | Readonly<Record<string, string>> | undefined,
  packageName: string,
): readonly BinEntry[] => {
  if (bin === undefined) return [];
  if (typeof bin === "string") {
    return [{name: packageName.split("/").at(-1) ?? packageName, target: bin}];
  }
  return Object.entries(bin).map(([name, target]) => ({name, target}));
};

export interface BinTargetInput {
  readonly packages: readonly PackageBins[];
  // trackedPaths are workspace relative, as git reports committed files.
  readonly trackedPaths: ReadonlySet<string>;
}

const normalize = (path: string): string =>
  path.replace(/^\.\//u, "").replace(/\/$/u, "");

// A files entry covers the target when it names it or is a directory above it.
// npm accepts globs there too; an unmatched glob reports a failure the author
// resolves by committing the target or naming it exactly.
const isPacked = (
  target: string,
  files: readonly string[] | undefined,
): boolean =>
  (files ?? []).some((entry) => {
    const packed = normalize(entry);
    return target === packed || target.startsWith(`${packed}/`);
  });

export const binTargetFailures = (input: BinTargetInput): readonly string[] => {
  const failures: string[] = [];
  for (const packageBins of input.packages) {
    for (const bin of packageBins.bins) {
      const target = normalize(bin.target);
      if (input.trackedPaths.has(`${packageBins.packageDirectory}/${target}`)) {
        continue;
      }
      if (!packageBins.isPrivate && isPacked(target, packageBins.files)) {
        continue;
      }
      failures.push(
        packageBins.isPrivate
          ? `${packageBins.manifestPath} bin ${bin.name} points at ${target}, which is not committed, so a cold install links no ${bin.name}`
          : `${packageBins.manifestPath} bin ${bin.name} points at ${target}, which is neither committed nor inside the files allowlist`,
      );
    }
  }
  return failures;
};
