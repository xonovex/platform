import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/adapters/file-system-memory.js";
import {packageJsonPaths, parseMoonProjects} from "../../../src/moon-query.js";

describe("parseMoonProjects", () => {
  it("retains project identity while accepting additional Moon fields", () => {
    const projects = parseMoonProjects(
      JSON.stringify({
        projects: [
          {
            id: "example",
            source: "packages/example",
            aliases: [{alias: "@xonovex/example", plugin: "javascript"}],
          },
        ],
      }),
    );

    expect(projects).toEqual([
      {
        id: "example",
        source: "packages/example",
        aliases: [{alias: "@xonovex/example", plugin: "javascript"}],
      },
    ]);
  });

  it("rejects project entries without a source", () => {
    expect(() =>
      parseMoonProjects(JSON.stringify({projects: [{id: "example"}]})),
    ).toThrow("invalid Moon project query output");
  });

  it("reports malformed JSON from a Moon query", () => {
    expect(() => parseMoonProjects("{")).toThrow(
      "invalid Moon project query output: malformed JSON",
    );
  });
});

describe("packageJsonPaths", () => {
  it("lists the manifest of every project that holds one", () => {
    const fs = memoryFileSystem({
      files: {"/repo/packages/with/package.json": "{}"},
      directories: ["/repo/packages/without"],
    });

    expect(
      packageJsonPaths(
        "/repo",
        [
          {id: "with", source: "packages/with"},
          {id: "without", source: "packages/without"},
        ],
        fs,
      ),
    ).toEqual(["/repo/packages/with/package.json"]);
  });
});
