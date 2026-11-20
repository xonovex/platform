import {mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import type {Scratch} from "../scratch.js";

/** The Scratch that creates directories under the system temporary directory. */
export const nodeScratch: Scratch = {
  makeTemporaryDirectory: (prefix) => mkdtempSync(join(tmpdir(), prefix)),
  remove: (path) => {
    rmSync(path, {recursive: true, force: true});
  },
};
