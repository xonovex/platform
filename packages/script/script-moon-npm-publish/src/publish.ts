import {spawnSync} from "node:child_process";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";

interface PackageIdentity {
  readonly name: string;
  readonly version: string;
}

interface PlatformMeta {
  readonly os: readonly string[];
  readonly cpu: readonly string[];
  readonly libc?: readonly string[];
}

interface NpmViewResult {
  readonly error?: Error;
  readonly status: number | null;
  readonly stderr: string;
}

interface PublishDependencies {
  readonly readPackageJson: () => string;
  readonly writePackageJson: (contents: string) => void;
  readonly currentDirectory: () => string;
  readonly readPlatformMeta: (directory: string) => PlatformMeta | undefined;
  readonly isPublished: (identity: PackageIdentity) => boolean;
  readonly publish: (dryRun: boolean) => void;
  readonly log: (message: string) => void;
}

type NpmView = (identity: PackageIdentity) => NpmViewResult;

const PACKAGE_NAME_RE = /^(?:@[a-z0-9][a-z0-9._-]*\/)?[a-z0-9][a-z0-9._-]*$/;
const VERSION_RE = /^\d+\.\d+\.\d+(?:-[0-9a-z.-]+)?(?:\+[0-9a-z.-]+)?$/i;

export const parsePackageIdentity = (value: unknown): PackageIdentity => {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error("package.json must contain a JSON object");
  }
  const {name, version} = value as {name?: unknown; version?: unknown};
  if (typeof name !== "string" || !PACKAGE_NAME_RE.test(name)) {
    throw new Error("package.json contains an invalid package name");
  }
  if (typeof version !== "string" || !VERSION_RE.test(version)) {
    throw new Error("package.json contains an invalid semantic version");
  }
  return {name, version};
};

export const publishArgs = (dryRun: boolean): readonly string[] =>
  dryRun
    ? ["publish", "--dry-run", "--access", "public"]
    : ["publish", "--provenance", "--access", "public"];

export const publishedFromResult = (
  {name, version}: PackageIdentity,
  result: NpmViewResult,
): boolean => {
  if (result.error) throw result.error;
  if (result.status === 0) return true;
  if (/\bE404\b|404 Not Found/i.test(result.stderr)) return false;
  const errorDetail = result.stderr.trim() || `exit ${String(result.status)}`;
  throw new Error(`npm view failed for ${name}@${version}: ${errorDetail}`);
};

const runNpmView: NpmView = ({name, version}) => {
  const result = spawnSync(
    resolveExecutable("npm"),
    ["view", `${name}@${version}`, "version"],
    {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    },
  );
  return {
    ...(result.error ? {error: result.error} : {}),
    status: result.status,
    stderr: result.stderr,
  };
};

export const isPublished = (
  identity: PackageIdentity,
  view: NpmView = runNpmView,
): boolean => publishedFromResult(identity, view(identity));

export const publishPackage = (
  dryRun: boolean,
  dependencies: PublishDependencies,
): void => {
  const original = dependencies.readPackageJson();
  const identity = parsePackageIdentity(JSON.parse(original) as unknown);
  const {name, version} = identity;

  if (dependencies.isPublished(identity)) {
    dependencies.log(`Skipping ${name}@${version} — already published`);
    return;
  }

  const platformMeta = dependencies.readPlatformMeta(
    dependencies.currentDirectory(),
  );
  if (platformMeta) {
    const pkg = JSON.parse(original) as Record<string, unknown>;
    pkg.os = platformMeta.os;
    pkg.cpu = platformMeta.cpu;
    if (platformMeta.libc) pkg.libc = platformMeta.libc;
    dependencies.writePackageJson(`${JSON.stringify(pkg, null, 2)}\n`);
    dependencies.log(`Injected platform fields for ${name}`);
  }

  try {
    dependencies.publish(dryRun);
  } finally {
    if (platformMeta) {
      dependencies.writePackageJson(original);
      dependencies.log(`Restored original package.json for ${name}`);
    }
  }
};

export type {
  NpmView,
  NpmViewResult,
  PackageIdentity,
  PlatformMeta,
  PublishDependencies,
};
