import {boundedBatches} from "@xonovex/script-moon-common/batches";

export interface BatchPlan<T> {
  readonly batches: readonly (readonly T[])[];
  // The model runs of the largest batch, which the per-batch cap is checked against.
  readonly maxBatchRuns: number;
}

/** Splits items into batches of at most batchSize; no size keeps one batch. */
export const planBatches = <T>(
  items: readonly T[],
  batchSize: number | undefined,
  modelRuns: (itemCount: number) => number,
): BatchPlan<T> => {
  const batches = boundedBatches(items, batchSize ?? Math.max(items.length, 1));
  return {
    batches,
    maxBatchRuns: Math.max(
      0,
      ...batches.map((batch) => modelRuns(batch.length)),
    ),
  };
};
