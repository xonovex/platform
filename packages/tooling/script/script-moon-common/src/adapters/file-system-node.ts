import {
  accessSync,
  constants,
  cpSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import {dirname} from "node:path";
import type {FileSystem} from "../file-system.js";

/**
 * The adapter that reaches the real disk. It is the one implementation the unit
 * tier cannot drive, so test/specs/integration/adapters/file-system-node.test.ts
 * covers it.
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
  copyDirectory: (source, target) => {
    cpSync(source, target, {recursive: true});
  },
};
