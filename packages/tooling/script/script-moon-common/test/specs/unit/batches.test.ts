import {describe, expect, it} from "vitest";
import {boundedBatches} from "../../../src/batches.js";

describe("boundedBatches", () => {
  it("preserves item order while bounding every batch", () => {
    const batches = boundedBatches([1, 2, 3, 4, 5], 2);

    expect(batches).toEqual([[1, 2], [3, 4], [5]]);
  });

  it.each([0, -1, 1.5])("rejects invalid batch size %s", (batchSize) => {
    expect(() => boundedBatches([1], batchSize)).toThrow(
      "batch size must be a positive integer",
    );
  });
});
