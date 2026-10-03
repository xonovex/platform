export interface CatalogRecord {
  readonly type: string;
  readonly name: string;
  readonly description: string;
  readonly plugin: string;
  readonly harnesses: readonly string[];
}

export interface CatalogFilters {
  readonly type: string;
  readonly query: string;
  readonly plugin: string;
  readonly harness: string;
}

export const filterComponents = <T extends CatalogRecord>(
  entries: readonly T[],
  filters: CatalogFilters,
): readonly T[] => {
  const terms = filters.query
    .trim()
    .toLowerCase()
    .split(/\s+/u)
    .filter(Boolean);
  return entries.filter(
    (entry) =>
      entry.type === filters.type &&
      (filters.plugin === "" || entry.plugin === filters.plugin) &&
      (filters.harness === "" || entry.harnesses.includes(filters.harness)) &&
      terms.every((term) =>
        (entry.name + " " + entry.description + " " + entry.plugin)
          .toLowerCase()
          .includes(term),
      ),
  );
};
