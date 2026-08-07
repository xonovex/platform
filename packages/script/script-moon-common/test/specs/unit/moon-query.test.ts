import {describe, expect, it} from "vitest";
import {parseMoonProjects} from "../../../src/moon-query.js";

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
});
