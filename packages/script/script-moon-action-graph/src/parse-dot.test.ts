import {describe, expect, it} from "vitest";
import {buildFilteredDot, filterDotGraph} from "./parse-dot.js";

const sampleDot = `digraph {
    0 [label="SyncWorkspace"]
    1 [label="RunTask(core:npm-publish)"]
    2 [label="RunTask(skills:npm-publish)"]
    3 [label="RunTask(core:build)"]
    0 -> 1
    0 -> 2
    1 -> 2
    1 -> 3
}`;

describe("filterDotGraph", () => {
  it("should extract nodes matching the task filter", () => {
    const {nodes} = filterDotGraph(sampleDot, "npm-publish");
    expect([...nodes.values()]).toEqual(["core", "skills"]);
  });

  it("should extract edges between matching nodes", () => {
    const {edges} = filterDotGraph(sampleDot, "npm-publish");
    expect(edges).toEqual(['    "core" -> "skills"']);
  });

  it("should support comma-separated filters", () => {
    const {nodes} = filterDotGraph(sampleDot, "npm-publish,build");
    expect([...nodes.values()]).toEqual(["core", "skills", "core"]);
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
    expect(result).toContain('"core"');
    expect(result).toContain('"skills"');
    expect(result).toContain('"core" -> "skills"');
    expect(result).toContain("rankdir=LR");
  });
});
