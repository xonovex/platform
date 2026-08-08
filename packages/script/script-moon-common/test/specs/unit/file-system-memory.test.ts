import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/file-system-memory.js";

describe("memoryFileSystem", () => {
  it("reads a file it was seeded with", () => {
    const fs = memoryFileSystem({files: {"/a/b.txt": "data"}});

    expect(fs.readText("/a/b.txt")).toBe("data");
    expect(fs.isFile("/a/b.txt")).toBe(true);
  });

  it("throws when reading an absent file", () => {
    expect(() => memoryFileSystem().readText("/missing")).toThrow("ENOENT");
  });

  it("implies every directory above a seeded file", () => {
    const fs = memoryFileSystem({files: {"/a/b/c/d.txt": "data"}});

    for (const directory of ["/a", "/a/b", "/a/b/c", "/"]) {
      expect(fs.isDirectory(directory)).toBe(true);
    }
    expect(fs.isFile("/a/b")).toBe(false);
  });

  it("holds a directory that carries no file", () => {
    const fs = memoryFileSystem({directories: ["/a/empty"]});

    expect(fs.isDirectory("/a/empty")).toBe(true);
    expect(fs.readDirectory("/a/empty")).toEqual([]);
    expect(fs.readDirectory("/a")).toEqual(["empty"]);
  });

  it("lists only the names directly beneath a directory, sorted", () => {
    const fs = memoryFileSystem({
      files: {
        "/root/second.txt": "b",
        "/root/first.txt": "a",
        "/root/nested/deep/file.txt": "c",
      },
    });

    expect(fs.readDirectory("/root")).toEqual([
      "first.txt",
      "nested",
      "second.txt",
    ]);
    expect(fs.readDirectory("/root/nested")).toEqual(["deep"]);
  });

  it("throws when listing something that is not a directory", () => {
    const fs = memoryFileSystem({files: {"/root/file.txt": "a"}});

    expect(() => fs.readDirectory("/root/file.txt")).toThrow("ENOENT");
    expect(() => fs.readDirectory("/missing")).toThrow("ENOENT");
  });

  it("ignores a trailing separator on a directory path", () => {
    const fs = memoryFileSystem({files: {"/root/file.txt": "a"}});

    expect(fs.isDirectory("/root/")).toBe(true);
    expect(fs.readDirectory("/root/")).toEqual(["file.txt"]);
  });

  it("reads back what it wrote and implies the directories above it", () => {
    const fs = memoryFileSystem();

    fs.writeFile("/out/nested/file.txt", "written");

    expect(fs.readText("/out/nested/file.txt")).toBe("written");
    expect(fs.isDirectory("/out/nested")).toBe(true);
    expect(fs.readDirectory("/out")).toEqual(["nested"]);
  });

  it("overwrites a file that already exists", () => {
    const fs = memoryFileSystem({files: {"/a.txt": "first"}});

    fs.writeFile("/a.txt", "second");

    expect(fs.readText("/a.txt")).toBe("second");
  });

  it("creates a directory and every directory above it", () => {
    const fs = memoryFileSystem();

    fs.makeDirectory("/deep/nested/leaf");

    expect(fs.isDirectory("/deep/nested/leaf")).toBe(true);
    expect(fs.isDirectory("/deep")).toBe(true);
  });

  it("treats only a seeded executable file as executable", () => {
    const fs = memoryFileSystem({
      files: {"/bin/run": "#!/bin/sh\n", "/bin/data": "text"},
      executables: ["/bin/run", "/bin/absent"],
      directories: ["/bin/dir"],
    });

    expect(fs.isExecutableFile("/bin/run")).toBe(true);
    expect(fs.isExecutableFile("/bin/data")).toBe(false);
    expect(fs.isExecutableFile("/bin/absent")).toBe(false);
    expect(fs.isExecutableFile("/bin/dir")).toBe(false);
  });
});
