import {execSync} from "node:child_process";
import {writeFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dirname, "..");

const target = process.argv[2] ?? ":npm-publish";
const taskFilter = process.argv[3] ?? "npm-publish";
const output = process.argv[4] ?? join(root, "npm-publish-graph.png");

const dot = execSync(`npx moon action-graph ${target} --dot`, {
	cwd: root,
	encoding: "utf-8",
	stdio: ["pipe", "pipe", "pipe"],
});

const lines = dot.split("\n");
const nodes = new Map<string, string>();
const edges: string[] = [];

const nodeRe = /^\s+(\d+)\s+\[.*label="RunTask\(([^:]+):([^)]+)\)"/;
const edgeRe = /^\s+(\d+)\s+->\s+(\d+)/;

for (const line of lines) {
	const m = nodeRe.exec(line);
	if (m && m[3] === taskFilter) {
		nodes.set(m[1], m[2]);
	}
}

for (const line of lines) {
	const m = edgeRe.exec(line);
	if (m && nodes.has(m[1]) && nodes.has(m[2])) {
		edges.push(`    "${nodes.get(m[1])}" -> "${nodes.get(m[2])}"`);
	}
}

const filtered = [
	"digraph {",
	"    rankdir=LR",
	"    node [shape=box, style=filled, fillcolor=lightblue]",
	...[...nodes.values()].map((name) => `    "${name}"`),
	...edges,
	"}",
].join("\n");

const png = execSync("dot -Tpng", {
	input: filtered,
	maxBuffer: 10 * 1024 * 1024,
});

writeFileSync(output, png);
console.log(`Wrote ${output} (${nodes.size} nodes, ${edges.length} edges)`);
