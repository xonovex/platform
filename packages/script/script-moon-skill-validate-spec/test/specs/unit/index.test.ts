import {cpSync, mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main as validateSkill} from "../../../src/validate-skill.js";

describe("skill validator entrypoints", () => {
  const temporaryDirectories: string[] = [];

  afterEach(() => {
    vi.restoreAllMocks();
    for (const directory of temporaryDirectories) {
      rmSync(directory, {recursive: true, force: true});
    }
    temporaryDirectories.length = 0;
  });

  it("renders validator help without reading a skill", () => {
    expect(validateSkill(["--help"])).toBe(0);
  });

  it("rejects unknown validator options", () => {
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    expect(validateSkill(["--unknown"])).toBe(2);
  });

  it("validates a complete repository skill", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../../../skill/skill-code-quality",
    );
    vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill(["--strict", skill]);

    expect(result).toBe(0);
  });

  it("keeps drift findings advisory in warn mode and blocking in enforce mode", () => {
    // A copied skill in a throwaway workspace keeps the assertion independent of
    // whatever the real catalog currently measures.
    const root = mkdtempSync(join(tmpdir(), "skill-drift-"));
    temporaryDirectories.push(root);
    mkdirSync(join(root, ".moon"));
    const guide = join(root, "packages", "skill", "skill-code-quality");
    mkdirSync(join(root, "packages", "skill"), {recursive: true});
    cpSync(
      resolve(import.meta.dirname, "../../../../../skill/skill-code-quality"),
      guide,
      {recursive: true},
    );
    writeFileSync(
      join(root, "budgets.json"),
      JSON.stringify({
        "packages/skill/skill-code-quality/code-quality-guide/SKILL.md": 1,
      }),
      "utf8",
    );
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    expect(validateSkill(["--strict", guide])).toBe(0);
    expect(log.mock.calls.flat().join("\n")).toContain("[DRIFT] budget:");

    process.env.XONOVEX_LINT_MODE = "enforce";
    try {
      expect(validateSkill(["--strict", guide])).toBe(1);
    } finally {
      delete process.env.XONOVEX_LINT_MODE;
    }
  });

  it("allows the catalog's explicit Claude Code adapter name", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../../../skill/skill-claude-code",
    );
    vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill(["--strict", skill]);

    expect(result).toBe(0);
  });

  it("reports malformed frontmatter and missing catalog files", () => {
    const root = mkdtempSync(join(tmpdir(), "skill-validate-"));
    temporaryDirectories.push(root);
    const skill = join(root, "invalid-skill");
    mkdirSync(skill);
    mkdirSync(join(skill, "references", "nested"), {recursive: true});
    writeFileSync(
      join(skill, "references", "nested", "sources.md"),
      "# Details\n\n## External References\n\n- https://example.com/nested-source\n",
    );
    writeFileSync(
      join(skill, "SKILL.md"),
      `---
name: Invalid Skill
description: Use when checking a deliberately invalid skill. Triggers on validator test fixtures, even when the user doesn't say validation.
unknown-field: true
---
# Guide

Use references/missing_file.md with @references/legacy.md.

## External References

- https://example.com/misplaced-source
`,
    );
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill([skill]);

    expect(result).toBe(1);
    const output = log.mock.calls.flat().join("\n");
    expect(output).toContain("frontmatter: name 'Invalid Skill'");
    expect(output).toContain(
      "frontmatter: 'description' must use one double-quoted scalar",
    );
    expect(output).toContain("catalog: missing evals.json");
    expect(output).toContain(
      "sources: '## External References' belongs in SOURCES.md, not SKILL.md, references/nested/sources.md",
    );
    expect(output).toContain("Result: FAIL");
  });

  it("fails on an em dash or ellipsis anywhere in the package", () => {
    const root = mkdtempSync(join(tmpdir(), "skill-punctuation-"));
    temporaryDirectories.push(root);
    const packageDir = join(root, "skill-dashed");
    const skill = join(packageDir, "dashed-guide");
    mkdirSync(join(skill, "references"), {recursive: true});
    writeFileSync(join(packageDir, "package.json"), '{"name": "skill-dashed"}');
    writeFileSync(
      join(skill, "SKILL.md"),
      `---
name: dashed-guide
description: "Use when checking punctuation. Triggers on validator fixtures, even when the user doesn't say punctuation."
---
# Guide

- **Label** \u2014 detail
`,
    );
    writeFileSync(
      join(skill, "references", "one.md"),
      "# One\n\n- Use when\u2026\n",
    );
    writeFileSync(
      join(skill, "eval-queries.json"),
      String.raw`{"queries": [{"query": "a \u2014 b"}]}`,
    );
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill([skill]);

    expect(result).toBe(1);
    const output = log.mock.calls.flat().join("\n");
    expect(output).toContain(
      "punctuation: em dash in dashed-guide/SKILL.md line 7, " +
        "use a comma, colon, or full stop:",
    );
    expect(output).toContain(
      "punctuation: ellipsis in dashed-guide/references/one.md line 3, " +
        "use three periods:",
    );
    expect(output).toContain(
      "punctuation: escaped em dash in dashed-guide/eval-queries.json line 1",
    );
  });

  it("returns an input error for invalid YAML frontmatter", () => {
    const root = mkdtempSync(join(tmpdir(), "skill-validate-"));
    temporaryDirectories.push(root);
    const skill = join(root, "invalid-yaml");
    mkdirSync(skill);
    writeFileSync(join(skill, "SKILL.md"), "---\nname: [\n---\n# Guide\n");
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);

    const result = validateSkill([skill]);

    expect(result).toBe(2);
  });
});
