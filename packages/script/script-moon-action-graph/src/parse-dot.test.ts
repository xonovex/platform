import {describe, expect, it} from "vitest";
import {buildFilteredDot, filterDotGraph} from "./parse-dot.js";

const sampleDot = `digraph {
    0 [label="SyncWorkspace"]
    1 [label="RunTask(core:npm-publish)"]
    2 [label="RunTask(skills:npm-publish)"]
    3 [label="RunTask(core:build)"]
    4 [label="RunTask(config:compile)"]
    0 -> 1
    0 -> 2
    1 -> 2
    1 -> 3
    3 -> 4
}`;

describe("filterDotGraph", () => {
  it("should extract nodes matching the task filter", () => {
    const {nodes} = filterDotGraph(sampleDot, "npm-publish");
    expect([...nodes.values()]).toEqual([
      "core:npm-publish",
      "skills:npm-publish",
      "core:build",
      "config:compile",
    ]);
  });

  it("should preserve the full dependency closure and its edges", () => {
    const {edges} = filterDotGraph(sampleDot, "npm-publish");
    expect(edges).toEqual([
      '    "core:npm-publish" -> "skills:npm-publish"',
      '    "core:npm-publish" -> "core:build"',
      '    "core:build" -> "config:compile"',
    ]);
  });

  it("should support comma-separated filters", () => {
    const {nodes} = filterDotGraph(sampleDot, "npm-publish,build");
    expect([...nodes.values()]).toEqual([
      "core:npm-publish",
      "skills:npm-publish",
      "core:build",
      "config:compile",
    ]);
  });

  it("should return empty for non-matching filter", () => {
    const {nodes, edges} = filterDotGraph(sampleDot, "lint");
    expect(nodes.size).toBe(0);
    expect(edges).toEqual([]);
  });
});

describe("buildFilteredDot", () => {
  it("should produce valid DOT output", () => {
    const graph = filterDotGraph(sampleDot, "npm-publish");
    const result = buildFilteredDot(graph);
    expect(result).toContain("digraph {");
    expect(result).toContain('"core:npm-publish"');
    expect(result).toContain('"skills:npm-publish"');
    expect(result).toContain('"core:npm-publish" -> "skills:npm-publish"');
    expect(result).toContain("rankdir=LR");
  });
});
