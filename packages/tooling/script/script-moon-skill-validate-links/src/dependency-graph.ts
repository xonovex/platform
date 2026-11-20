// A cycle reads the same from any start, so it is named from its smallest
// rotation and reported once.
const canonicalCycle = (cycle: readonly string[]): string => {
  const withoutRepeatedEnd = cycle.slice(0, -1);
  const canonical = withoutRepeatedEnd
    .map((_, index) => [
      ...withoutRepeatedEnd.slice(index),
      ...withoutRepeatedEnd.slice(0, index),
    ])
    .toSorted((left, right) => left.join("\0").localeCompare(right.join("\0")))
    .at(0);
  const first = canonical?.at(0);
  return canonical === undefined || first === undefined
    ? ""
    : [...canonical, first].join(" → ");
};

/** The dependency cycles of a graph, each as "a → b → a"; unknown nodes are ignored. */
export const dependencyCycles = (
  graph: ReadonlyMap<string, readonly string[]>,
): ReadonlySet<string> => {
  const state = new Map<string, "visiting" | "visited">();
  const cycles = new Set<string>();
  const visit = (name: string, path: readonly string[]): void => {
    if (state.get(name) === "visited") return;
    if (state.get(name) === "visiting") {
      cycles.add(canonicalCycle([...path.slice(path.indexOf(name)), name]));
      return;
    }
    state.set(name, "visiting");
    for (const dependency of graph.get(name) ?? []) {
      if (graph.has(dependency)) visit(dependency, [...path, name]);
    }
    state.set(name, "visited");
  };
  for (const name of graph.keys()) visit(name, []);
  return cycles;
};
