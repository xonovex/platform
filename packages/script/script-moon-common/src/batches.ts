export const boundedBatches = <T>(
  items: readonly T[],
  batchSize: number,
): readonly (readonly T[])[] => {
  if (!Number.isInteger(batchSize) || batchSize < 1) {
    throw new Error("batch size must be a positive integer");
  }

  const batches: T[][] = [];
  for (let offset = 0; offset < items.length; offset += batchSize) {
    batches.push(items.slice(offset, offset + batchSize));
  }
  return batches;
};
