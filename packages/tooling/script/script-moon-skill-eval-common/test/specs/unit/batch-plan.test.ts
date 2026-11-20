import {describe, expect, it} from "vitest";
import {planBatches} from "../../../src/batch-plan.js";

describe("planBatches", () => {
  it("bounds each batch and reports the runs of the largest", () => {
    expect(planBatches([1, 2, 3], 2, (count) => count * 3)).toEqual({
      batches: [[1, 2], [3]],
      maxBatchRuns: 6,
    });
  });

  it("keeps one batch without a size and none without items", () => {
    expect(planBatches([1, 2, 3], undefined, (count) => count)).toEqual({
      batches: [[1, 2, 3]],
      maxBatchRuns: 3,
    });
    expect(planBatches([], undefined, (count) => count).maxBatchRuns).toBe(0);
  });
});
