/** A file the scan read, or the reason it could not. */
export type FileRead =
  | {readonly kind: "read"; readonly text: string}
  | {readonly kind: "unreadable"; readonly message: string};

export type ParsedJson =
  | {readonly kind: "parsed"; readonly value: unknown}
  | {readonly kind: "invalid"; readonly message: string};

const errorMessage = (error: unknown): string =>
  error instanceof Error ? error.message : String(error);

/** Parses a read file as JSON; an unreadable file is invalid with its read error. */
export const parseJsonFile = (file: FileRead): ParsedJson => {
  if (file.kind === "unreadable") {
    return {kind: "invalid", message: file.message};
  }
  try {
    return {kind: "parsed", value: JSON.parse(file.text) as unknown};
  } catch (error) {
    return {kind: "invalid", message: errorMessage(error)};
  }
};

// The FileSystem port reports a missing file with this message, so a file the
// scan never reached reads the same as one it failed to open.
export const missingFile = (path: string): FileRead => ({
  kind: "unreadable",
  message: `ENOENT: no such file or directory, open '${path}'`,
});
