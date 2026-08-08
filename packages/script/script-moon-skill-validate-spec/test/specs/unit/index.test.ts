import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main as validateSkill} from "../../../src/validate-skill.js";

const ROOT = "/repo";

describe("skill validator entrypoints", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders validator help without reading a skill", () => {
    expect(validateSkill(["--help"])).toBe(0);
  });

  it("rejects unknown validator options", () => {
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);
    expect(validateSkill(["--unknown"])).toBe(2);
  });

  it("reports malformed frontmatter and missing catalog files", () => {
    const fs = memoryFileSystem({directories: [ROOT]});
    const skill = join(ROOT, "invalid-skill");
    fs.makeDirectory(join(skill, "references", "nested"));
    fs.writeFile(
      join(skill, "references", "nested", "sources.md"),
      "# Details\n\n## External References\n\n- https://example.com/nested-source\n",
    );
    fs.writeFile(
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

    const result = validateSkill([skill], fs);

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
    const fs = memoryFileSystem({directories: [ROOT]});
    const packageDir = join(ROOT, "skill-dashed");
    const skill = join(packageDir, "dashed-guide");
    fs.makeDirectory(join(skill, "references"));
    fs.writeFile(join(packageDir, "package.json"), '{"name": "skill-dashed"}');
    fs.writeFile(
      join(skill, "SKILL.md"),
      `---
name: dashed-guide
description: "Use when checking punctuation. Triggers on validator fixtures, even when the user doesn't say punctuation."
---
# Guide

- **Label** \u2014 detail
`,
    );
    fs.writeFile(
      join(skill, "references", "one.md"),
      "# One\n\n- Use when\u2026\n",
    );
    fs.writeFile(
      join(skill, "eval-queries.json"),
      String.raw`{"queries": [{"query": "a \u2014 b"}]}`,
    );
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill([skill], fs);

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
    const fs = memoryFileSystem({directories: [ROOT]});
    const skill = join(ROOT, "invalid-yaml");
    fs.writeFile(join(skill, "SKILL.md"), "---\nname: [\n---\n# Guide\n");
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);

    const result = validateSkill([skill], fs);

    expect(result).toBe(2);
  });
});
