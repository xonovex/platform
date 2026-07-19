import {spawnSync} from "node:child_process";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";

interface PackageIdentity {
  readonly name: string;
  readonly version: string;
}

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

export const isPublished = ({name, version}: PackageIdentity): boolean => {
  const result = spawnSync(
    resolveExecutable("npm"),
    ["view", `${name}@${version}`, "version"],
    {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "pipe"],
    },
  );
  if (result.error) throw result.error;
  if (result.status === 0) return true;
  if (/\bE404\b|404 Not Found/i.test(result.stderr)) return false;
  const errorDetail = result.stderr.trim() || `exit ${String(result.status)}`;
  throw new Error(`npm view failed for ${name}@${version}: ${errorDetail}`);
};

export type {PackageIdentity};
