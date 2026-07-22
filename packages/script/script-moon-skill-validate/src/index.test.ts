import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main as validateLinks} from "./validate-links.js";
import {main as validateSkill} from "./validate-skill.js";

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

  it("renders link-validator help without scanning the repository", () => {
    expect(validateLinks(["--help"])).toBe(0);
  });

  it("validates a complete repository skill", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../skill/skill-code-quality",
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
    writeFileSync(
      join(skill, "SKILL.md"),
      `---
name: Invalid Skill
description: Use when checking a deliberately invalid skill. Triggers on validator test fixtures, even when the user doesn't say validation.
unknown-field: true
---
# Guide

Use references/missing_file.md with @references/legacy.md.
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
    expect(output).toContain("Result: FAIL");
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
