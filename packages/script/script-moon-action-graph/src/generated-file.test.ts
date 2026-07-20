import {mkdtempSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {writeOrCheckGeneratedFile} from "./generated-file.js";

describe("writeOrCheckGeneratedFile", () => {
  const temporaryDirectories: string[] = [];

  afterEach(() => {
    for (const directory of temporaryDirectories) {
      rmSync(directory, {recursive: true, force: true});
    }
    temporaryDirectories.length = 0;
  });

  function temporaryFile(): string {
    const directory = mkdtempSync(join(tmpdir(), "moon-action-graph-"));
    temporaryDirectories.push(directory);
    return join(directory, "graph.png");
  }

  it("keeps a current generated file unchanged in check mode", () => {
    const filePath = temporaryFile();
    writeFileSync(filePath, "current");

    writeOrCheckGeneratedFile(filePath, "current", true);

    expect(readFileSync(filePath, "utf8")).toBe("current");
  });

  it("rejects a stale generated file in check mode", () => {
    const filePath = temporaryFile();
    writeFileSync(filePath, "stale");

    expect(() => writeOrCheckGeneratedFile(filePath, "current", true)).toThrow(
      `Generated file is stale: ${filePath}`,
    );
  });

  it("rejects a missing generated file in check mode", () => {
    const filePath = temporaryFile();

    expect(() => writeOrCheckGeneratedFile(filePath, "current", true)).toThrow(
      `Generated file is missing: ${filePath}`,
    );
  });

  it("writes a generated file outside check mode", () => {
    const filePath = temporaryFile();

    writeOrCheckGeneratedFile(filePath, "current", false);

    expect(readFileSync(filePath, "utf8")).toBe("current");
  });
});
