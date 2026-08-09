import {z} from "zod";

export interface SourceFile {
  /** Repository-relative path, which is how a failure names the file. */
  readonly path: string;
  readonly text: string;
}

const FS_IMPORT_RE =
  /^\s*(?:import|export)[^;]*?from\s+["']node:fs(?:\/promises)?["']/mu;

/**
 * The modules allowed to reach node:fs directly, read from the repository's
 * filesystem-allowlist.json so the release gate and the editor rule cannot
 * disagree about which they are. Each entry carries the reason the module cannot
 * go through the FileSystem port, so adding one is a reviewed decision rather
 * than a silent import.
 */
export const FilesystemAllowlistSchema = z.record(
  z.string(),
  z.string().min(1),
);

export type FilesystemAllowlist = z.infer<typeof FilesystemAllowlistSchema>;

export const FILESYSTEM_ALLOWLIST_FILE = "filesystem-allowlist.json";

export const parseFilesystemAllowlist = (
  text: string,
): {readonly allowed: FilesystemAllowlist; readonly error?: string} => {
  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    return {
      allowed: {},
      error: `${FILESYSTEM_ALLOWLIST_FILE} is not valid JSON: ${detail}`,
    };
  }
  const result = FilesystemAllowlistSchema.safeParse(parsed);
  return result.success
    ? {allowed: result.data}
    : {
        allowed: {},
        error: `${FILESYSTEM_ALLOWLIST_FILE} must map a repository-relative path to the reason it cannot use the port`,
      };
};

/**
 * Reports a module under src that imports node:fs without being named above. The
 * port is only a boundary while every reader goes through it, and a direct import
 * puts a real disk back into whatever tier reaches that module.
 */
export const filesystemImportFailures = (
  files: readonly SourceFile[],
  allowed: FilesystemAllowlist,
): readonly string[] =>
  files
    .filter(
      (file) =>
        allowed[file.path] === undefined && FS_IMPORT_RE.test(file.text),
    )
    .map(
      (file) =>
        `${file.path} imports node:fs directly: take the FileSystem port from @xonovex/script-moon-common/file-system as a defaulted last parameter, or add the file to filesystem-allowlist.json with the reason it cannot`,
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
  allowed: FilesystemAllowlist,
): readonly string[] => {
  const scanned = new Map(files.map((file) => [file.path, file.text]));
  return Object.keys(allowed)
    .filter((path) => {
      const text = scanned.get(path);
      return text !== undefined && !FS_IMPORT_RE.test(text);
    })
    .map(
      (path) =>
        `${path} is allowed to import node:fs but no longer does: drop it from filesystem-allowlist.json`,
    )
    .toSorted();
};
