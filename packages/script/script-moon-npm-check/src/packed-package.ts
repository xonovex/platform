import {spawnSync} from "node:child_process";
import {posix} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {isRecord} from "@xonovex/script-moon-common/records";

type ReadPackedFile = (path: string) => string;

const normalizedPath = (path: string): string =>
  path.replaceAll("\\", "/").replace(/^\.\//, "");

const firstArrayEntry = (value: readonly unknown[]): unknown => value[0];

const collectStringTargets = (value: unknown): readonly string[] => {
  if (typeof value === "string") return [value];
  if (Array.isArray(value)) return value.flatMap(collectStringTargets);
  if (!isRecord(value)) return [];
  return Object.values(value).flatMap(collectStringTargets);
};

const targetIsPacked = (
  target: string,
  packagedFiles: ReadonlySet<string>,
): boolean => {
  if (!target.startsWith("./")) return true;
  const path = normalizedPath(target);
  const wildcardIndex = path.indexOf("*");
  if (wildcardIndex === -1) return packagedFiles.has(path);
  const prefix = path.slice(0, wildcardIndex);
  const suffix = path.slice(wildcardIndex + 1);
  return [...packagedFiles].some(
    (file) => file.startsWith(prefix) && file.endsWith(suffix),
  );
};

const binTargets = (bin: unknown): readonly string[] => {
  if (typeof bin === "string") return [bin];
  if (!isRecord(bin)) return [];
  return Object.values(bin).filter(
    (target): target is string => typeof target === "string",
  );
};

const relativeImportTargets = (source: string): readonly string[] => {
  const importPattern =
    /(?:\bfrom\s*|\bimport\s*\(\s*|\bimport\s*|\brequire\s*\(\s*)["'](\.{1,2}\/[^"']+)["']/g;
  return [...source.matchAll(importPattern)]
    .map((match) => match[1])
    .filter((target): target is string => target !== undefined);
};

const importIsPacked = (
  importingFile: string,
  target: string,
  packagedFiles: ReadonlySet<string>,
): boolean => {
  const withoutQuery = target.split(/[?#]/, 1)[0] ?? target;
  const resolved = posix.normalize(
    posix.join(posix.dirname(importingFile), withoutQuery),
  );
  return [
    resolved,
    `${resolved}.js`,
    `${resolved}.json`,
    posix.join(resolved, "index.js"),
  ].some((candidate) => packagedFiles.has(candidate));
};

export const parsePackagedFilePaths = (output: string): readonly string[] => {
  let value: unknown;
  try {
    value = JSON.parse(output);
  } catch (error) {
    throw new Error("npm pack returned malformed JSON", {cause: error});
  }
  const result = Array.isArray(value) ? firstArrayEntry(value) : undefined;
  const files = isRecord(result) ? result.files : undefined;
  if (!Array.isArray(files)) {
    throw new TypeError("npm pack did not report packaged files");
  }
  return files.map((file) => {
    if (!isRecord(file) || typeof file.path !== "string") {
      throw new TypeError("npm pack reported an invalid file entry");
    }
    return normalizedPath(file.path);
  });
};

export const packagedFilePaths = (packageRoot: string): readonly string[] => {
  const result = spawnSync(
    resolveExecutable("npm"),
    ["pack", "--dry-run", "--json"],
    {
      cwd: packageRoot,
      encoding: "utf8",
      maxBuffer: 10_485_760,
    },
  );
  if (result.error) throw result.error;
  if (result.status !== 0) {
    const detail = result.stderr.trim() || `exit ${String(result.status)}`;
    throw new Error(`npm pack failed: ${detail}`);
  }
  return parsePackagedFilePaths(result.stdout);
};

export const validatePackedPackage = (
  pkg: PackageJson,
  files: readonly string[],
  readFile: ReadPackedFile,
): readonly string[] => {
  const packagedFiles = new Set(files.map(normalizedPath));
  const errors: string[] = [];
  const manifestTargets = [
    ...collectStringTargets(pkg.exports).map(
      (target) => ["export", target] as const,
    ),
    ...binTargets(pkg.bin).map((target) => ["bin", target] as const),
    ...[pkg.main, pkg.module, pkg.types]
      .filter((target): target is string => target !== undefined)
      .map((target) => ["manifest", target] as const),
  ];

  for (const [kind, target] of manifestTargets) {
    if (!targetIsPacked(target, packagedFiles)) {
      errors.push(`Packed package is missing ${kind} target: "${target}"`);
    }
  }

  for (const file of packagedFiles) {
    if (!/\.[cm]?js$/.test(file)) continue;
    const source = readFile(file);
    for (const target of relativeImportTargets(source)) {
      if (!importIsPacked(file, target, packagedFiles)) {
        errors.push(
          `Packed file "${file}" imports missing relative target: "${target}"`,
        );
      }
    }
  }

  return errors;
};
