import type {VersionGit} from "../../src/git.js";

/**
 * A repository with no history: no file at any commit, no path history, and no
 * commits. A case overrides the reads it depends on.
 */
export const emptyHistory = (
  overrides: Partial<VersionGit> = {},
): VersionGit => ({
  readFile: () => void 0,
  pathHistory: () => [],
  hasParent: () => false,
  commitsSince: () => [],
  ...overrides,
});

/** Reads each manifest at HEAD from a path-to-manifest record. */
export const manifestsAtHead =
  (manifests: Readonly<Record<string, unknown>>): VersionGit["readFile"] =>
  (commit, path) => {
    const manifest = manifests[path];
    return commit === "HEAD" && manifest !== undefined
      ? JSON.stringify(manifest)
      : undefined;
  };
