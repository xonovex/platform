import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {memoryScratch} from "../../../../src/adapters/scratch-memory.js";

describe("memoryScratch", () => {
  it("creates a numbered directory per call in the FileSystem", () => {
    const fs = memoryFileSystem();
    const scratch = memoryScratch(fs, "/runs");

    const first = scratch.makeTemporaryDirectory("run-");
    const second = scratch.makeTemporaryDirectory("run-");

    expect([first, second]).toEqual(["/runs/run-1", "/runs/run-2"]);
    expect(fs.isDirectory(first)).toBe(true);
  });

  it("lists each removed path", () => {
    const scratch = memoryScratch(memoryFileSystem());

    scratch.remove("/scratch/run-1");

    expect(scratch.removed()).toEqual(["/scratch/run-1"]);
  });
});
