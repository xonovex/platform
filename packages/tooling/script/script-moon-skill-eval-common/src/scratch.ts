/**
 * The disk operations an evaluator needs beyond the FileSystem port: a private
 * directory for each harness run, and removal of that directory and of stale
 * evidence. adapters/scratch-node.ts reaches the real disk, and
 * adapters/scratch-memory.ts works on a memory FileSystem.
 */
export interface Scratch {
  /** Creates a new empty directory whose name starts with prefix. */
  readonly makeTemporaryDirectory: (prefix: string) => string;
  /** Removes path and everything under it; a missing path is not an error. */
  readonly remove: (path: string) => void;
}
