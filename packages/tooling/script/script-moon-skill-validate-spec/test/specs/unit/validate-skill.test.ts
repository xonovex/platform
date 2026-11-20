import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {DriftLintMode} from "@xonovex/script-moon-common/drift-budgets";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/validate-skill.js";

const ROOT = "/repo";

// Runs the validator over fs and returns the exit code with what it wrote.
const run = (
  argv: readonly string[],
  fs: FileSystem = memoryFileSystem({directories: [ROOT]}),
  driftMode: DriftLintMode = "warn",
) => {
  const output = memoryOutput();
  const code = main(argv, ROOT, {fs, output, driftMode});
  return {
    code,
    stdout: output.stdoutLines().join("\n"),
    stderr: output.stderrLines().join("\n"),
  };
};

describe("skill validator", () => {
  it("renders validator help without reading a skill", () => {
    const result = run(["--help"]);

    expect(result.code).toBe(0);
    expect(result.stdout).toContain("usage: moon-skill-validate-spec");
  });

  it("rejects unknown validator options", () => {
    expect(run(["--unknown"])).toMatchObject({
      code: 2,
      stderr: "Error: unrecognized argument: --unknown",
    });
  });

  it("rejects a target that does not exist", () => {
    expect(run(["missing"])).toMatchObject({
      code: 2,
      stderr: "Error: target not found: /repo/missing",
    });
  });

  it("rejects a directory that holds no SKILL.md", () => {
    const fs = memoryFileSystem({directories: [join(ROOT, "empty")]});

    expect(run(["empty"], fs)).toMatchObject({
      code: 2,
      stderr: "Error: SKILL.md not found at /repo/empty/SKILL.md",
    });
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
    const {code, stdout: output} = run([skill], fs);

    expect(code).toBe(1);
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

  it("rejects a kebab-case name that does not start with a gerund", () => {
    const fs = memoryFileSystem({directories: [ROOT]});
    const skill = join(ROOT, "git-skill");
    fs.writeFile(join(skill, "SKILL.md"), "---\nname: git-skill\n---\n# Git\n");
    const result = run([skill], fs);

    expect(result.code).toBe(1);
    expect(result.stdout).toContain(
      "frontmatter: name 'git-skill' is not a gerund and then its object (regex ^[a-z0-9]+ing(-[a-z0-9]+)+$)",
    );
  });

  it("fails on an em dash or ellipsis anywhere in the package", () => {
    const fs = memoryFileSystem({directories: [ROOT]});
    const packageDir = join(ROOT, "skill-dashed");
    const skill = join(packageDir, "using-dashed");
    fs.makeDirectory(join(skill, "references"));
    fs.writeFile(join(packageDir, "package.json"), '{"name": "skill-dashed"}');
    fs.writeFile(
      join(skill, "SKILL.md"),
      `---
name: using-dashed
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
    const {code, stdout: output} = run([skill], fs);

    expect(code).toBe(1);
    expect(output).toContain(
      "punctuation: em dash in using-dashed/SKILL.md line 7, " +
        "use a comma, colon, or full stop:",
    );
    expect(output).toContain(
      "punctuation: ellipsis in using-dashed/references/one.md line 3, " +
        "use three periods:",
    );
    expect(output).toContain(
      "punctuation: escaped em dash in using-dashed/eval-queries.json line 1",
    );
  });

  it.each([
    [820, "description: 820 chars (within 820)"],
    [821, "description: 821 chars (over the 820-character catalog limit)"],
  ])("checks a %i-character description against the limit", (length, line) => {
    const fs = memoryFileSystem({directories: [ROOT]});
    const skill = join(ROOT, "using-limits");
    const lead =
      "Use when checking the description limit. Triggers on validator fixtures, even when the user doesn't say limit.";
    fs.writeFile(
      join(skill, "SKILL.md"),
      `---\nname: using-limits\ndescription: "${lead.padEnd(length, "x")}"\n---\n# Guide\n`,
    );
    expect(run([skill], fs).stdout).toContain(line);
  });

  it("returns an input error for invalid YAML frontmatter", () => {
    const fs = memoryFileSystem({directories: [ROOT]});
    const skill = join(ROOT, "invalid-yaml");
    fs.writeFile(join(skill, "SKILL.md"), "---\nname: [\n---\n# Guide\n");
    const result = run([skill], fs);

    expect(result.code).toBe(2);
    expect(result.stderr).toContain("Error: Invalid YAML frontmatter:");
  });
});
