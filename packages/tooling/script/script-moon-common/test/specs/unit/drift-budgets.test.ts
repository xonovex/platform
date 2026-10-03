import {describe, expect, it} from "vitest";
import {
  countProseWords,
  evaluateBudgets,
  readBudgetManifest,
  resolveDriftLintMode,
  seedBudgets,
} from "../../../src/drift-budgets.js";
import {memoryFileSystem} from "../../../src/file-system-memory.js";

const file = (path: string, text: string, kind = "reference" as const) => ({
  kind,
  path,
  text,
});

const words = (count: number) =>
  Array.from({length: count}, () => "word").join(" ");

describe("drift budgets", () => {
  it("counts prose words and excludes fenced code", () => {
    expect(countProseWords("one two three")).toBe(3);
    expect(
      countProseWords("one two\n\n```ts\nconst ignored = 1;\n```\n\nthree"),
    ).toBe(3);
  });

  it("passes a file within its recorded budget", () => {
    const findings = evaluateBudgets([file("a.md", "one two three")], {
      "a.md": 3,
    });

    expect(findings).toEqual([]);
  });

  it("flags a file over its recorded budget", () => {
    const findings = evaluateBudgets([file("a.md", "one two three four")], {
      "a.md": 3,
    });

    expect(findings).toHaveLength(1);
    expect(findings[0]?.message).toContain("over its 3-word budget");
  });

  it("applies the kind cap to a file absent from the manifest", () => {
    const long = Array.from({length: 700}, () => "word").join(" ");

    const findings = evaluateBudgets([file("new.md", long)], {});

    expect(findings).toHaveLength(1);
    expect(findings[0]?.message).toContain("650-word cap for a new reference");
  });

  it("caps each kind separately", () => {
    const findings = evaluateBudgets(
      [
        {kind: "command", path: "c.md", text: words(260)},
        {kind: "skill", path: "s.md", text: words(260)},
      ],
      {},
    );

    expect(findings.map(({path}) => path)).toEqual(["c.md"]);
  });

  it("seeds sorted budgets from current sizes and skips empty files", () => {
    const manifest = seedBudgets([
      file("b.md", "one two"),
      file("a.md", "one"),
      file("empty.md", " ".repeat(3)),
    ]);

    expect(Object.keys(manifest)).toEqual(["a.md", "b.md"]);
    expect(manifest).toEqual({"a.md": 1, "b.md": 2});
  });

  it("orders seeded keys by code point rather than by locale", () => {
    // A locale-aware sort ignores case and puts references/ first, which
    // reorders the whole manifest on every reseed.
    const manifest = seedBudgets([
      file("guide/references/a.md", "one"),
      file("guide/SKILL.md", "one"),
    ]);

    expect(Object.keys(manifest)).toEqual([
      "guide/SKILL.md",
      "guide/references/a.md",
    ]);
  });

  it("round-trips seeded budgets without findings", () => {
    const files = [file("a.md", "one two three"), file("b.md", "four five")];

    expect(evaluateBudgets(files, seedBudgets(files))).toEqual([]);
  });

  it("reads a manifest and reports malformed content", () => {
    const fs = memoryFileSystem({
      files: {
        "/repo/valid.json": JSON.stringify({"a.md": 12}),
        "/repo/malformed.json": "{",
        "/repo/negative.json": JSON.stringify({"a.md": -1}),
      },
    });

    expect(readBudgetManifest("/repo/valid.json", fs).manifest).toEqual({
      "a.md": 12,
    });
    expect(readBudgetManifest("/repo/missing.json", fs).manifest).toEqual({});
    expect(readBudgetManifest("/repo/malformed.json", fs).error).toContain(
      "not valid JSON",
    );
    expect(readBudgetManifest("/repo/negative.json", fs).error).toContain(
      "positive word budget",
    );
  });

  it("defaults to warn mode and opts into enforce explicitly", () => {
    expect(resolveDriftLintMode(undefined)).toBe("warn");
    expect(resolveDriftLintMode("")).toBe("warn");
    expect(resolveDriftLintMode("anything")).toBe("warn");
    expect(resolveDriftLintMode(" enforce ")).toBe("enforce");
  });
});
