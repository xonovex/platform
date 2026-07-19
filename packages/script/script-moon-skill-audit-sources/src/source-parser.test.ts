import {describe, expect, it} from "vitest";
import {parseSources} from "./source-parser.js";

describe("parseSources", () => {
  it("parses singular, plural-list, and plural-inline URLs", () => {
    const sources = parseSources(`# Sources

## Singular
- **URL:** https://example.com/one
- **Last reviewed:** 2026-07-19

## List
- **URLs:**
  - https://example.com/two
  - https://example.com/three
- **Last reviewed:** 2026-07-19

## Inline
- **URLs:** https://example.com/four · https://example.com/five
- **Last reviewed:** 2026-07-19
`);

    expect(sources.map((source) => source.urls)).toEqual([
      ["https://example.com/one"],
      ["https://example.com/two", "https://example.com/three"],
      ["https://example.com/four", "https://example.com/five"],
    ]);
  });

  it("retains repository-original provenance without inventing a URL", () => {
    const sources = parseSources(`# Sources

## Repository synthesis
- **Provenance:** Repository-original conventions distilled from project practice
- **Last reviewed:** 2026-07-19
`);

    expect(sources).toHaveLength(1);
    expect(sources[0]?.url).toBeUndefined();
    expect(sources[0]?.provenance).toContain("Repository-original");
  });
});
