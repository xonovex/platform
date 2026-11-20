import {describe, expect, it} from "vitest";
import {
  foldLines,
  ignoreLines,
  splitCompleteLines,
  type LineFold,
} from "../../../src/process-runner.js";

// Counts lines until it has seen two, the smallest fold that settles part way.
const countToTwo: LineFold<number> = {
  initial: 0,
  step: (count) => count + 1,
  settled: (count) => count >= 2,
};

describe("splitCompleteLines", () => {
  it("keeps the partial line after the last break", () => {
    expect(splitCompleteLines("one\r\ntwo\nthr")).toEqual({
      lines: ["one", "two"],
      rest: "thr",
    });
  });

  it("returns no complete line for text without a break", () => {
    expect(splitCompleteLines("partial")).toEqual({lines: [], rest: "partial"});
  });
});

describe("foldLines", () => {
  it("stops at the first settled state", () => {
    expect(foldLines(countToTwo, 0, ["a", "b", "c", "d"])).toBe(2);
  });

  it("leaves a settled state untouched", () => {
    expect(foldLines(countToTwo, 2, ["a"])).toBe(2);
  });
});

describe("ignoreLines", () => {
  it("never settles and keeps its state", () => {
    expect(foldLines(ignoreLines, ignoreLines.initial, ["a", "b"])).toBeNull();
    expect(ignoreLines.settled(null)).toBe(false);
  });
});
