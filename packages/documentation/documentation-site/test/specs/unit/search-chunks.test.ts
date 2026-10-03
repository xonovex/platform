import {describe, expect, it} from "vitest";
import {
  searchChunkName,
  searchChunks,
  splitSearchIndex,
} from "../../../src/search-chunks.js";

describe("large local search indexes", () => {
  it("keeps the complete index across separate bounded modules", () => {
    const index = JSON.stringify({
      text: String.raw`Examples with "quotes", \ paths, and {{templates}}. `.repeat(
        20_000,
      ),
    });
    const result = splitSearchIndex(
      "export default " + JSON.stringify(index),
      "/@localSearchIndexroot",
    );
    expect(result).toBeDefined();
    const parts = [...(result?.modules.values() ?? [])].map((code) => {
      const value: unknown = JSON.parse(code.slice("export default ".length));
      return value;
    });
    expect(parts.join("")).toBe(index);
    expect(
      [...(result?.modules.values() ?? [])].every(
        (part) => part.length < 500_000,
      ),
    ).toBe(true);
  });

  it("leaves unrelated modules and small indexes unchanged", () => {
    expect(
      splitSearchIndex("export default '{}'", "/@localSearchIndexroot"),
    ).toBeUndefined();
    expect(
      splitSearchIndex(
        "export default " + JSON.stringify("x".repeat(400_000)),
        "/unrelated",
      ),
    ).toBeUndefined();
    expect(searchChunkName("/unrelated")).toBeUndefined();
  });

  it("resolves generated parts for the production bundle", () => {
    const plugin = searchChunks();
    const code = plugin.transform(
      "export default " + JSON.stringify("x".repeat(400_000)),
      "/@localSearchIndexroot",
    );
    expect(code).toContain(".join('')");
    expect(
      plugin.resolveId("xonovex-search-part:@localSearchIndexroot:0"),
    ).toBe("xonovex-search-part:@localSearchIndexroot:0");
    expect(
      plugin.load("xonovex-search-part:@localSearchIndexroot:0"),
    ).toContain("export default");
  });
});
