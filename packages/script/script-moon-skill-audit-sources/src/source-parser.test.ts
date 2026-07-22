import {describe, expect, it} from "vitest";
import {hasReferenceMapping, parseSources} from "./source-parser.js";

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

  it("parses explicit reference coverage and all-reference coverage", () => {
    const sources = parseSources(`# Sources

## Selected references
- **URL:** https://example.com/selected
- **References:** references/one.md, \`two.md\`
- **Last reviewed:** 2026-07-19

## Whole guide
- **Provenance:** Repository-original guidance
- **References:** all
- **Last reviewed:** 2026-07-19
`);

    expect([...(sources[0]?.refs ?? [])]).toEqual(["one.md", "two.md"]);
    expect(sources[0]?.coversAllReferences).toBe(false);
    expect(sources[1]?.coversAllReferences).toBe(true);
  });

  it("recognizes legacy all-reference and bare arrow mappings", () => {
    const sources = parseSources(`# Sources

## Legacy wildcard
- **URL:** https://example.com/all
- **Used for:** SKILL.md and all \`references/\`
- **Last reviewed:** 2026-07-19

## Bare arrow
- **Provenance:** Repository-local implementation
- **Aspects extracted:** architecture split → \`architecture.md\`
- **Last reviewed:** 2026-07-19
`);

    expect(sources[0]?.coversAllReferences).toBe(true);
    expect([...(sources[1]?.refs ?? [])]).toEqual(["architecture.md"]);
  });

  it("distinguishes mapped source blocks from unmapped provenance", () => {
    const sources = parseSources(`# Sources

## Mapped
- **URL:** https://example.com/mapped
- **References:** references/one.md
- **Last reviewed:** 2026-07-19

## Unmapped
- **Provenance:** Repository-original guidance
- **Last reviewed:** 2026-07-19
`);

    const mapped = sources[0];
    const unmapped = sources[1];
    if (mapped === undefined || unmapped === undefined) {
      throw new Error("fixture source blocks were not parsed");
    }
    expect(hasReferenceMapping(mapped)).toBe(true);
    expect(hasReferenceMapping(unmapped)).toBe(false);
  });

  it("keeps human-readable version baselines", () => {
    const [source] = parseSources(`# Sources

## Versioned CLI
- **URL:** https://example.com/docs
- **Version:** \`acli 1.3.22-stable\`
- **Last reviewed:** 2026-07-19
`);

    expect(source?.version).toBe("acli 1.3.22-stable");
  });
});
