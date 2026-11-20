/**
 * The filesystem operations this workspace's scripts reach for. Every module that
 * walks or reads a tree takes this as a port, so the unit tier drives it from
 * adapters/file-system-memory.ts and only the integration tier touches a real
 * disk through adapters/file-system-node.ts.
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
  /**
   * Copies the tree under source to target byte for byte, creating target and the
   * directories above it. A skill directory can hold binary resources, which a
   * read and a write of text would corrupt.
   */
  readonly copyDirectory: (source: string, target: string) => void;
}
