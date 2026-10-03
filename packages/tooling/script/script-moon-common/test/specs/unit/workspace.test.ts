import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/file-system-memory.js";
import {findWorkspaceRoot} from "../../../src/workspace.js";

describe("findWorkspaceRoot", () => {
  it("should find the directory containing .moon", () => {
    const fs = memoryFileSystem({
      directories: ["/repo/.moon", "/repo/a/b/c"],
    });

    expect(findWorkspaceRoot("/repo/a/b/c", undefined, fs)).toBe("/repo");
  });

  it("should return start dir if it contains .moon", () => {
    const fs = memoryFileSystem({directories: ["/repo/.moon"]});

    expect(findWorkspaceRoot("/repo", undefined, fs)).toBe("/repo");
  });

  it("should stop at the boundary when no .moon directory is found", () => {
    const fs = memoryFileSystem({directories: ["/repo/a"]});

    expect(() => findWorkspaceRoot("/repo/a", "/repo/a", fs)).toThrow(
      "Could not find workspace root",
    );
  });

  it("should stop at the filesystem root when no boundary is given", () => {
    const fs = memoryFileSystem({directories: ["/repo/a"]});

    expect(() => findWorkspaceRoot("/repo/a", undefined, fs)).toThrow(
      "Could not find workspace root",
    );
  });
});
