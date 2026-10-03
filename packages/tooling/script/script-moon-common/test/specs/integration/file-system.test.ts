import {chmodSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {nodeFileSystem} from "../../../src/file-system.js";

// nodeFileSystem is the one FileSystem implementation the unit tier cannot drive,
// because it is the adapter onto a real disk. These cases pin its behaviour so the
// in-memory implementation the unit tier uses is answering the same questions.
describe("nodeFileSystem", () => {
  let root: string;

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "moon-common-file-system-"));
  });

  afterEach(() => {
    rmSync(root, {recursive: true, force: true});
  });

  it("distinguishes files, directories, and missing paths", () => {
    const file = join(root, "file.txt");
    writeFileSync(file, "data");

    expect(nodeFileSystem.isFile(file)).toBe(true);
    expect(nodeFileSystem.isFile(root)).toBe(false);
    expect(nodeFileSystem.isDirectory(root)).toBe(true);
    expect(nodeFileSystem.isDirectory(file)).toBe(false);
    expect(nodeFileSystem.isFile(join(root, "missing"))).toBe(false);
    expect(nodeFileSystem.isDirectory(join(root, "missing"))).toBe(false);
  });

  it("reads a file and throws for one that is absent", () => {
    writeFileSync(join(root, "file.txt"), "data");

    expect(nodeFileSystem.readText(join(root, "file.txt"))).toBe("data");
    expect(() => nodeFileSystem.readText(join(root, "missing"))).toThrow(
      "ENOENT",
    );
  });

  it("lists the entries directly beneath a directory", () => {
    writeFileSync(join(root, "a.txt"), "a");
    nodeFileSystem.makeDirectory(join(root, "nested", "deep"));

    expect(nodeFileSystem.readDirectory(root).toSorted()).toEqual([
      "a.txt",
      "nested",
    ]);
    expect(() => nodeFileSystem.readDirectory(join(root, "missing"))).toThrow(
      "ENOENT",
    );
  });

  it("creates the directories above a file it writes", () => {
    const file = join(root, "created", "nested", "file.txt");

    nodeFileSystem.writeFile(file, "written");

    expect(nodeFileSystem.readText(file)).toBe("written");
    expect(nodeFileSystem.isDirectory(join(root, "created", "nested"))).toBe(
      true,
    );
  });

  it("reports only an executable file as executable", () => {
    const executable = join(root, "run");
    const plain = join(root, "data.txt");
    writeFileSync(executable, "#!/bin/sh\n");
    chmodSync(executable, 0o755);
    writeFileSync(plain, "text");

    expect(nodeFileSystem.isExecutableFile(executable)).toBe(true);
    expect(nodeFileSystem.isExecutableFile(plain)).toBe(false);
    expect(nodeFileSystem.isExecutableFile(join(root, "missing"))).toBe(false);
  });

  it("does not report an executable directory as an executable file", () => {
    // A directory carries the execute bit to allow traversal, so a bare access
    // check would accept one; isExecutableFile has to require a file as well.
    expect(nodeFileSystem.isExecutableFile(root)).toBe(false);
  });
});
