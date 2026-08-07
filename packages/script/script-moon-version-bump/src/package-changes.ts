import {existsSync, readFileSync} from "node:fs";
import {dirname, join, relative} from "node:path";
import {logInfo} from "@xonovex/script-moon-common/logging";
import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {
  determineBumpLevel,
  generateChangelogEntry,
  renderUpdatedChangelog,
  type DepUpdate,
} from "./changelog.js";
import type {FileChange} from "./file-transaction.js";
import {findChangelogRange, getCommitsSince} from "./git-log.js";
import type {GitRunner} from "./git.js";

interface ChangelogRequest {
  readonly rootDir: string;
  readonly packagePath: string;
  readonly packageName: string;
  readonly oldVersion: string;
  readonly newVersion: string;
  readonly dryRun: boolean;
  readonly depUpdates: readonly DepUpdate[];
  readonly changelogFilename: string | undefined;
  readonly gitBase: string | undefined;
  readonly includedTypes: ReadonlySet<string> | undefined;
  readonly git: GitRunner;
}

const serializePackage = (pkg: PackageJson): string =>
  JSON.stringify(pkg, null, 2) + "\n";

const planChangelog = (request: ChangelogRequest): FileChange | undefined => {
  const pkgDir = relative(request.rootDir, dirname(request.packagePath));
  const filename = request.changelogFilename ?? "CHANGELOG.md";
  const range =
    request.gitBase === undefined
      ? findChangelogRange(
          request.rootDir,
          pkgDir,
          request.oldVersion,
          request.git,
        )
      : {since: request.gitBase};

  if (range === undefined) {
    logInfo(
      `${request.packageName}: no previous version found, skipping changelog.`,
    );
    return undefined;
  }

  const bumpLevel = determineBumpLevel(request.oldVersion, request.newVersion);
  const commits = getCommitsSince(
    request.rootDir,
    pkgDir,
    range.since,
    request.git,
  );
  const entry = generateChangelogEntry(
    request.newVersion,
    commits,
    bumpLevel,
    request.depUpdates,
    request.includedTypes,
  );

  if (request.dryRun) {
    logInfo(
      `[dry-run] Changelog entry for ${request.packageName}@${request.newVersion}:`,
    );
    console.log(entry);
    return undefined;
  }
  const changelogPath = join(dirname(request.packagePath), filename);
  const existing = existsSync(changelogPath)
    ? readFileSync(changelogPath, "utf8")
    : undefined;
  return {
    path: changelogPath,
    content: renderUpdatedChangelog(existing, request.packageName, entry),
  };
};

export {planChangelog, serializePackage};
export type {ChangelogRequest};
