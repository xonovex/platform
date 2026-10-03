import {randomUUID} from "node:crypto";
import {existsSync, renameSync, rmSync, writeFileSync} from "node:fs";
import {basename, dirname, join} from "node:path";

interface FileChange {
  readonly path: string;
  readonly content: string;
}

interface FileTransactionIo {
  readonly exists: (path: string) => boolean;
  readonly write: (path: string, content: string) => void;
  readonly rename: (from: string, to: string) => void;
  readonly remove: (path: string) => void;
  readonly id: () => string;
}

interface StagedChange extends FileChange {
  readonly temporaryPath: string;
  readonly backupPath: string;
  readonly existed: boolean;
  backedUp: boolean;
  installed: boolean;
}

const defaultIo: FileTransactionIo = {
  exists: existsSync,
  write: (path, content) => {
    writeFileSync(path, content, {encoding: "utf8", flag: "wx"});
  },
  rename: renameSync,
  remove: (path) => {
    rmSync(path, {force: true});
  },
  id: randomUUID,
};

const transactionPath = (
  target: string,
  id: string,
  suffix: "backup" | "temporary",
): string => join(dirname(target), `.${basename(target)}.${id}.${suffix}`);

const validateChanges = (changes: readonly FileChange[]): void => {
  const paths = new Set<string>();
  for (const change of changes) {
    if (paths.has(change.path)) {
      throw new Error(`duplicate file change for ${change.path}`);
    }
    paths.add(change.path);
  }
};

const rollback = (
  staged: readonly StagedChange[],
  io: FileTransactionIo,
): readonly Error[] => {
  const errors: Error[] = [];
  for (const change of staged.toReversed()) {
    try {
      if (change.installed) io.remove(change.path);
      if (change.backedUp) io.rename(change.backupPath, change.path);
      io.remove(change.temporaryPath);
    } catch (error) {
      errors.push(error instanceof Error ? error : new Error(String(error)));
    }
  }
  return errors;
};

const applyFileChanges = (
  changes: readonly FileChange[],
  io: FileTransactionIo = defaultIo,
): void => {
  validateChanges(changes);
  const staged: StagedChange[] = changes.map((change) => {
    const id = io.id();
    return {
      ...change,
      temporaryPath: transactionPath(change.path, id, "temporary"),
      backupPath: transactionPath(change.path, id, "backup"),
      existed: io.exists(change.path),
      backedUp: false,
      installed: false,
    };
  });

  try {
    for (const change of staged) {
      io.write(change.temporaryPath, change.content);
    }
    for (const change of staged) {
      if (change.existed) {
        io.rename(change.path, change.backupPath);
        change.backedUp = true;
      }
      io.rename(change.temporaryPath, change.path);
      change.installed = true;
    }
  } catch (error) {
    const rollbackErrors = rollback(staged, io);
    if (rollbackErrors.length > 0) {
      throw new AggregateError(
        [error, ...rollbackErrors],
        "failed to apply file changes and restore every original file",
      );
    }
    throw new Error("failed to apply file changes; original files restored", {
      cause: error,
    });
  }

  for (const change of staged) {
    if (change.backedUp) io.remove(change.backupPath);
  }
};

export {applyFileChanges};
export type {FileChange, FileTransactionIo};
