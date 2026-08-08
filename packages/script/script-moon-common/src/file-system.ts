import {
  accessSync,
  constants,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import {dirname} from "node:path";

/**
 * The filesystem operations this workspace's scripts reach for. Every module that
 * walks or reads a tree takes this as a port, so the unit tier drives it from the
 * in-memory implementation in file-system-memory.ts and only the integration tier
 * touches a real disk.
 */
export interface FileSystem {
  /** Returns the file's text, throwing when it is absent or is a directory. */
  readonly readText: (path: string) => string;
  /** Returns the entry names directly under path, throwing when it is not a directory. */
  readonly readDirectory: (path: string) => readonly string[];
  readonly isFile: (path: string) => boolean;
  readonly isDirectory: (path: string) => boolean;
  /** Whether path is a file the current process may execute. */
  readonly isExecutableFile: (path: string) => boolean;
  /** Writes text to path, creating the directories above it. */
  readonly writeFile: (path: string, text: string) => void;
  readonly makeDirectory: (path: string) => void;
}

/**
 * The adapter that reaches the real disk. It is the one implementation the unit
 * tier cannot drive, so test/specs/integration/file-system.test.ts covers it.
 */
export const nodeFileSystem: FileSystem = {
  readText: (path) => readFileSync(path, "utf8"),
  readDirectory: (path) => readdirSync(path),
  isFile: (path) => {
    try {
      return statSync(path).isFile();
    } catch {
      return false;
    }
  },
  isDirectory: (path) => {
    try {
      return statSync(path).isDirectory();
    } catch {
      return false;
    }
  },
  isExecutableFile: (path) => {
    try {
      accessSync(path, constants.X_OK);
      return statSync(path).isFile();
    } catch {
      return false;
    }
  },
  writeFile: (path, text) => {
    mkdirSync(dirname(path), {recursive: true});
    writeFileSync(path, text, "utf8");
  },
  makeDirectory: (path) => {
    mkdirSync(path, {recursive: true});
  },
};
