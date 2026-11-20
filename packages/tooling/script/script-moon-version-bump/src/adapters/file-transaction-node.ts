import {randomUUID} from "node:crypto";
import {existsSync, renameSync, rmSync, writeFileSync} from "node:fs";
import {
  applyFileChanges,
  type FileChange,
  type FileTransactionIo,
} from "../file-transaction.js";

/** The transaction steps on a real disk. */
export const nodeFileTransactionIo: FileTransactionIo = {
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

export const applyFileChangesOnDisk = (
  changes: readonly FileChange[],
): void => {
  applyFileChanges(changes, nodeFileTransactionIo);
};
