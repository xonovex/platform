export interface FilteredGraph {
  readonly nodes: ReadonlyMap<string, string>;
  readonly edges: readonly string[];
}

const nodeRe = /^\s+(\d+)\s+\[.*label="RunTask\(([^:]+):([^)]+)\)"/;
const edgeRe = /^\s+(\d+)\s+->\s+(\d+)/;

export const filterDotGraph = (
  dot: string,
  taskFilter: string,
): FilteredGraph => {
  const filters = new Set(taskFilter.split(","));
  const lines = dot.split("\n");
  const nodes = new Map<string, string>();
  const edges: string[] = [];

  for (const line of lines) {
    const m = nodeRe.exec(line);
    if (
      m?.[3] !== undefined &&
      filters.has(m[3]) &&
      m[1] !== undefined &&
      m[2] !== undefined
    ) {
      nodes.set(m[1], m[2]);
    }
  }

  for (const line of lines) {
    const m = edgeRe.exec(line);
    if (
      m?.[1] !== undefined &&
      m[2] !== undefined &&
      nodes.has(m[1]) &&
      nodes.has(m[2])
    ) {
      edges.push(
        `    "${nodes.get(m[1]) ?? ""}" -> "${nodes.get(m[2]) ?? ""}"`,
      );
    }
  }

  return {nodes, edges};
};

export const buildFilteredDot = (graph: FilteredGraph): string =>
  [
    "digraph {",
    "    rankdir=LR",
    "    node [shape=box, style=filled, fillcolor=lightblue]",
    ...[...graph.nodes.values()].map((name) => `    "${name}"`),
    ...graph.edges,
    "}",
  ].join("\n");
