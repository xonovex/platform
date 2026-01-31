#!/usr/bin/env node
import {execSync} from "node:child_process";
import {writeFileSync} from "node:fs";
import {dirname, join} from "node:path";
import {fileURLToPath} from "node:url";
import {findWorkspaceRoot, logSuccess} from "@xonovex/moon-scripts-common";
import {buildFilteredDot, filterDotGraph} from "./parse-dot.js";

const root = findWorkspaceRoot(dirname(fileURLToPath(import.meta.url)));

const target = process.argv[2] ?? ":npm-publish";
const taskFilter = process.argv[3] ?? "npm-publish";
const output = process.argv[4] ?? join(root, "npm-publish-graph.png");

const dot = execSync(`npx moon action-graph ${target} --dot`, {
  cwd: root,
  encoding: "utf8",
  stdio: ["pipe", "pipe", "pipe"],
});

const graph = filterDotGraph(dot, taskFilter);
const filtered = buildFilteredDot(graph);

// eslint-disable-next-line sonarjs/no-os-command-from-path
const png = execSync("dot -Tpng", {
  input: filtered,
  maxBuffer: 10 * 1024 * 1024,
});

writeFileSync(output, png);
logSuccess(`Wrote ${output} (${String(graph.nodes.size)} nodes, ${String(graph.edges.length)} edges)`);
