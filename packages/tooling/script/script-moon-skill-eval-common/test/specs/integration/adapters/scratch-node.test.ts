import {existsSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {basename, dirname, join} from "node:path";
import {describe, expect, it} from "vitest";
import {nodeScratch} from "../../../../src/adapters/scratch-node.js";

describe("nodeScratch", () => {
  it("creates a new directory under the system temporary directory", () => {
    const directory = nodeScratch.makeTemporaryDirectory("xonovex-scratch-");
    try {
      expect(dirname(directory)).toBe(tmpdir());
      expect(basename(directory).startsWith("xonovex-scratch-")).toBe(true);
      expect(existsSync(directory)).toBe(true);
    } finally {
      nodeScratch.remove(directory);
    }
  });

  it("removes a tree and accepts a missing path", () => {
    const directory = nodeScratch.makeTemporaryDirectory("xonovex-scratch-");
    writeFileSync(join(directory, "file"), "text");

    nodeScratch.remove(directory);
    nodeScratch.remove(directory);

    expect(existsSync(directory)).toBe(false);
  });
});
