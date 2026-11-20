import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {memoryScratch} from "../../../src/adapters/scratch-memory.js";
import {
  clearEvidence,
  writeJsonEvidence,
  writeJsonLinesEvidence,
} from "../../../src/evidence.js";

describe("evidence files", () => {
  it("writes indented JSON and JSON lines, each ending in a newline", () => {
    const fs = memoryFileSystem();

    writeJsonEvidence(fs, "/out/summary.json", {passed: 1});
    writeJsonLinesEvidence(fs, "/out/results.jsonl", [{a: 1}, {b: 2}]);

    expect(fs.readText("/out/summary.json")).toBe('{\n  "passed": 1\n}\n');
    expect(fs.readText("/out/results.jsonl")).toBe('{"a":1}\n{"b":2}\n');
  });

  it("removes every stale evidence file", () => {
    const scratch = memoryScratch(memoryFileSystem());

    clearEvidence(scratch, ["/out/summary.json", "/out/invalid-run.json"]);

    expect(scratch.removed()).toEqual([
      "/out/summary.json",
      "/out/invalid-run.json",
    ]);
  });
});
