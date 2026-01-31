#!/usr/bin/env node
import {execSync} from "node:child_process";
import {writeFileSync} from "node:fs";
import {dirname, join} from "node:path";
import {fileURLToPath} from "node:url";
import {
  findWorkspaceRoot,
  logSuccess,
  parseCliArgs,
} from "@xonovex/moon-scripts-common";
import {buildFilteredDot, filterDotGraph} from "./parse-dot.js";

const root = findWorkspaceRoot(dirname(fileURLToPath(import.meta.url)));

const {values, positionals} = parseCliArgs({
  name: "moon-action-graph",
  description: "Generate a filtered PNG from a moon action graph",
  options: {
    target: {
      type: "string",
      short: "t",
      description: "Moon target to graph (default: :npm-publish)",
    },
    filter: {
      type: "string",
      short: "f",
      description: "Task name filter for graph nodes (default: npm-publish)",
    },
    output: {type: "string", short: "o", description: "Output PNG path"},
  },
});

const target =
  (values.target as string | undefined) ?? positionals[0] ?? ":npm-publish";
const taskFilter =
  (values.filter as string | undefined) ?? positionals[1] ?? "npm-publish";
const output =
  (values.output as string | undefined) ??
  positionals[2] ??
  join(root, "npm-publish-graph.png");

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
logSuccess(
  `Wrote ${output} (${String(graph.nodes.size)} nodes, ${String(graph.edges.length)} edges)`,
);
