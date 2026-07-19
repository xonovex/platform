export interface FilteredGraph {
  readonly nodes: ReadonlyMap<string, string>;
  readonly edges: readonly string[];
}

const nodeRe = /^\s+(\d+)\s+\[.*label="RunTask\(([^:]+):([^)]+)\)"/;
const edgeRe = /^\s+(\d+)\s+->\s+(\d+)/;

interface TaskNode {
  readonly project: string;
  readonly task: string;
}

interface Edge {
  readonly from: string;
  readonly to: string;
}

export const filterDotGraph = (
  dot: string,
  taskFilter: string,
): FilteredGraph => {
  const filters = new Set(
    taskFilter
      .split(",")
      .map((filter) => filter.trim())
      .filter(Boolean),
  );
  const lines = dot.split("\n");
  const allNodes = new Map<string, TaskNode>();
  const allEdges: Edge[] = [];

  for (const line of lines) {
    const m = nodeRe.exec(line);
    if (m?.[1] !== undefined && m[2] !== undefined && m[3] !== undefined) {
      allNodes.set(m[1], {project: m[2], task: m[3]});
    }
  }

  for (const line of lines) {
    const m = edgeRe.exec(line);
    if (m?.[1] !== undefined && m[2] !== undefined) {
      allEdges.push({from: m[1], to: m[2]});
    }
  }

  const selected = new Set<string>();
  const pending: string[] = [];
  for (const [id, node] of allNodes) {
    if (filters.has(node.task)) {
      selected.add(id);
      pending.push(id);
    }
  }

  while (pending.length > 0) {
    const id = pending.pop();
    if (id === undefined) break;
    for (const edge of allEdges) {
      if (edge.from === id && allNodes.has(edge.to) && !selected.has(edge.to)) {
        selected.add(edge.to);
        pending.push(edge.to);
      }
    }
  }

  const nodes = new Map<string, string>();
  for (const [id, node] of allNodes) {
    if (selected.has(id)) nodes.set(id, `${node.project}:${node.task}`);
  }
  const edges = allEdges
    .filter((edge) => selected.has(edge.from) && selected.has(edge.to))
    .map(
      (edge) =>
        `    "${nodes.get(edge.from) ?? ""}" -> "${nodes.get(edge.to) ?? ""}"`,
    );

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
